package tfidf

import (
	"math"
	"sort"

	"github.com/deanrtaylor1/gosearch/bm25"
	"github.com/deanrtaylor1/gosearch/lexer"
)

type TermFreq map[string]int
type TermFreqPerDoc map[string]DocData
type DocFreq = map[string]int

type DocData struct {
	TermCount int
	Terms     TermFreq
}

type Model struct {
	TFPD TermFreqPerDoc
	DF   DocFreq
}

// This function computes the term frequency for each document in a corpus and returns a map of the term frequencies for each document
// this uses tfidf and it is a backup to bm25 as
// the bm25 gets better results but can return 0 if the term is generic where as tfidf will increase the rank
// of the document if the term is generic
func CalculateTfidf(model *bm25.Model, query string) ([]bm25.ResultsMap, int) {
	var result []bm25.ResultsMap
	var count int
	model.ModelLock.Lock()
	defer model.ModelLock.Unlock()
	for path, table := range model.TFPD {
		querylexer := lexer.NewLexer(string(query))
		var rank float32 = 0
		for {
			token, err := querylexer.Next()
			if err != nil {
				break
			}
			rank += ComputeTF(token, table.TermCount, TermFreq(table.Terms)) * ComputeIDF(token, len(model.TFPD), model.DF)
			count += 1
		}
		result = append(result, bm25.ResultsMap{Name: model.UrlFiles[path], Path: path, TF: rank})
		sort.Slice(result, func(i, j int) bool {
			return result[i].TF > result[j].TF
		})

	}
	return result, count
}

// This function computes the term frequency of a given term in a document using tfidf
func ComputeTF(t string, N int, d TermFreq) float32 {
	//T is the term we are looking for
	//N is the total number of terms (not unique) in the document
	//d is the map of terms to their frequency in the document
	if _, ok := d[t]; ok {
		return float32(d[t]) / float32(N)
	}
	return 0
}

// Compute Inverse document frequency, that is to say, it computes the importance of a term in the collection.
// By seeing how frequent it is in all other documents vs the current document.
func ComputeIDF(t string, N int, df DocFreq) float32 {
	//N The total number of documents in the collection.

	//The number of documents in the collection that contain the term.
	//fmt.Println(df[t])
	M := float64(df[t])
	//Loop through each document in the collection and check if it contains the term t. If it does, increment M by 1.

	//If M is 0, set it to 1 to avoid division by zero errors.
	M = math.Max(M, 1)
	//using the log10 function to make the IDF values more readable

	//Calculate the IDF using the formula log(N/M), where N is the total number of documents in the collection and M is the number of documents that contain the term t.
	return float32(math.Log10(float64(N) / M))
}
