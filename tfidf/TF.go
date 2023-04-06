package tfidf

func ComputeTF(t string, N int, d TermFreq) float32 {
	//T is the term we are looking for
	//N is the total number of terms (not unique) in the document
	//d is the map of terms to their frequency in the document
	if _, ok := d[t]; ok {
		return float32(d[t]) / float32(N)
	}
	return 0
}
