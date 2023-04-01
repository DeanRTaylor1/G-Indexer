package bm25

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/deanrtaylor1/gosearch/src/lexer"
	"github.com/deanrtaylor1/gosearch/src/util"
)

const (
	k1 = 1.2
	b  = 0.75
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
	//DF is the Document Frequency of a term
	DF DocFreq
	//DA is the average document length
	DA        float32
	TermCount int
	DocCount  int
}

func AddFolderToModel(dirPath string, model *Model) {

	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	for _, fi := range fileInfos {
		if fi.IsDir() {
			if fi.Name() == "specs" {
				continue
			}
			subDirPath := filepath.Join(dirPath, fi.Name())
			AddFolderToModel(subDirPath, model)
		}
		model.DocCount += 1
		switch filepath.Ext(fi.Name()) {
		case ".xhtml", ".xml":
			filePath := dirPath + "/" + fi.Name()
			fmt.Println("Indexing file: ", filePath)
			content := lexer.ReadEntireXMLFile(filePath)
			fileSize := len(content)

			fmt.Println(filePath, " => ", fileSize)
			tf := make(TermFreq)

			lexer := lexer.NewLexer(content)
			for {
				token, err := lexer.Next()
				if err != nil {
					fmt.Println("EOF")
					break
				}

				tf[token] += 1
				//stats := mapToSortedSlice(tf)
			}
			for token := range tf {
				model.TermCount += 1
				model.DF[token] += 1
			}

			model.TFPD[filePath] = ConvertToDocData(tf)

		case ".html":
			fmt.Println("TODO IMPLEMENT HTML PARSER")
			filePath := dirPath + "/" + fi.Name()
			fmt.Println("Indexing file: ", filePath)
			content := lexer.ReadEntireHTMLFile(filePath)
			fileSize := len(content)

			fmt.Println(filePath, " => ", fileSize)
			tf := make(TermFreq)

			lexer := lexer.NewLexer(content)
			for {
				token, err := lexer.Next()
				if err != nil {
					fmt.Println("EOF")
					break
				}

				tf[token] += 1
				//stats := mapToSortedSlice(tf)
			}
			for token := range tf {
				model.DF[token] += 1
			}

			model.TFPD[filePath] = ConvertToDocData(tf)

		default:
			fmt.Fprint(os.Stderr, "\033[31mSkipping file:", fi.Name(), "(not HTML. .xhtml or .xml)\033[0m")
			fmt.Println()
			continue
		}
	}

}

func ConvertToDocData(tf TermFreq) DocData {
	var termCount int

	for _, freq := range tf {
		termCount += freq

	}

	docData := &DocData{
		TermCount: termCount,
		Terms:     tf,
	}
	return *docData
}

func ComputeTF(t string, n int, d TermFreq, DA float32) float32 {
	//t is the term we are looking for
	//n is the total number of terms (not unique) in the document
	//d is the map of terms to their frequency in the document
	//da is the average document length found in the model

	if _, ok := d[t]; ok {
		M := float32(d[t]) * (k1 + 1)
		N := float32(d[t]) + (k1 * (1 - b + (b * (float32(n) / DA))))

		return float32(M) / float32(N)
	}
	return 0
}

func ComputeIDF(t string, N int, df DocFreq) float32 {
	//N The total number of documents in the collection.

	//df The number of documents in the collection that contain the term.
	//fmt.Println(df[t])

	M := float64(df[t]) + 0.5

	n := math.Max(float64(N)-float64(df[t])+0.5, M)

	//If M is 0, set it to 1 to avoid division by zero errors.

	//using the log10 function to make the IDF values more readable

	//Calculate the IDF using the formula log(N/M), where N is the total number of documents in the collection and M is the number of documents that contain the term t.
	return float32(math.Log10(n / M))
}

func ModelToJSON(m Model, createFile bool, filename string) string {
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println("error:", err)
	}
	if createFile {
		util.JSONToFile(b, filename)
	}
	return string(b)
}
