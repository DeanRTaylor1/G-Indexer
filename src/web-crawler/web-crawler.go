package webcrawler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	nu "net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/deanrtaylor1/gosearch/src/lexer"
)

func Crawl(domain string, url string, c chan string, recursive bool, mutex *sync.Mutex, visited *map[string]bool, urls *map[string]string) string {
	//fmt.Println((*visited))
	if (*visited)[url] {
		if c != nil {
			c <- fmt.Sprintf("error: %s", "already visited")
		}
		return fmt.Sprintf("error: %s", "already visited")
	}
	mutex.Lock()
	(*visited)[url] = true
	mutex.Unlock()

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		return fmt.Sprintf("error: %s", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		fmt.Println(err)
		return fmt.Sprintf("error: %s", err)
	}

	dirName := sanitizeDirectoryName(domain)
	err = os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		return fmt.Sprintf("error: %s", err)
	}
	filename := url[len(domain):]
	fmt.Println(filename, url, domain)
	if domain == url {
		filename = "index"
	}
	filename = strings.ReplaceAll(filename, "/", "_")

	f, err := os.Create(dirName + "/" + filename + ".html")
	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		return fmt.Sprintf("error: %s", err)
	}
	l, err := f.Write(body)

	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		f.Close()
		return fmt.Sprintf("error: %s", err)
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		fmt.Println(err)

	}

	// mutex.Lock()
	// fmt.Println("adding to urls", filename+".html", url)
	// (*urls)[filename+".html"] = url
	// mutex.Unlock()

	if recursive {
		crawlLinks(domain, filename, mutex, visited, urls)
	}
	if c != nil {
		c <- url
	}
	return dirName
}

func crawlLinks(domain string, filename string, mutex *sync.Mutex, visited *map[string]bool, urls *map[string]string) {
	// read the contents of the file
	dirName := sanitizeDirectoryName(domain)

	file, err := os.ReadFile(dirName + "/" + filename + ".html")
	if err != nil {
		log.Fatal(err)
	}

	// extract the links from the file
	links := lexer.ParseLinks(string(file))

	// create a channel to communicate with the Go Routines
	c := make(chan string)

	// create a wait group to wait for all the Go Routines to finish
	wg := sync.WaitGroup{}

	// loop through the links and create a Go Routine for each link
	for _, link := range links {
		url, err := normalizeUrl(domain, link)
		fmt.Println(url, err)
		if err != nil {
			fmt.Println(err)
			continue
		}
		wg.Add(1)

		// create a closure to capture the link variable and call the Crawl function
		go func(link string) {
			Crawl(domain, url, c, true, mutex, visited, urls)
			wg.Done()
		}(link)
	}

	// create another Go Routine to wait for the wait group to finish and close the channel
	go func() {
		wg.Wait()
		close(c)
	}()

	// read from the channel and print the urls
	for url := range c {
		fmt.Println(url)
	}
	// fmt.Println(*urls)
	// mutex.Lock()
	// defer mutex.Unlock()
	// lexer.MapToJSON(*urls, true, dirName+"/urls.json")

}

func normalizeUrl(baseUrl string, href string) (string, error) {
	// Parse the base URL
	base, err := nu.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	// Parse the relative URL
	rel, err := nu.Parse(href)
	if err != nil {
		return "", err
	}

	// Resolve the relative URL against the base URL
	abs := base.ResolveReference(rel)
	// Check if the resulting URL is within the specified domain
	if abs.Host != base.Host {
		return "", errors.New("URL not within specified domain")
	}

	if !strings.HasPrefix(abs.String(), baseUrl) {
		return "", fmt.Errorf("URL not within domain: %s", abs.String())
	}

	// Return the absolute URL as a string
	return abs.String(), nil
}

func sanitizeDirectoryName(dirName string) string {
	// Regular expression to match characters not allowed in a directory name
	// See: https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#naming-conventions
	var invalidChars = regexp.MustCompile(`[\x00-\x1f<>:"/\\|?*\x7f]`)

	// Replace invalid characters with a space
	sanitized := invalidChars.ReplaceAllString(dirName, " ")

	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")

	// Remove any remaining spaces and replace them with underscores
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	sanitized = strings.ReplaceAll(sanitized, ":", "")

	sanitized = strings.ReplaceAll(sanitized, "https", "")

	sanitized = strings.ReplaceAll(sanitized, "http", "")

	return sanitized
}
