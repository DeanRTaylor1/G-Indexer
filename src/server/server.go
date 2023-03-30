package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/deanrtaylor1/gosearch/src/Lexer"
	"github.com/deanrtaylor1/gosearch/src/Types"
	"github.com/deanrtaylor1/gosearch/src/Util"
)

type Response struct {
	Message string `json:"Message"`
}

/*type SearchRequestBody struct {*/
/*Query string `json:"query"`*/
/*}*/

func handleRequests(tfIndex Types.TermFreqIndex) http.HandlerFunc {
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
			for path, table := range tfIndex {
				//fmt.Println(path)
				querylexer := Lexer.NewLexer(string(requestBodyBytes))
				var rank float32 = 0
				for {
					token, err := querylexer.Next()
					if err != nil {
						break
					}
					rank += Util.TF(token, table) * Util.IDF(token, tfIndex)

					//stats := mapToSortedSlice(tf)
				}
				result = append(result, struct {
					Path string  `json:"path"`
					TF   float32 `json:"tf"`
				}{path, rank})
				sort.Slice(result, func(i, j int) bool {
					return result[i].TF > result[j].TF
				})

			}
			for i := 0; i < 10; i++ {
				fmt.Println(result[i].Path, " => ", result[i].TF)
			}
			// for i, v := range result {
			// 	fmt.Println(i, v.Path, " => ", v.TF)
			// }

			response := Response{"SUCCESS"}
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
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 Not Found")

		}

	}
}

func Serve(tfIndex Types.TermFreqIndex) {
	http.HandleFunc("/", handleRequests(tfIndex))
	fmt.Println("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
