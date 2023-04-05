package webcrawler

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"os"
	"strings"
	"sync"

	"github.com/deanrtaylor1/gosearch/src/bm25"
	"github.com/deanrtaylor1/gosearch/src/lexer"
	"github.com/deanrtaylor1/gosearch/src/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func closeFile(f *os.File, errChan chan<- error) {
	err := f.Close()
	if err != nil {
		errChan <- fmt.Errorf("error closing file: %w", err)
	}
}

// Add a helper function to extract the domain name from a URL
func extractDomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}

func crawlPage(urlToCrawl string, foundUrls chan<- string, dirName string, errChan chan<- error) {
	// Add your web crawling logic here
	// When you find a new URL, send it to the channel: foundUrls <- newURL

	fmt.Println("initiating get request", urlToCrawl)
	resp, err := http.Get(urlToCrawl)

	if err != nil {
		errChan <- fmt.Errorf("error accessing site file: %w", err)
		return
	}

	defer resp.Body.Close()
	fmt.Println("accessing http body", urlToCrawl)
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		errChan <- fmt.Errorf("error reading file: %w", err)
		return
	}
	fullUrl, err := url.Parse(urlToCrawl)
	if err != nil {
		fmt.Println(err)
	}

	filename := fullUrl.Path

	filename = strings.ReplaceAll(filename, "/", "_")

	f, err := os.Create(dirName + "/" + filename + ".html")
	defer func() {
		closeFile(f, errChan)
	}()

	if err != nil {
		errChan <- fmt.Errorf("error creating file: %w", err)

	}
	l, err := f.Write(body)

	if err != nil {
		errChan <- fmt.Errorf("error writing file: %w", err)

	}
	fmt.Println(l, "bytes written successfully")

	fmt.Println("reading file", urlToCrawl)
	file, err := os.ReadFile(dirName + "/" + filename + ".html")
	if err != nil {
		errChan <- fmt.Errorf("error reading file: %w", err)
	}

	// extract the links from the file
	links := lexer.ParseLinks(string(file))
	// fmt.Println("parsing links", links)
	for _, link := range links {
		fmt.Println(link)
		// check if the link is a relative link
		parsedLink, err := url.Parse(link)
		if err != nil {
			errChan <- fmt.Errorf("error parsing link: %w", err)
			continue
		}

		if !parsedLink.IsAbs() {
			fmt.Println("link is relative")
			// Resolve the relative link against the base URL
			resolvedLink := fullUrl.ResolveReference(parsedLink)
			link = resolvedLink.String()
			fmt.Println("new link", link)
		}

		foundUrls <- link

	}

}

func CrawlDomain(domain string) {
	fmt.Println("crawling domain: ", domain)

	visited := make(map[string]bool)
	urlFiles := make(map[string]string)

	visitedMutex := sync.Mutex{}
	urlsMutex := sync.Mutex{}

	fullUrl, err := url.Parse(domain)
	if err != nil {
		fmt.Println(err)
	}
	dirName := fullUrl.Host
	fmt.Println("creating dir", dirName)

	err = os.MkdirAll(dirName, os.ModePerm)

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	// Use a buffered channel to store found URLs
	foundUrls := make(chan string, 10)
	errChan := make(chan error, 10)
	// Use a WaitGroup to track the number of active goroutines
	var wg sync.WaitGroup

	// Start with the initial URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		crawlPage(domain, foundUrls, dirName, errChan)
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case newURL := <-foundUrls:
			fmt.Println("Received new URL: ", newURL, "")
			visitedMutex.Lock()

			// If the URL has already been visited, skip it
			if visited[newURL] {
				fmt.Println("URL already visited: ", newURL)
				visitedMutex.Unlock()
				continue
			}

			// Mark the URL as visited
			visited[newURL] = true
			visitedMutex.Unlock()

			// Check if the new URL has the same domain
			if extractDomain(newURL) != extractDomain(domain) {
				fmt.Println("URL is not in the same domain: ", newURL)
				continue
			}

			fmt.Println("URL is new, adding to the queue: ", newURL)
			urlPath, err := url.Parse(newURL)
			if err != nil {
				fmt.Println(err)
			}
			fileName := strings.ReplaceAll(urlPath.Path, "/", "_")
			urlsMutex.Lock()
			urlFiles[fileName] = newURL
			urlsMutex.Unlock()
			wg.Add(1)
			go func(urlToCrawl string) {
				defer wg.Done()
				crawlPage(urlToCrawl, foundUrls, dirName, errChan)
			}(newURL)

		case err := <-errChan:
			fmt.Printf("Error: %v\n", err)

		case <-done:
			urlsMutex.Lock()
			util.MapToJSON(urlFiles, true, dirName+"/urls.json")
			urlsMutex.Unlock()
			return
		}

	}

}

func crawlPageV2(urlToCrawl string, foundUrls chan<- string, dirName string, errChan chan<- error, cachedDataMutex *sync.Mutex, cachedData *map[string]util.IndexedData) {
	// Add your web crawling logic here
	// When you find a new URL, send it to the channel: foundUrls <- newURL

	fmt.Println("initiating get request", urlToCrawl)
	resp, err := http.Get(urlToCrawl)

	if err != nil {
		errChan <- fmt.Errorf("error accessing site file: %w", err)
		return
	}

	defer resp.Body.Close()
	fmt.Println("accessing http body", urlToCrawl)
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		errChan <- fmt.Errorf("error reading html response body: %w", err)
		return
	}

	fullUrl, err := url.Parse(urlToCrawl)
	if err != nil {
		fmt.Println(err)
	}

	// filename := fullUrl.Path

	// filename = strings.ReplaceAll(filename, "/", "_")

	textContent := lexer.ParseHtmlTextContent(string(body))

	IndexedData := util.IndexedData{
		URL:     urlToCrawl,
		Content: textContent,
	}
	cachedDataMutex.Lock()
	(*cachedData)[urlToCrawl] = IndexedData
	cachedDataMutex.Unlock()
	// extract the links from the file
	links := lexer.ParseLinks(string(body))
	// fmt.Println("parsing links", links)
	for _, link := range links {
		fmt.Println(link)
		// check if the link is a relative link
		parsedLink, err := url.Parse(link)
		if err != nil {
			errChan <- fmt.Errorf("error parsing link: %w", err)
			continue
		}

		if !parsedLink.IsAbs() {
			fmt.Println("link is relative")
			// Resolve the relative link against the base URL
			resolvedLink := fullUrl.ResolveReference(parsedLink)
			link = resolvedLink.String()
			fmt.Println("new link", link)
		}

		foundUrls <- link

	}

}

const maxURLsToCrawl = 10000

func CrawlDomainV2(domain string) {
	fmt.Println("crawling domain: ", domain)

	cachedData := make(map[string]util.IndexedData)
	visited := make(map[string]bool)
	urlFiles := make(map[string]string)

	cachedDataMutex := sync.Mutex{}
	visitedMutex := sync.Mutex{}
	urlsMutex := sync.Mutex{}

	fullUrl, err := url.Parse(domain)
	if err != nil {
		fmt.Println(err)
	}
	dirName := fullUrl.Host
	fmt.Println("creating dir", dirName)

	err = os.MkdirAll(dirName, os.ModePerm)

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	// Use a buffered channel to store found URLs
	foundUrls := make(chan string, 10)
	errChan := make(chan error, 10)
	// Use a WaitGroup to track the number of active goroutines
	var wg sync.WaitGroup

	// Start with the initial URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		crawlPageV2(domain, foundUrls, dirName, errChan, &cachedDataMutex, &cachedData)
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

outerLoop:
	for {

		select {
		case newURL := <-foundUrls:
			fmt.Println("Received new URL: ", newURL, "")
			visitedMutex.Lock()
			numberOfVisitedURLs := len(visited)
			if numberOfVisitedURLs >= maxURLsToCrawl {
				fmt.Println("Reached max number of URLs to crawl: ", maxURLsToCrawl)
				visitedMutex.Unlock()
				urlsMutex.Lock()
				var compressedData bytes.Buffer
				gzipWriter := gzip.NewWriter(&compressedData)

				encoder := gob.NewEncoder(gzipWriter)
				if err := encoder.Encode(cachedData); err != nil {
					log.Fatalf("Error encoding indexed data: %v", err)
				}

				if err := gzipWriter.Close(); err != nil {
					log.Fatalf("Error closing gzip writer: %v", err)
				}
				filename := fmt.Sprintf("indexed_data_%s.gz", time.Now().Format("20060102_150405"))
				if err := os.WriteFile(dirName+"./"+filename, compressedData.Bytes(), 0644); err != nil {
					log.Fatalf("Error writing compressed data to disk: %v", err)
				}
				cachedDataMutex.Unlock()
				break outerLoop
			}
			// If the URL has already been visited, skip it
			if visited[newURL] {
				fmt.Println("URL already visited: ", newURL)
				visitedMutex.Unlock()
				continue
			}

			// Mark the URL as visited
			visited[newURL] = true
			visitedMutex.Unlock()

			// Check if the new URL has the same domain
			if extractDomain(newURL) != extractDomain(domain) {
				fmt.Println("URL is not in the same domain: ", newURL)
				continue
			}

			fmt.Println("URL is new, adding to the queue: ", newURL)
			urlPath, err := url.Parse(newURL)
			if err != nil {
				fmt.Println(err)
			}
			fileName := urlToName(urlPath.Path)
			fmt.Println("Filename: ", fileName)
			urlsMutex.Lock()
			urlFiles[newURL] = fileName
			urlsMutex.Unlock()
			wg.Add(1)
			go func(urlToCrawl string) {
				defer wg.Done()
				crawlPageV2(urlToCrawl, foundUrls, dirName, errChan, &cachedDataMutex, &cachedData)
			}(newURL)

		case err := <-errChan:
			fmt.Printf("Error: %v\n", err)

		case <-done:
			cachedDataMutex.Lock()
			var compressedData bytes.Buffer
			gzipWriter := gzip.NewWriter(&compressedData)

			encoder := gob.NewEncoder(gzipWriter)
			if err := encoder.Encode(cachedData); err != nil {
				log.Fatalf("Error encoding indexed data: %v", err)
			}

			if err := gzipWriter.Close(); err != nil {
				log.Fatalf("Error closing gzip writer: %v", err)
			}
			filename := "indexed-data.gz"
			if err := os.WriteFile(dirName+"./"+filename, compressedData.Bytes(), 0644); err != nil {
				log.Fatalf("Error writing compressed data to disk: %v", err)
			}
			cachedDataMutex.Unlock()
			urlsMutex.Lock()
			var compressedData2 bytes.Buffer
			gzipWriter2 := gzip.NewWriter(&compressedData2)

			encoder2 := gob.NewEncoder(gzipWriter2)
			if err := encoder2.Encode(urlFiles); err != nil {
				log.Fatalf("Error encoding indexed data: %v", err)
			}

			if err := gzipWriter2.Close(); err != nil {
				log.Fatalf("Error closing gzip writer: %v", err)
			}
			filename2 := "url-files.gz"
			if err := os.WriteFile(dirName+"./"+filename2, compressedData2.Bytes(), 0644); err != nil {
				log.Fatalf("Error writing compressed data to disk: %v", err)
			}
			urlsMutex.Unlock()
			return
		}

	}

}

func crawlPageUpdateModel(urlToCrawl string, foundUrls chan<- string, dirName string, errChan chan<- error, cachedDataMutex *sync.Mutex, cachedData *map[string]util.IndexedData, model *bm25.Model) {
	// Add your web crawling logic here
	// When you find a new URL, send it to the channel: foundUrls <- newURL

	fmt.Println("initiating get request", urlToCrawl)
	resp, err := http.Get(urlToCrawl)

	if err != nil {
		errChan <- fmt.Errorf("error accessing site file: %w", err)
		return
	}

	defer resp.Body.Close()
	fmt.Println("accessing http body", urlToCrawl)
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		errChan <- fmt.Errorf("error reading html response body: %w", err)
		return
	}

	fullUrl, err := url.Parse(urlToCrawl)
	if err != nil {
		fmt.Println(err)
	}

	// filename := fullUrl.Path

	// filename = strings.ReplaceAll(filename, "/", "_")

	textContent := lexer.ParseHtmlTextContent(string(body))

	IndexedData := util.IndexedData{
		URL:     urlToCrawl,
		Content: textContent,
	}
	fmt.Println("Locking cached data")
	cachedDataMutex.Lock()
	(*cachedData)[urlToCrawl] = IndexedData
	cachedDataMutex.Unlock()
	fmt.Println("Unlocking cached data")
	fmt.Println("Locking model")
	model.ModelLock.Lock()
	model.DirLength += 1
	model.ModelLock.Unlock()
	fmt.Println("Unlocking model")
	fmt.Println("Locking model")
	model.ModelLock.Lock()
	model.DocCount += 1
	model.ModelLock.Unlock()
	fmt.Println("Unlocking model")
	//fmt.Println(filePath)
	content := IndexedData.Content
	//fmt.Println(filePath, content)

	fileSize := len(content)

	fmt.Println(IndexedData.URL, " => ", fileSize)
	tf := make(bm25.TermFreq)

	tokenLexer := lexer.NewLexer(content)
	for {
		token, err := tokenLexer.Next()
		if err != nil {
			fmt.Println("EOF")
			break
		}

		tf[token] += 1
		//stats := mapToSortedSlice(tf)
		//fmt.Println(filePath, " => ", token, " => ", tf[token])
	}
	fmt.Println("Locking model")
	model.ModelLock.Lock()
	for token := range tf {
		model.TermCount += 1
		model.DF[token] += 1
	}
	model.TFPD[IndexedData.URL] = bm25.ConvertToDocData(tf)
	model.ModelLock.Unlock()
	fmt.Println("Unlocking model")

	// extract the links from the file
	links := lexer.ParseLinks(string(body))
	// fmt.Println("parsing links", links)
	for _, link := range links {
		fmt.Println(link)
		if shouldIgnoreLink(link) {
			continue
		}
		// check if the link is a relative link
		parsedLink, err := url.Parse(link)
		if err != nil {
			errChan <- fmt.Errorf("error parsing link: %w", err)
			continue
		}

		if !parsedLink.IsAbs() {
			fmt.Println("link is relative")
			// Resolve the relative link against the base URL
			resolvedLink := fullUrl.ResolveReference(parsedLink)
			link = resolvedLink.String()
			fmt.Println("new link", link)
		}

		foundUrls <- link

	}

}

func CrawlDomainUpdateModel(domain string, model *bm25.Model) {
	fmt.Println("crawling domain: ", domain)

	cachedData := make(map[string]util.IndexedData)
	visited := make(map[string]bool)
	urlFiles := make(map[string]string)

	cachedDataMutex := sync.Mutex{}
	visitedMutex := sync.Mutex{}
	urlsMutex := sync.Mutex{}

	fullUrl, err := url.Parse(domain)
	if err != nil {
		fmt.Println(err)
	}
	dirName := fullUrl.Host
	fmt.Println("creating dir", dirName)

	err = os.MkdirAll(dirName, os.ModePerm)
	model.ModelLock.Lock()
	model.Name = dirName
	model.ModelLock.Unlock()

	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	// Use a buffered channel to store found URLs
	foundUrls := make(chan string, 10)
	errChan := make(chan error, 10)
	// Use a WaitGroup to track the number of active goroutines
	var wg sync.WaitGroup

	// Start with the initial URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		crawlPageUpdateModel(domain, foundUrls, dirName, errChan, &cachedDataMutex, &cachedData, model)
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

outerLoop:
	for {

		select {
		case newURL := <-foundUrls:
			fmt.Println("Received new URL: ", newURL, "")
			visitedMutex.Lock()
			numberOfVisitedURLs := len(visited)
			if numberOfVisitedURLs >= maxURLsToCrawl {
				fmt.Println("Reached max number of URLs to crawl: ", maxURLsToCrawl)
				visitedMutex.Unlock()
				cachedDataMutex.Lock()
				var compressedData bytes.Buffer
				gzipWriter := gzip.NewWriter(&compressedData)

				encoder := gob.NewEncoder(gzipWriter)
				if err := encoder.Encode(cachedData); err != nil {
					log.Fatalf("Error encoding indexed data: %v", err)
				}

				if err := gzipWriter.Close(); err != nil {
					log.Fatalf("Error closing gzip writer: %v", err)
				}
				filename := "indexed-data.gz"
				if err := os.WriteFile(dirName+"./"+filename, compressedData.Bytes(), 0644); err != nil {
					log.Fatalf("Error writing compressed data to disk: %v", err)
				}
				cachedDataMutex.Unlock()
				urlsMutex.Lock()
				var compressedData2 bytes.Buffer
				gzipWriter2 := gzip.NewWriter(&compressedData2)

				encoder2 := gob.NewEncoder(gzipWriter2)
				if err := encoder2.Encode(urlFiles); err != nil {
					log.Fatalf("Error encoding indexed data: %v", err)
				}

				if err := gzipWriter2.Close(); err != nil {
					log.Fatalf("Error closing gzip writer: %v", err)
				}
				filename2 := "url-files.gz"
				if err := os.WriteFile(dirName+"./"+filename2, compressedData2.Bytes(), 0644); err != nil {
					log.Fatalf("Error writing compressed data to disk: %v", err)
				}
				urlsMutex.Unlock()
				fmt.Println("\033[31m------------------------------------")
				fmt.Println("\033[31mFINISHED CRAWLING LIMIT REACHED")
				fmt.Println("\033[31m------------------------------------\033[0m")
				break outerLoop
			}
			// If the URL has already been visited, skip it
			if visited[newURL] {
				fmt.Println("URL already visited: ", newURL)
				visitedMutex.Unlock()
				continue
			}

			// Mark the URL as visited
			visited[newURL] = true
			visitedMutex.Unlock()

			// Check if the new URL has the same domain
			if extractDomain(newURL) != extractDomain(domain) {
				fmt.Println("URL is not in the same domain: ", newURL)
				continue
			}

			fmt.Println("URL is new, adding to the queue: ", newURL)
			urlPath, err := url.Parse(newURL)
			if err != nil {
				fmt.Println(err)
			}
			fileName := urlToName(urlPath.Path)
			fmt.Println("Filename: ", fileName)
			urlsMutex.Lock()
			urlFiles[newURL] = fileName
			urlsMutex.Unlock()
			model.ModelLock.Lock()
			model.UrlFiles[newURL] = fileName
			model.ModelLock.Unlock()
			wg.Add(1)
			go func(urlToCrawl string) {
				visitedMutex.Lock()
				numberOfVisitedURLs := len(visited)
				fmt.Println("Number of visited URLs: ", numberOfVisitedURLs)
				visitedMutex.Unlock()
				defer wg.Done()
				crawlPageUpdateModel(urlToCrawl, foundUrls, dirName, errChan, &cachedDataMutex, &cachedData, model)
			}(newURL)

		case err := <-errChan:
			fmt.Printf("Error: %v\n", err)

		case <-done:
			model.ModelLock.Lock()
			model.IsComplete = true
			model.ModelLock.Unlock()
			cachedDataMutex.Lock()
			var compressedData bytes.Buffer
			gzipWriter := gzip.NewWriter(&compressedData)

			encoder := gob.NewEncoder(gzipWriter)
			if err := encoder.Encode(cachedData); err != nil {
				log.Fatalf("Error encoding indexed data: %v", err)
			}

			if err := gzipWriter.Close(); err != nil {
				log.Fatalf("Error closing gzip writer: %v", err)
			}
			filename := "indexed-data.gz"
			if err := os.WriteFile(dirName+"./"+filename, compressedData.Bytes(), 0644); err != nil {
				log.Fatalf("Error writing compressed data to disk: %v", err)
			}
			cachedDataMutex.Unlock()
			urlsMutex.Lock()
			var compressedData2 bytes.Buffer
			gzipWriter2 := gzip.NewWriter(&compressedData2)

			encoder2 := gob.NewEncoder(gzipWriter2)
			if err := encoder2.Encode(urlFiles); err != nil {
				log.Fatalf("Error encoding indexed data: %v", err)
			}

			if err := gzipWriter2.Close(); err != nil {
				log.Fatalf("Error closing gzip writer: %v", err)
			}
			filename2 := "url-files.gz"
			if err := os.WriteFile(dirName+"./"+filename2, compressedData2.Bytes(), 0644); err != nil {
				log.Fatalf("Error writing compressed data to disk: %v", err)
			}
			urlsMutex.Unlock()
			fmt.Println("\033[32m------------------------------------")
			fmt.Println("\033[32mFINISHED CRAWLING")
			fmt.Println("\033[32m------------------------------------\033[0m")
			return
		}
	}

}

func urlToName(urlPath string) string {
	// Remove common file extensions
	urlPath = strings.TrimSuffix(urlPath, ".html")
	urlPath = strings.TrimSuffix(urlPath, ".php")
	urlPath = strings.TrimSuffix(urlPath, ".asp")

	// Split the path into components
	components := strings.Split(urlPath, "/")
	// Create a Caser for title casing in English without lowercasing the entire string first
	caser := cases.Title(language.English, cases.NoLower)

	// Process each component
	for i, component := range components {
		// Replace hyphens and underscores with spaces
		component = strings.ReplaceAll(component, "-", " ")
		component = strings.ReplaceAll(component, "_", " ")

		// Convert to title case
		components[i] = caser.String(component)
	}

	// Join components with " > "
	return strings.Join(components, " > ")
}

func shouldIgnoreLink(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return true
	}

	return parsedURL.Fragment != ""
}

/*
 */
