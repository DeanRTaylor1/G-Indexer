package Types

type TermFreq map[string]int
type TermFreqPerDoc map[string]TermFreq
type DocFreq = map[string]int

type Model struct {
	TFPD TermFreqPerDoc
	DF   DocFreq
}
