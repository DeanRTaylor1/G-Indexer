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
	fmt.Println("parsing links", links)
	for _, link := range links {
		fmt.Println(link)
		// check if the link is a relative link
		if strings.HasPrefix(link, "/") {
			fmt.Println("link is relative")
			// create a new url with the domain
			link = fullUrl.Scheme + "://" + fullUrl.Host + link
			fmt.Println("new link", link)
		}

		// check if the link is a valid url
		_, err := url.Parse(link)
		if err != nil {
			errChan <- fmt.Errorf("error parsing url file: %w", err)
			continue
		}

		foundUrls <- link

	}

}

func CrawlDomain(domain string) {
	fmt.Println("crawling domain: ", domain)

	visited := make(map[string]bool)

	visitedMutex := sync.Mutex{}

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
			// Check if the new URL has the same domain
			if extractDomain(newURL) != extractDomain(domain) {
				fmt.Println("URL is not in the same domain: ", newURL)
				continue
			}

			visitedMutex.Lock()
			if !visited[newURL] {
				fmt.Println("URL is new, adding to the queue: ", newURL)
				visited[newURL] = true
				wg.Add(1)
				go func(urlToCrawl string) {
					defer wg.Done()
					crawlPage(urlToCrawl, foundUrls, dirName, errChan)
				}(newURL)
			} else {
				fmt.Println("URL already visited: ", newURL)
			}
			visitedMutex.Unlock()

		case err := <-errChan:
			fmt.Printf("Error: %v\n", err)

		case <-done:
			return
		}
	}
}

// Wait for all goroutines to finish

// 	urlsMutex.Lock()
// 	util.MapToJSON(urls, true, dirName+"/urls.json")
// 	urlsMutex.Unlock()
// }

// func Crawl(domain string, url string, c chan string, recursive bool, visitedMutex *sync.Mutex, visited *map[string]bool, urlsMutex *sync.Mutex, urls *map[string]string, wg *sync.WaitGroup) string {
// 	//fmt.Println((*visited))
// 	if wg != nil {
// 		defer wg.Done()
// 	}
// 	fmt.Println("crawling: ", url)
// 	visitedMutex.Lock()

// 	if (*visited)[url] {
// 		visitedMutex.Unlock()
// 		if c != nil {
// 			c <- fmt.Sprintf("error: %s", "already visited")
// 		}
// 		return fmt.Sprintf("error: %s", "already visited")
// 	}
// 	(*visited)[url] = true
// 	visitedMutex.Unlock()
// 	fmt.Println("initiating get request", url)

// 	resp, err := httpClient.Get(url)

// 	if err != nil {
// 		fmt.Println(err)
// 		if c != nil {
// 			c <- fmt.Sprintf("error: %s", err)
// 		}
// 		return fmt.Sprintf("error: %s", err)
// 	}

// 	defer resp.Body.Close()
// 	fmt.Println(resp.StatusCode)

// 	body, err := io.ReadAll(resp.Body)
// 	fmt.Println("accessing http body", url)
// 	if err != nil {
// 		if c != nil {
// 			c <- fmt.Sprintf("error: %s", err)
// 		}
// 		fmt.Println(err)
// 		return fmt.Sprintf("error: %s", err)
// 	}

// 	dirName := sanitizeDirectoryName(domain)
// 	fmt.Println("creating dir", url)
// 	err = os.MkdirAll(dirName, os.ModePerm)

// 	if err != nil {
// 		fmt.Println(err)
// 		if c != nil {
// 			c <- fmt.Sprintf("error: %s", err)
// 		}
// 		return fmt.Sprintf("error: %s", err)
// 	}
// 	filename := url[len(domain):]
// 	fmt.Println(filename, url, domain)
// 	if domain == url {
// 		filename = "index"
// 	}
// 	filename = strings.ReplaceAll(filename, "/", "_")

// 	f, err := os.Create(dirName + "/" + filename + ".html")
// 	if err != nil {
// 		fmt.Println(err)
// 		if c != nil {
// 			c <- fmt.Sprintf("error: %s", err)
// 		}
// 		return fmt.Sprintf("error: %s", err)
// 	}
// 	l, err := f.Write(body)

// 	if err != nil {
// 		fmt.Println(err)
// 		if c != nil {
// 			c <- fmt.Sprintf("error: %s", err)
// 		}
// 		f.Close()
// 		return fmt.Sprintf("error: %s", err)
// 	}
// 	fmt.Println(l, "bytes written successfully")
// 	err = f.Close()
// 	if err != nil {
// 		if c != nil {
// 			c <- fmt.Sprintf("error: %s", err)
// 		}
// 		fmt.Println(err)

// 	}

// 	urlsMutex.Lock()
// 	(*urls)[filename+".html"] = url
// 	urlsMutex.Unlock()
// 	fmt.Println("finished crawling: ", url)

// 	if recursive {
// 		crawlLinks(domain, filename, visitedMutex, visited, urlsMutex, urls)
// 	}
// 	// if c != nil {
// 	// 	wg.Done()
// 	// }

// 	return dirName
// }

// func crawlLinks(domain string, filename string, visitedMutex *sync.Mutex, visited *map[string]bool, urlsMutex *sync.Mutex, urls *map[string]string) {
// 	// read the contents of the file
// 	dirName := sanitizeDirectoryName(domain)

// 	file, err := os.ReadFile(dirName + "/" + filename + ".html")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// extract the links from the file
// 	links := lexer.ParseLinks(string(file))

// 	// create a channel to communicate with the Go Routines
// 	c := make(chan string, 10)

// 	// create a wait group to wait for all the Go Routines to finish
// 	wg := sync.WaitGroup{}

// 	// loop through the links and create a Go Routine for each link
// 	for _, link := range links {
// 		url, err := normalizeUrl(domain, link)
// 		//fmt.Println(url, err)
// 		if err != nil {
// 			fmt.Println(err)
// 			continue
// 		}
// 		wg.Add(1)

// 		// create a closure to capture the link variable and call the Crawl function
// 		go func(link string) {
// 			Crawl(domain, url, c, true, visitedMutex, visited, urlsMutex, urls, &wg)
// 			//wg.Done()
// 		}(link)
// 		if len(c) >= 10 {
// 			<-c
// 		}
// 	}

// 	// create another Go Routine to wait for the wait group to finish and close the channel

// 	wg.Wait()
// 	close(c)

// 	// read from the channel and print the urls
// 	for url := range c {
// 		fmt.Println(url)
// 	}
// 	//fmt.Println(*urls)
// 	//mutex.Lock()
// 	//defer mutex.Unlock()
// 	//lexer.MapToJSON(*urls, true, dirName+"/urls.json")

// }

// func normalizeUrl(baseUrl string, href string) (string, error) {
// 	// Parse the base URL
// 	base, err := nu.Parse(baseUrl)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Parse the relative URL
// 	rel, err := nu.Parse(href)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Resolve the relative URL against the base URL
// 	abs := base.ResolveReference(rel)
// 	// Check if the resulting URL is within the specified domain
// 	if abs.Host != base.Host {
// 		return "", errors.New("URL not within specified domain")
// 	}

// 	if !strings.HasPrefix(abs.String(), baseUrl) {
// 		return "", fmt.Errorf("URL not within domain: %s", abs.String())
// 	}

// 	// Return the absolute URL as a string
// 	return abs.String(), nil
// }

// func sanitizeDirectoryName(dirName string) string {
// 	// Regular expression to match characters not allowed in a directory name
// 	// See: https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#naming-conventions
// 	var invalidChars = regexp.MustCompile(`[\x00-\x1f<>:"/\\|?*\x7f]`)

// 	// Replace invalid characters with a space
// 	sanitized := invalidChars.ReplaceAllString(dirName, " ")

// 	// Remove leading/trailing spaces and dots
// 	sanitized = strings.Trim(sanitized, " .")

// 	// Remove any remaining spaces and replace them with underscores
// 	sanitized = strings.ReplaceAll(sanitized, " ", "_")

// 	sanitized = strings.ReplaceAll(sanitized, ":", "")

// 	sanitized = strings.ReplaceAll(sanitized, "https", "")

// 	sanitized = strings.ReplaceAll(sanitized, "http", "")

// 	return sanitized
// }
