package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/deanrtaylor1/gosearch/src/lexer"
	"github.com/deanrtaylor1/gosearch/src/types"
	"github.com/deanrtaylor1/gosearch/src/util"
	"github.com/tebeka/snowball"
)

type Response struct {
	Message string `json:"Message"`
}

/*type SearchRequestBody struct {*/
/*Query string `json:"query"`*/
/*}*/

func handleRequests(model types.Model) http.HandlerFunc {
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
			for path, table := range model.TFPD {
				//fmt.Println(path)
				querylexer := lexer.NewLexer(string(requestBodyBytes))
				var rank float32 = 0
				for {
					token, err := querylexer.Next()
					if err != nil {
						break
					}
					//fmt.Println(Util.ComputeTF(token, table.TermCount, table.Terms), Util.ComputeIDF(token, table.TermCount, model.DF))
					rank += util.ComputeTF(token, table.TermCount, table.Terms) * util.ComputeIDF(token, len(model.TFPD), model.DF)
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
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 Not Found")

		}

	}
}

func Serve(model types.Model) {
	http.HandleFunc("/", handleRequests(model))
	fmt.Println("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
