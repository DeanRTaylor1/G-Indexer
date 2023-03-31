package Util

import "github.com/deanrtaylor1/gosearch/src/Types"

func ComputeTF(t string, d Types.TermFreq) float32 {
	sum := 0
	for _, v := range d {
		sum += v
	}
	if _, ok := d[t]; ok {
		return float32(d[t]) / float32(sum)
	}
	return 0
}
