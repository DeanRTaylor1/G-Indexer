package util

import (
	"math"

	"github.com/deanrtaylor1/gosearch/src/types"
)

func ComputeIDF(t string, N int, df types.DocFreq) float32 {
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
