package webcrawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"os"
	"strings"
	"sync"

	"github.com/deanrtaylor1/gosearch/src/lexer"
	"github.com/deanrtaylor1/gosearch/src/util"
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
