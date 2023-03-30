package Util

import (
	"math"

	"github.com/deanrtaylor1/gosearch/src/Types"
)

// In it's current state this function iterates over every document for every word so it's
// Big O notation is O(n^2). This is not ideal and should be improved.
// We will implement caching to improve the performance of this function.
func IDF(t string, d Types.TermFreqIndex) float32 {
	//The total number of documents in the collection.
	N := len(d)
	//The number of documents in the collection that contain the term.
	var M float64 = 0
	//Loop through each document in the collection and check if it contains the term t. If it does, increment M by 1.
	for _, table := range d {
		if _, ok := table[t]; ok {
			M++
		}
	}
	//If M is 0, set it to 1 to avoid division by zero errors.
	M = math.Max(float64(M), 1)
	//using the log10 function to make the IDF values more readable

	//Calculate the IDF using the formula log(N/M), where N is the total number of documents in the collection and M is the number of documents that contain the term t.
	return float32(math.Log10(float64(N) / M))
}
