package webcrawler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/deanrtaylor1/gosearch/bm25"
)

func TestShouldIgnoreLink(t *testing.T) {
	cases := []struct {
		link string
		want bool
	}{
		{"/", false},
		{"javascript.info/Learn#introduction", true},
		{"http://www.google.com", false},
		{"https://www.javascript.info", false},
		{"http://www.google.com/package.zip", true},
	}

	for _, v := range cases {
		got := shouldIgnoreLink(v.link)
		if got != v.want {
			t.Errorf("shouldIgnoreLink(%q) == %t, want %t", v.link, got, v.want)
		}
	}

}

func TestUrlToName(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"/article/js-animation/width/", "Article > Js Animation > Width"},
		{"/class-inheritance", "Class Inheritance"},
		{"/async-await", "Async Await"},
		{"/task/calculator-extendable", "Task > Calculator Extendable"},
	}

	for _, v := range cases {
		got := urlToName(v.url)
		if got != v.want {
			t.Errorf("urlToName(%q) == %q, want %q", v.url, got, v.want)
		}
	}
}

func TestExtractDomain(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"https://www.javascript.info", "www.javascript.info"},
		{"https://www.javascript.info/", "www.javascript.info"},
		{"https://www.javascript.info/async-await", "www.javascript.info"},
		{"https://www.javascript.info/async-await/", "www.javascript.info"},
		{"https://www.javascript.info/async-await/async-await", "www.javascript.info"},
	}

	for _, v := range cases {
		got := extractDomain(v.url)
		if got != v.want {
			t.Errorf("extractDomain(%q) == %q, want %q", v.url, got, v.want)
		}
	}
}

func TestCrawlDomainUpdateModel(t *testing.T) {
	// Create a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `<html><body><a href="/test-page">Test Page</a></body></html>`)
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer ts.Close()

	// Create an empty BM25 model
	model := bm25.NewEmptyModel()

	done := make(chan struct{})

	// Crawl the test server
	go CrawlDomainUpdateModel(ts.URL, model, bm25.FileOpsNoOp{}, 10)
	go waitForModelCompletion(model, done)
	<-done

	if model.DocCount != 2 {
		t.Fatalf("Expected 2 documents in the model, got %d", model.DocCount)
	}

	// Reset the model and crawl with a different limit
	bm25.ResetModel(model)
	//Because the links can not actually be crawled, we expect that only the original page will be added to the model and it will break from the loop immediately
	go CrawlDomainUpdateModel(ts.URL, model, bm25.FileOpsNoOp{}, 0)
	go waitForModelCompletion(model, done)
	<-done
	if model.DocCount != 1 {
		t.Fatalf("Expected 1 documents in the model, got %d", model.DocCount)
	}

	// ...

}

func waitForModelCompletion(model *bm25.Model, done chan<- struct{}) {
	for {
		time.Sleep(10 * time.Millisecond) // Sleep for a short duration
		model.ModelLock.Lock()
		isComplete := model.IsComplete
		model.ModelLock.Unlock()
		if isComplete {
			break
		}
	}
	done <- struct{}{}
}
