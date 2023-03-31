package webcrawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/deanrtaylor1/gosearch/src/lexer"
)

func Crawl(url string, c chan string, recursive bool) {

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		fmt.Println(err)
		return
	}

	DirPath := "pages"
	err = os.MkdirAll(DirPath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		return
	}
	filename := strings.ReplaceAll(url[15:], "/", "-")
	fmt.Println(filename)
	f, err := os.Create("pages/" + filename + ".html")
	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		return
	}
	l, err := f.Write(body)

	if err != nil {
		fmt.Println(err)
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		if c != nil {
			c <- fmt.Sprintf("error: %s", err)
		}
		fmt.Println(err)

	}
	if recursive {
		crawlLinks(filename)
	}
	if c != nil {
		c <- url
	}
}

func crawlLinks(filename string) {
	// read the contents of the file
	file, err := os.ReadFile("pages/" + filename + ".html")
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
		wg.Add(1)
		fmt.Println(link)

		// create a closure to capture the link variable and call the Crawl function
		go func(link string) {
			Crawl("https://go.dev"+link, c, false)
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
}
