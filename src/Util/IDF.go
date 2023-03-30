package Util

import (
	"math"

	"github.com/deanrtaylor1/gosearch/src/Types"
)

func IDF(t string, d Types.TermFreqIndex) float32 {
	N := len(d)
	var M float64 = 0

	for _, table := range d {
		if _, ok := table[t]; ok {
			M++
		}
	}

	M = math.Max(float64(M), 1)
	//using the log10 function to make the IDF values more readable
	return float32(math.Log10(float64(N) / M))
}
