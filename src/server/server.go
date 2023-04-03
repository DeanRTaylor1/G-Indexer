package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/deanrtaylor1/gosearch/src/bm25"
	"github.com/deanrtaylor1/gosearch/src/lexer"
	"github.com/deanrtaylor1/gosearch/src/tfidf"

	"github.com/tebeka/snowball"
)

type resultsMap struct {
	Path string  `json:"path"`
	TF   float32 `json:"tf"`
}

type Response struct {
	Message string       `json:"Message"`
	Data    []resultsMap `json:"Data"`
}

/*type SearchRequestBody struct {*/
/*Query string `json:"query"`*/
/*}*/

func handleRequests(model interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method, r.URL.Path)
		switch {
		case r.Method == "GET" && r.URL.Path == "/":
			http.ServeFile(w, r, "src/static/index.html")
		case r.Method == "GET" && r.URL.Path == "/index.html":
			http.ServeFile(w, r, "src/static/index.html")
		case r.Method == "GET" && r.URL.Path == "/index.js":
			http.ServeFile(w, r, "src/static/index.js")
		case r.Method == "POST" && r.URL.Path == "/api/search":
			start := time.Now()
			stemmer, err := snowball.New("english")
			if err != nil {
				log.Fatal(err)
			}

			defer stemmer.Close()
			switch v := model.(type) {
			case tfidf.Model:

				requestBodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(requestBodyBytes))
				var result []resultsMap

				count := 0
				for path, table := range v.TFPD {
					//fmt.Println(path)
					querylexer := lexer.NewLexer(string(requestBodyBytes))
					var rank float32 = 0
					for {
						token, err := querylexer.Next()
						if err != nil {
							break
						}
						//fmt.Println(Util.ComputeTF(token, table.TermCount, table.Terms), Util.ComputeIDF(token, table.TermCount, model.DF))
						rank += tfidf.ComputeTF(token, table.TermCount, table.Terms) * tfidf.ComputeIDF(token, len(v.TFPD), v.DF)
						count += 1
						//stats := mapToSortedSlice(tf)
						//fmt.Println(token, " => ", rank)
					}
					result = append(result, resultsMap{path, rank})
					sort.Slice(result, func(i, j int) bool {
						return result[i].TF > result[j].TF
					})

				}

				for i := 0; i < 20; i++ {
					fmt.Println(result[i].Path, " => ", result[i].TF)
				}
				// for i, v := range result {
				// 	fmt.Println(i, v.Path, " => ", v.TF)
				// }
				elapsed := time.Since(start)
				response := &Response{
					Message: fmt.Sprintf("Queried %d documents in %d Ms", count, elapsed.Milliseconds()),
					Data:    result[:20],
				}
				jsonBytes, err := json.Marshal(response)

				if err != nil {
					fmt.Println(err)
					return
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
			case bm25.Model:
				requestBodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(requestBodyBytes))
				var result []resultsMap

				count := 0
				for path, table := range v.TFPD {
					//fmt.Println(path)
					querylexer := lexer.NewLexer(string(requestBodyBytes))
					var rank float32 = 0
					for {
						token, err := querylexer.Next()
						if err != nil {
							break
						}

						// fmt.Println(bm25.ComputeTF(token, table.TermCount, table.Terms, v.DA), bm25.ComputeIDF(token, len(v.TFPD), v.DF))
						rank += bm25.ComputeTF(token, table.TermCount, table.Terms, v.DA) * bm25.ComputeIDF(token, len(v.TFPD), v.DF)
						count += 1
						//stats := mapToSortedSlice(tf)
						// fmt.Println(token, " => ", rank)
					}
					result = append(result, resultsMap{path, rank})
					sort.Slice(result, func(i, j int) bool {
						return result[i].TF > result[j].TF
					})

				}
				for i := 0; i < 20; i++ {
					fmt.Println(result[i].Path, " => ", result[i].TF)
				}

				for i, v := range v.UrlFiles {
					fmt.Println(i, v)
				}

				if result[0].TF > 0 && v.UrlFiles != nil {
					for i := range result {
						paths := strings.Split(result[i].Path, "/")
						fmt.Println(paths)
						result[i].Path = v.UrlFiles[paths[len(paths)-1]]

					}

				}

				if err != nil {
					fmt.Println(err)
					return
				}
				var result2 []resultsMap
				if result[0].TF == 0 {
					fmt.Println("No results found, trying again with tfidf")

					for path, table := range v.TFPD {

						querylexer := lexer.NewLexer(string(requestBodyBytes))
						var rank float32 = 0
						for {
							token, err := querylexer.Next()
							if err != nil {
								break
							}
							rank += tfidf.ComputeTF(token, table.TermCount, tfidf.TermFreq(table.Terms)) * tfidf.ComputeIDF(token, len(v.TFPD), v.DF)
							count += 1
						}
						result2 = append(result2, resultsMap{path, rank})
						sort.Slice(result2, func(i, j int) bool {
							return result2[i].TF > result2[j].TF
						})

					}

					if v.UrlFiles != nil {
						for i := range result {
							paths := strings.Split(result2[i].Path, "/")
							result2[i].Path = v.UrlFiles[paths[len(paths)-1]]
						}
					}

					for i := 0; i < 20; i++ {
						fmt.Println(result2[i].Path, " => ", result2[i].TF)
					}

				}
				var data []resultsMap
				if result2 != nil {
					if result2[0].TF == 0 {
						data = []resultsMap{{
							Path: "No results found",
							TF:   0,
						}}
					} else {
						data = result2[:20]
					}
				} else {
					data = result[:20]
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
			default:
				fmt.Println("Unknown model type")
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 Not Found")

		}

	}
}

func Serve(model interface{}) {
	http.HandleFunc("/", handleRequests(model))
	fmt.Println("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
