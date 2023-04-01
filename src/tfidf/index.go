package tfidf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/deanrtaylor1/gosearch/src/lexer"
	"github.com/deanrtaylor1/gosearch/src/util"
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

func CheckIndex(path string) (TermFreqPerDoc, error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer f.Close()

	var index TermFreqPerDoc

	err = json.NewDecoder(f).Decode(&index)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// for k, v := range index {
	// 	LogStats(k, MapToSortedSlice(v), 10)
	// }
	return index, nil
}
