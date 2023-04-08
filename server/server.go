package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/deanrtaylor1/gosearch/bm25"
	"github.com/deanrtaylor1/gosearch/tfidf"
	"github.com/deanrtaylor1/gosearch/util"
	webcrawler "github.com/deanrtaylor1/gosearch/web-crawler"

	"github.com/tebeka/snowball"
)

type Response struct {
	Message string            `json:"Message"`
	Data    []bm25.ResultsMap `json:"Data"`
}

type IndexResponse struct {
	Message string
	Data    []string
}

type ProgressResponse struct {
	Message       string  `json:"message"`
	IsComplete    bool    `json:"is_complete"`
	IndexProgress float32 `json:"index_progress"`
	IndexName     string  `json:"index_name"`
	DocCount      int     `json:"doc_count"`
	DirLength     float32 `json:"dir_length"`
	TermCount     int     `json:"term_count"`
}

type ProgressResponseData struct {
	Name  string      `json:"data_name"`
	Value interface{} `json:"data_value"`
}

// Server route to initialize the crawl on a go routine
func handleApiCrawl(w http.ResponseWriter, r *http.Request, model *bm25.Model) {
	requestBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(requestBodyBytes))
	urlToCrawl := string(requestBodyBytes)
	_, err = url.ParseRequestURI(urlToCrawl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response, err := json.Marshal(struct{ Message string }{Message: "Invalid URL"})
		if err != nil {
			log.Println(err)
			return
		}
		_, err = w.Write(response)
		if err != nil {
			log.Println(err)
			return
		}
		return
	}

	bm25.ResetModel(model)

	go func() {
		webcrawler.CrawlDomainUpdateModel(urlToCrawl, model, bm25.FileOpsImpl{}, 10000)
		model.ModelLock.Lock()
		model.Name = urlToCrawl
		model.DA = float32(model.TermCount) / float32(model.DocCount)
		model.ModelLock.Unlock()
	}()

	response := &Response{
		Message: fmt.Sprintf("INTIALIZING CRAWLER THROUGH %v", urlToCrawl),
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Println("Unable to marshal json: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("------------------")
	fmt.Printf("/33]32m INTIALIZING CRAWLER THROUGH %v", string(requestBodyBytes))
	log.Println("------------------")

}

// Server route to get the status of the crawl and index
func handleApiProgress(w http.ResponseWriter, r *http.Request, model *bm25.Model) {
	model.ModelLock.Lock()

	if model.DocCount == 0 {
		response := ProgressResponse{
			Message:       "Not Started",
			IsComplete:    false,
			IndexProgress: 0,
			IndexName:     "",
		}
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			log.Println(util.TerminalRed+"Unable to marshal json: ", err, util.TerminalRed)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonBytes)
		if err != nil {
			log.Println(err)
			model.ModelLock.Unlock()
			return
		}
		model.ModelLock.Unlock()
		return
	}
	indexName := model.Name
	indexProgress := float32(model.DocCount) / model.DirLength
	isComplete := model.IsComplete
	docCount := model.DocCount
	dirLength := model.DirLength
	termCount := model.TermCount
	var message string
	if model.IsComplete {
		message = "Complete"
	} else {
		message = "In Progress"
	}
	model.ModelLock.Unlock()
	response := ProgressResponse{
		Message:       message,
		IsComplete:    isComplete,
		IndexProgress: indexProgress,
		IndexName:     indexName,
		DocCount:      docCount,
		DirLength:     dirLength,
		TermCount:     termCount,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Println("Unable to marshal json: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Println(err)
		return
	}
}

// Server route to start the search on a go routine
func handleApiSearch(w http.ResponseWriter, r *http.Request, model *bm25.Model) {
	start := time.Now()
	stemmer, err := snowball.New("english")
	if err != nil {
		log.Fatal(err)
	}

	defer stemmer.Close()

	requestBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(requestBodyBytes))

	var count int
	result, count := bm25.CalculateBm25(model, string(requestBodyBytes))

	var max int
	if len(result) < 20 {
		max = len(result)
	} else {
		max = 20
	}

	for i := 0; i < max; i++ {
		log.Println(result[i].Path, " => ", result[i].TF)
	}

	if err != nil {
		log.Println(err)
		return
	}

	if result[0].TF == 0 {
		log.Println("Query too generic, ranking with tf-idf")

		result, count = tfidf.CalculateTfidf(model, string(requestBodyBytes))

		for i := 0; i < max; i++ {
			log.Println(result[i].Path, " => ", result[i].TF)
		}

	}

	var data []bm25.ResultsMap

	if result[0].TF == 0 {
		data = []bm25.ResultsMap{{
			Path: "No results found",
			TF:   0,
		}}
	} else {
		data = bm25.FilterResults(result[:max], bm25.IsGreaterThanZero)
	}

	elapsed := time.Since(start)
	response := &Response{
		Message: fmt.Sprintf("Queried %d documents in %d Ms", count, elapsed.Milliseconds()),
		Data:    data,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Println("Unable to marshal json: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("------------------")
	log.Println(util.TerminalCyan+"Queried ", count, " documents in ", elapsed.Milliseconds(), " ms"+util.TerminalReset)
	log.Println("------------------")

}

// Server route to get the available indexes in the users index directory if there are any
func handleApiIndexes(w http.ResponseWriter, r *http.Request, model *bm25.Model) {
	directories := util.GetCurrentAvailableModelDirectories()

	response := &IndexResponse{
		Message: "Available indexes",
		Data:    directories,
	}

	jsonBytes, err := json.Marshal(response)

	if err != nil {
		log.Println("Unable to marshal json: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Println(err)
	}

}

// Server route to start indexing an existing directory and add it to the model
func handleApiIndex(w http.ResponseWriter, r *http.Request, model *bm25.Model) {
	log.Println("received")
	requestBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(requestBodyBytes))
	if isValid, err := util.CheckDirIsValid("./indexes/" + string(requestBodyBytes)); !isValid {
		if err != nil {
			log.Println(err)
		}
		customMessage := "Directory is not valid or does not exist"
		jsonBytes, err := json.Marshal(&struct{ Message string }{Message: customMessage})

		if err != nil {
			log.Println("Unable to marshal json: ", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonBytes)
		if err != nil {
			log.Println(err)
		}
		return
	}
	log.Println("received number 2")

	bm25.ResetModel(model)

	log.Println("Starting server and indexing directory: ", "./indexes/", string(requestBodyBytes))
	model.Name = string(requestBodyBytes)
	go func() {
		bm25.LoadCachedGobToModel("./indexes/"+string(requestBodyBytes), model)
		model.ModelLock.Lock()
		model.DA = float32(model.TermCount) / float32(model.DocCount)
		model.IsComplete = true
		model.ModelLock.Unlock()
	}()

	jsonBytes, err := json.Marshal(&struct{ Message string }{Message: "Indexing started"})

	if err != nil {
		log.Println("Unable to marshal json: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Println(err)
	}
}

// Route handler
func handleRequests(model *bm25.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		switch {
		case r.Method == "GET" && r.URL.Path == "/":
			http.ServeFile(w, r, "static/index.html")
		case r.Method == "GET" && r.URL.Path == "/favicon.ico":
			http.ServeFile(w, r, "static/favicon.ico")
		case r.Method == "GET" && r.URL.Path == "/index.html":
			http.ServeFile(w, r, "static/index.html")
		case r.Method == "GET" && r.URL.Path == "/styles.css":
			http.ServeFile(w, r, "static/styles.css")
		case r.Method == "GET" && r.URL.Path == "/index.js":
			http.ServeFile(w, r, "static/index.js")
		case r.Method == "GET" && r.URL.Path == "/api/indexes":
			handleApiIndexes(w, r, model)
		case r.Method == "GET" && r.URL.Path == "/api/progress":
			handleApiProgress(w, r, model)
		case r.Method == "POST" && r.URL.Path == "/api/crawl":
			handleApiCrawl(w, r, model)
		case r.Method == "POST" && r.URL.Path == "/api/index":
			handleApiIndex(w, r, model)
		case r.Method == "POST" && r.URL.Path == "/api/search":
			handleApiSearch(w, r, model)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 Not Found")

		}
	}
}

func Serve(model *bm25.Model) {
	http.HandleFunc("/", handleRequests(model))
	log.Println("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
