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

type Response struct {
	Message string `json:"Message"`
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
				var result []struct {
					Path string  `json:"path"`
					TF   float32 `json:"tf"`
				}

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
					result = append(result, struct {
						Path string  `json:"path"`
						TF   float32 `json:"tf"`
					}{path, rank})
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
				jsonBytes, err := json.Marshal(result[:20])

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
				elapsed := time.Since(start)
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
				var result []struct {
					Path string  `json:"path"`
					TF   float32 `json:"tf"`
				}

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
					result = append(result, struct {
						Path string  `json:"path"`
						TF   float32 `json:"tf"`
					}{path, rank})
					sort.Slice(result, func(i, j int) bool {
						return result[i].TF > result[j].TF
					})

				}
				for i := 0; i < 20; i++ {
					fmt.Println(result[i].Path, " => ", result[i].TF)
				}

				// for i, v := range v.UrlFiles {
				// 	fmt.Println(i, v)
				// }

				if result[0].TF > 0 {
					if v.UrlFiles != nil {
						for i := range result {
							paths := strings.Split(result[i].Path, "/")
							result[i].Path = v.UrlFiles[paths[len(paths)-1]]
						}

					}

				}
				jsonBytes, err := json.Marshal(result[:20])

				if err != nil {
					fmt.Println(err)
					return
				}

				if result[0].TF == 0 {
					fmt.Println("No results found, trying again with tfidf")
					var result2 []struct {
						Path string  `json:"path"`
						TF   float32 `json:"tf"`
					}
					for path, table := range v.TFPD {

						querylexer := lexer.NewLexer(string(requestBodyBytes))
						var rank float32 = 0
						for {
							token, err := querylexer.Next()
							if err != nil {
								break
							}
							//fmt.Println(Util.ComputeTF(token, table.TermCount, table.Terms), Util.ComputeIDF(token, table.TermCount, model.DF))
							rank += tfidf.ComputeTF(token, table.TermCount, tfidf.TermFreq(table.Terms)) * tfidf.ComputeIDF(token, len(v.TFPD), v.DF)
							count += 1
							//stats := mapToSortedSlice(tf)
							//fmt.Println(token, " => ", rank)
						}
						result2 = append(result, struct {
							Path string  `json:"path"`
							TF   float32 `json:"tf"`
						}{path, rank})
						sort.Slice(result2, func(i, j int) bool {
							return result2[i].TF > result2[j].TF
						})

					}
					// for i := 0; i < 20; i++ {
					// 	fmt.Println(result2[i].Path, " => ", result2[i].TF)
					// }

					if v.UrlFiles != nil {
						for i := range result {
							paths := strings.Split(result2[i].Path, "/")
							result2[i].Path = v.UrlFiles[paths[len(paths)-1]]
						}
					}
					jsonBytes, err = json.Marshal(result2[:20])
					if err != nil {
						fmt.Println(err)
						return
					}
					for i := 0; i < 20; i++ {
						fmt.Println(result2[i].Path, " => ", result2[i].TF)
					}

				}

				// for i, v := range result {
				// 	fmt.Println(i, v.Path, " => ", v.TF)
				// }

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
				elapsed := time.Since(start)
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
