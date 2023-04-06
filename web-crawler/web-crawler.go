package webcrawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"os"
	"strings"
	"sync"

	"github.com/deanrtaylor1/gosearch/bm25"
	"github.com/deanrtaylor1/gosearch/lexer"
	"github.com/deanrtaylor1/gosearch/logger"
	"github.com/deanrtaylor1/gosearch/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Add a helper function to extract the domain name from a URL
func extractDomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}

const maxURLsToCrawl = 10000

func crawlPageUpdateModel(urlToCrawl string, foundUrls chan<- string, dirName string, errChan chan<- error, cachedDataMutex *sync.Mutex, cachedData *map[string]util.IndexedData, model *bm25.Model) {
	// Add your web crawling logic here
	// When you find a new URL, send it to the channel: foundUrls <- newURL

	// log.Println("initiating get request to ", urlToCrawl)
	resp, err := http.Get(urlToCrawl)

	if err != nil {
		errChan <- fmt.Errorf("error accessing site file: %w", err)
		return
	}

	defer resp.Body.Close()
	//log.Println("accessing http body", urlToCrawl)
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		errChan <- fmt.Errorf("error reading html response body: %w", err)
		return
	}

	fullUrl, err := url.Parse(urlToCrawl)
	if err != nil {
		log.Println(err)
	}

	textContent := lexer.ParseHtmlTextContent(string(body))

	IndexedData := util.IndexedData{
		URL:     urlToCrawl,
		Content: textContent,
	}

	cachedDataMutex.Lock()
	(*cachedData)[urlToCrawl] = IndexedData
	cachedDataMutex.Unlock()

	model.ModelLock.Lock()
	model.DirLength += 1
	model.ModelLock.Unlock()

	model.ModelLock.Lock()
	model.DocCount += 1
	model.ModelLock.Unlock()

	content := IndexedData.Content

	// fileSize := len(content)

	// log.Println(IndexedData.URL, " => ", fileSize)
	tf := make(bm25.TermFreq)

	tokenLexer := lexer.NewLexer(content)
	for {
		token, err := tokenLexer.Next()
		if err != nil {
			//log.Println("EOF")
			break
		}

		tf[token] += 1
	}
	model.ModelLock.Lock()
	for token := range tf {
		model.TermCount += 1
		model.DF[token] += 1
	}
	model.TFPD[IndexedData.URL] = bm25.ConvertToDocData(tf)
	model.ModelLock.Unlock()

	// extract the links from the file
	links := lexer.ParseLinks(string(body))

	for _, link := range links {
		// log.Println(link)
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
			// log.Println("link is relative")
			// Resolve the relative link against the base URL
			resolvedLink := fullUrl.ResolveReference(parsedLink)
			link = resolvedLink.String()
			// log.Println("new link", link)
		}

		foundUrls <- link

	}

}

func CrawlDomainUpdateModel(domain string, model *bm25.Model) {
	log.Println("crawling domain: ", domain)
	start := time.Now()
	cachedData := make(map[string]util.IndexedData)
	visited := make(map[string]bool)
	urlFiles := make(map[string]string)
	reverseUrlFiles := make(map[string]string)

	cachedDataMutex := sync.Mutex{}
	visitedMutex := sync.Mutex{}
	urlsMutex := sync.Mutex{}
	reverseUrlsMutex := sync.Mutex{}

	fullUrl, err := url.Parse(domain)
	if err != nil {
		log.Println(err)
	}
	dirName := fmt.Sprint("indexes/" + fullUrl.Host)
	// log.Println("creating dir", dirName)

	err = os.MkdirAll(dirName, os.ModePerm)
	model.ModelLock.Lock()
	model.Name = fullUrl.Host
	model.ModelLock.Unlock()

	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	// Use a buffered channel to store found URLs
	foundUrls := make(chan string, 100)
	errChan := make(chan error, 100)
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
			// log.Println("Received new URL: ", newURL, "")
			visitedMutex.Lock()
			numberOfVisitedURLs := len(visited)
			if numberOfVisitedURLs >= maxURLsToCrawl {
				// log.Println("Reached max number of URLs to crawl: ", maxURLsToCrawl)
				visitedMutex.Unlock()
				model.ModelLock.Lock()
				model.IsComplete = true
				model.ModelLock.Unlock()

				cachedDataMutex.Lock()
				err := bm25.CompressAndWriteGzipFile("indexed-data.gz", cachedData, dirName)
				if err != nil {
					log.Fatal(err)
				}
				cachedDataMutex.Unlock()

				urlsMutex.Lock()
				err = bm25.CompressAndWriteGzipFile("url-files.gz", urlFiles, dirName)
				if err != nil {
					log.Fatal(err)
				}
				urlsMutex.Unlock()

				reverseUrlsMutex.Lock()
				err = bm25.CompressAndWriteGzipFile("reverse-url-files.gz", reverseUrlFiles, dirName)
				if err != nil {
					log.Fatal(err)
				}
				reverseUrlsMutex.Unlock()
				log.Println("\033[31m------------------------------------")
				log.Println("\033[31mFINISHED CRAWLING LIMIT REACHED")
				log.Println("\033[31m------------------------------------\033[0m")
				break outerLoop
			}
			// If the URL has already been visited, skip it
			if visited[newURL] {
				// log.Println("URL already visited: ", newURL)
				visitedMutex.Unlock()
				continue
			}

			// Mark the URL as visited
			visited[newURL] = true
			visitedMutex.Unlock()

			// Check if the new URL has the same domain
			if extractDomain(newURL) != extractDomain(domain) {
				// log.Println("URL is not in the same domain: ", newURL)
				continue
			}

			// log.Println("URL is new, adding to the queue: ", newURL)
			urlPath, err := url.Parse(newURL)
			if err != nil {
				log.Println(err)
			}
			fileName := urlToName(urlPath.Path)
			// log.Println("Filename: ", fileName)
			urlsMutex.Lock()
			urlFiles[newURL] = fileName
			reverseUrlFiles[fileName] = newURL
			urlsMutex.Unlock()

			reverseUrlsMutex.Lock()
			model.ReverseUrlFiles[fileName] = newURL
			reverseUrlsMutex.Unlock()

			model.ModelLock.Lock()
			model.UrlFiles[newURL] = fileName
			model.ModelLock.Unlock()

			wg.Add(1)
			go func(urlToCrawl string) {
				// visitedMutex.Lock()
				//numberOfVisitedURLs := len(visited)
				//log.Println("Number of visited URLs: ", numberOfVisitedURLs)
				// visitedMutex.Unlock()
				defer wg.Done()
				crawlPageUpdateModel(urlToCrawl, foundUrls, dirName, errChan, &cachedDataMutex, &cachedData, model)
			}(newURL)

		case err := <-errChan:
			logger.HandleError(err)

		case <-done:
			model.ModelLock.Lock()
			model.IsComplete = true
			model.ModelLock.Unlock()

			cachedDataMutex.Lock()
			err := bm25.CompressAndWriteGzipFile("indexed-data.gz", cachedData, dirName)
			if err != nil {
				log.Fatal(err)
			}
			cachedDataMutex.Unlock()

			urlsMutex.Lock()
			err = bm25.CompressAndWriteGzipFile("url-files.gz", urlFiles, dirName)
			if err != nil {
				log.Fatal(err)
			}
			urlsMutex.Unlock()

			reverseUrlsMutex.Lock()
			err = bm25.CompressAndWriteGzipFile("reverse-url-files.gz", reverseUrlFiles, dirName)
			if err != nil {
				log.Fatal(err)
			}
			reverseUrlsMutex.Unlock()
			elapsed := time.Since(start)
			log.Printf("\n\033[32m------------------------------------" + util.TerminalReset)
			fmt.Printf("\033[32mFINISHED CRAWLING %v in %dMs%v\n", fullUrl.Host, elapsed.Milliseconds(), util.TerminalReset)
			log.Printf("\033[32m------------------------------------\033[0m\n")
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

	// If the last component is empty, remove it
	if len(components) > 0 && components[len(components)-1] == "" {
		components = components[:len(components)-1]
	}

	// Skip the first component and join the remaining components with " > "
	if len(components) > 1 {
		return strings.Join(components[1:], " > ")
	}

	return ""
}

var ignoredExtensions = map[string]bool{
	".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".svg": true, ".webp": true,
	".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".m4a": true,
	".mp4": true, ".avi": true, ".mkv": true, ".flv": true, ".mov": true, ".wmv": true, ".webm": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".pages": true, ".key": true, ".numbers": true,
	".exe": true, ".msi": true, ".bin": true, ".dmg": true, ".apk": true, ".deb": true, ".rpm": true,
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true,
}

func shouldIgnoreLink(link string) bool {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return true
	}

	// Check if the URL contains a fragment
	if parsedURL.Fragment != "" {
		return true
	}

	// Check if the URL has a file extension in the ignoredExtensions map
	fileExtension := filepath.Ext(parsedURL.Path)
	if _, ok := ignoredExtensions[fileExtension]; ok {
		return true
	}

	return false
}
