package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/deanrtaylor1/gosearch/src/bm25"
	"github.com/deanrtaylor1/gosearch/src/lexer"
	"github.com/deanrtaylor1/gosearch/src/tfidf"
	webcrawler "github.com/deanrtaylor1/gosearch/src/web-crawler"

	"github.com/tebeka/snowball"
)

type resultsMap struct {
	Name string  `json:"name"`
	Path string  `json:"path"`
	TF   float32 `json:"tf"`
}

type Response struct {
	Message string       `json:"Message"`
	Data    []resultsMap `json:"Data"`
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

func getTopTerms(tf bm25.TermFreq, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	var freqs []kv
	for k, v := range tf {
		freqs = append(freqs, kv{k, v})
	}

	sort.Slice(freqs, func(i, j int) bool {
		return freqs[i].Value > freqs[j].Value
	})

	var topTerms []string
	for i := 0; i < n && i < len(freqs); i++ {
		topTerms = append(topTerms, freqs[i].Key)
	}

	return topTerms
}

func isGreaterThanZero(value float32) bool {
	return value > 0
}

func filterResults(results []resultsMap, filter func(float32) bool) []resultsMap {
	var filteredResults []resultsMap
	for _, result := range results {
		if filter(result.TF) {
			filteredResults = append(filteredResults, result)
		}
	}
	return filteredResults
}

func handleApiCrawl(w http.ResponseWriter, r *http.Request, model *bm25.Model) {
	requestBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(requestBodyBytes))
	urlToCrawl := string(requestBodyBytes)
	_, err = url.ParseRequestURI(urlToCrawl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response, err := json.Marshal(struct{ Message string }{Message: "Invalid URL"})
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = w.Write(response)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}

	go func() {
		webcrawler.CrawlDomainUpdateModel(urlToCrawl, model)
		model.ModelLock.Lock()
		model.Name = urlToCrawl
		model.DA = float32(model.TermCount) / float32(model.DocCount)
		fmt.Println(model.TermCount, model.DocCount, model.DA)

		model.ModelLock.Unlock()
	}()

	response := &Response{
		Message: fmt.Sprintf("INTIALIZING CRAWLER THROUGH %v", urlToCrawl),
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Unable to marshal json: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("------------------")
	fmt.Printf("/33]32m INTIALIZING CRAWLER THROUGH %v", string(requestBodyBytes))
	fmt.Println("------------------")

}

func handleApiProgress(w http.ResponseWriter, r *http.Request, model *bm25.Model) {
	model.ModelLock.Lock()
	defer model.ModelLock.Unlock()
	fmt.Println(model.DocCount, model.DirLength)
	if model.DocCount == 0 {
		response := ProgressResponse{
			Message:       "Not Started",
			IsComplete:    false,
			IndexProgress: 0,
			IndexName:     "",
		}
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			fmt.Println("Unable to marshal json: ", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonBytes)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}
	indexName := model.Name
	indexProgress := float32(model.DocCount) / model.DirLength
	isComplete := model.IsComplete
	docCount := model.DocCount
	dirLength := model.DirLength
	termCount := model.TermCount

	response := ProgressResponse{
		Message:       "In Progress",
		IsComplete:    isComplete,
		IndexProgress: indexProgress,
		IndexName:     indexName,
		DocCount:      docCount,
		DirLength:     dirLength,
		TermCount:     termCount,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Unable to marshal json: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handleApiSearch(w http.ResponseWriter, r *http.Request, model *bm25.Model) {
	start := time.Now()
	stemmer, err := snowball.New("english")
	if err != nil {
		log.Fatal(err)
	}

	defer stemmer.Close()

	requestBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(requestBodyBytes))
	var result []resultsMap

	count := 0
	model.ModelLock.Lock()
	defer model.ModelLock.Unlock()

	for path, table := range model.TFPD {
		//fmt.Println(path)
		querylexer := lexer.NewLexer(string(requestBodyBytes))
		var rank float32 = 0
		for {
			token, err := querylexer.Next()
			if err != nil {
				break
			}

			//fmt.Println(bm25.ComputeTF(token, table.TermCount, table.Terms, model.DA), bm25.ComputeIDF(token, len(model.TFPD), model.DF), model.DA)
			rank += bm25.ComputeTF(token, table.TermCount, table.Terms, model.DA) * bm25.ComputeIDF(token, len(model.TFPD), model.DF)
			count += 1
			//stats := mapToSortedSlice(tf)
			// fmt.Println(token, " => ", rank)
		}
		result = append(result, resultsMap{model.UrlFiles[path], path, rank})
		sort.Slice(result, func(i, j int) bool {
			return result[i].TF > result[j].TF
		})

	}

	for i := 0; i < 20; i++ {
		fmt.Println(result[i].Path, " => ", result[i].TF)
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	var result2 []resultsMap
	if result[0].TF == 0 {
		fmt.Println("No results found, trying again with tfidf")
		for path, table := range model.TFPD {
			querylexer := lexer.NewLexer(string(requestBodyBytes))
			var rank float32 = 0
			for {
				token, err := querylexer.Next()
				if err != nil {
					break
				}
				rank += tfidf.ComputeTF(token, table.TermCount, tfidf.TermFreq(table.Terms)) * tfidf.ComputeIDF(token, len(model.TFPD), model.DF)
				count += 1
			}
			result2 = append(result2, resultsMap{model.UrlFiles[path], path, rank})
			sort.Slice(result2, func(i, j int) bool {
				return result2[i].TF > result2[j].TF
			})

		}

		for i := 0; i < 20; i++ {
			fmt.Println(result2[i].Path, " => ", result2[i].TF)
		}

	}
	//fmt.Println(result2)

	var data []resultsMap
	if result2 != nil {
		if result2[0].TF == 0 {
			data = []resultsMap{{
				Path: "No results found",
				TF:   0,
			}}
		} else {
			data = filterResults(result2[:20], isGreaterThanZero)
		}
	} else {
		data = filterResults(result[:20], isGreaterThanZero)
	}

	elapsed := time.Since(start)
	response := &Response{
		Message: fmt.Sprintf("Queried %d documents in %d Ms", count, elapsed.Milliseconds()),
		Data:    data,
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Unable to marshal json: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("------------------")
	fmt.Println("Queried ", count, " documents in ", elapsed.Milliseconds(), " ms")
	fmt.Println("------------------")

}

func handleRequests(model *bm25.Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method, r.URL.Path)
		switch {
		case r.Method == "GET" && r.URL.Path == "/":
			http.ServeFile(w, r, "src/static/index.html")
		case r.Method == "GET" && r.URL.Path == "/index.html":
			http.ServeFile(w, r, "src/static/index.html")
		case r.Method == "GET" && r.URL.Path == "/styles.css":
			http.ServeFile(w, r, "src/static/styles.css")
		case r.Method == "GET" && r.URL.Path == "/index.js":
			http.ServeFile(w, r, "src/static/index.js")
		case r.Method == "POST" && r.URL.Path == "/api/crawl":
			handleApiCrawl(w, r, model)
		case r.Method == "GET" && r.URL.Path == "/api/progress":
			handleApiProgress(w, r, model)
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
	fmt.Println("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
