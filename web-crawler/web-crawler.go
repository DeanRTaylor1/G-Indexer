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
	// Start go routine, send urls to foundUrl Channel
	//Send get request
	logger.HandleLog(fmt.Sprintf("Initiating get request to %s", urlToCrawl))
	resp, err := http.Get(urlToCrawl)

	if err != nil {
		errChan <- fmt.Errorf("error accessing site file: %w", err)
		return
	}
	defer resp.Body.Close()

	//Read html body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errChan <- fmt.Errorf("error reading html response body: %w", err)
		return
	}

	//get the full Url for later use
	fullUrl, err := url.Parse(urlToCrawl)
	if err != nil {
		log.Println(err)
	}

	//Prse html text tokens
	textContent := lexer.ParseHtmlTextContent(string(body))
	//Create model of indexed data for storage
	IndexedData := util.IndexedData{
		URL:     urlToCrawl,
		Content: textContent,
	}

	//Cache the data, ensure we lock the model before accessing, this is used for disk storage
	cachedDataMutex.Lock()
	(*cachedData)[urlToCrawl] = IndexedData
	cachedDataMutex.Unlock()

	//Update the model so that we can send progress to the end user
	model.ModelLock.Lock()
	model.DirLength += 1
	model.ModelLock.Unlock()

	//Prepare content for parsing
	content := IndexedData.Content

	fileSize := len(content)
	logger.HandleLog(fmt.Sprintf("%s => %v", IndexedData.URL, fileSize))
	// tf := make(bm25.TermFreq)
	bm25.ConvertContentToModel(content, IndexedData.URL, model)

	model.ModelLock.Lock()
	model.DocCount += 1
	model.ModelLock.Unlock()

	// extract the links from the file
	links := lexer.ParseLinks(string(body))

	//Parse the links
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
			// Resolve the relative link against the base URL to ensure we don't leave our current domain
			resolvedLink := fullUrl.ResolveReference(parsedLink)
			link = resolvedLink.String()
		}

		foundUrls <- link

	}

}

func CrawlDomainUpdateModel(domain string, model *bm25.Model) {
	logger.HandleLog(fmt.Sprintf("crawling domain: %s", domain))
	//Start timer for benchmarking
	start := time.Now()
	//Initiate data models
	cachedData := make(map[string]util.IndexedData)
	//Keep a track of Visited urls
	visited := make(map[string]bool)

	//These two maps are used to store the url and the file name for later mapping to the user.
	//We store both to save time on the reverse lookup
	urlFiles := make(map[string]string)
	reverseUrlFiles := make(map[string]string)

	//Mutexes for each shared data structure
	cachedDataMutex := sync.Mutex{}
	visitedMutex := sync.Mutex{}
	urlsMutex := sync.Mutex{}
	reverseUrlsMutex := sync.Mutex{}

	//Create a directory for the domain in the indexes folder
	fullUrl, err := url.Parse(domain)
	if err != nil {
		log.Println(err)
	}
	dirName := fmt.Sprint("indexes/" + fullUrl.Host)
	err = os.MkdirAll(dirName, os.ModePerm)

	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	//Update the model name for the user
	model.ModelLock.Lock()
	model.Name = fullUrl.Host
	model.ModelLock.Unlock()

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
		//Loop through the found urls and crawl them
		select {
		case newURL := <-foundUrls:
			visitedMutex.Lock()
			numberOfVisitedURLs := len(visited)
			if numberOfVisitedURLs >= maxURLsToCrawl {
				//If we have reached the max number of urls to crawl, we can stop the crawler, this is a failsafe for testing and to stop the crawler from running forever
				visitedMutex.Unlock()
				model.ModelLock.Lock()
				model.IsComplete = true
				model.ModelLock.Unlock()

				//Write the cached data to disk
				cachedDataMutex.Lock()
				err := bm25.CompressAndWriteGzipFile("indexed-data.gz", cachedData, dirName)
				if err != nil {
					log.Fatal(err)
				}
				cachedDataMutex.Unlock()
				//Write the url files to disk
				urlsMutex.Lock()
				err = bm25.CompressAndWriteGzipFile("url-files.gz", urlFiles, dirName)
				if err != nil {
					log.Fatal(err)
				}
				urlsMutex.Unlock()
				//Write the reverse url files to disk
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

			urlPath, err := url.Parse(newURL)
			if err != nil {
				log.Println(err)
			}
			fileName := urlToName(urlPath.Path)

			//Add the url and file name to the maps
			urlsMutex.Lock()
			urlFiles[newURL] = fileName
			reverseUrlFiles[fileName] = newURL
			urlsMutex.Unlock()

			//Add the url and file name to the model so that the user can access them immediately
			reverseUrlsMutex.Lock()
			model.ReverseUrlFiles[fileName] = newURL
			reverseUrlsMutex.Unlock()

			model.ModelLock.Lock()
			model.UrlFiles[newURL] = fileName
			model.ModelLock.Unlock()

			wg.Add(1)
			go func(urlToCrawl string) {
				defer wg.Done()
				crawlPageUpdateModel(urlToCrawl, foundUrls, dirName, errChan, &cachedDataMutex, &cachedData, model)
			}(newURL)
		//If there is an error, log it and continue
		case err := <-errChan:
			logger.HandleError(err)
		//If the crawler is complete, write the data to disk
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

// We don't want to crawl files that are not html as our search engine is text based
var ignoredExtensions = map[string]bool{
	".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".svg": true, ".webp": true,
	".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".m4a": true,
	".mp4": true, ".avi": true, ".mkv": true, ".flv": true, ".mov": true, ".wmv": true, ".webm": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".pages": true, ".key": true, ".numbers": true,
	".exe": true, ".msi": true, ".bin": true, ".dmg": true, ".apk": true, ".deb": true, ".rpm": true,
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true,
}

// shouldIgnoreLink returns true if the link should be ignored
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
