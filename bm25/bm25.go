package bm25

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/deanrtaylor1/gosearch/lexer"
	"github.com/deanrtaylor1/gosearch/util"
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
	Name string
	TFPD TermFreqPerDoc
	//DF is the Document Frequency of a term
	DF DocFreq
	//DA is the average document length
	DA         float32
	TermCount  int
	DocCount   int
	DirLength  float32
	UrlFiles   map[string]string
	ModelLock  *sync.Mutex
	IsComplete bool
}

func ResetModel(model *Model) {
	model.ModelLock.Lock()
	defer model.ModelLock.Unlock()
	model.TFPD = make(map[string]DocData)
	model.DF = make(map[string]int)
	model.UrlFiles = make(map[string]string)
	model.DocCount = 0
	model.TermCount = 0
	model.DirLength = 0
	model.DA = 0
	model.Name = ""
	model.IsComplete = false
}

func NewEmptyModel() *Model {
	return &Model{
		TFPD:      make(map[string]DocData),
		DF:        make(map[string]int),
		UrlFiles:  make(map[string]string),
		ModelLock: &sync.Mutex{},
	}
}

// func getFileUrl(filePath string) (string, error) {
// 	absolutePath, err := filepath.Abs(filePath)
// 	if err != nil {
// 		log.Println("unable to get absolute path", err)
// 		return "", err
// 	}

// 	fileUrl := &url.URL{
// 		Scheme: "file",
// 		Path:   filepath.ToSlash(absolutePath),
// 	}

// 	return fileUrl.String(), nil

// }

func readUrlFiles(dirPath string, fileName string, model *Model) {
	compressedData, err := os.ReadFile(dirPath + "/" + fileName)
	if err != nil {
		log.Println(err)
		return
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		log.Println(err)
		return
	}

	model.ModelLock.Lock()

	decoder := gob.NewDecoder(gzipReader)
	gzipReader.Close()
	var decompressedURLFiles map[string]string
	if err := decoder.Decode(&decompressedURLFiles); err != nil {
		log.Println(err)
		return
	}
	model.UrlFiles = decompressedURLFiles

	model.ModelLock.Unlock()
	fmt.Println("\033[32mmapped urls\033[0m")
}

func readCompressedFilesToModel(dirPath string, fileName string, model *Model) {
	compressedData, err := os.ReadFile(dirPath + "/" + fileName)
	if err != nil {
		log.Println(err)
		return
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		log.Println(err)
		return
	}

	decoder := gob.NewDecoder(gzipReader)
	gzipReader.Close()
	var decompressedDataMap map[string]util.IndexedData
	if err := decoder.Decode(&decompressedDataMap); err != nil {
		log.Println(err)
		return
	}
	model.ModelLock.Lock()
	model.DirLength += float32(len(decompressedDataMap))
	model.ModelLock.Unlock()

	for filePath, v := range decompressedDataMap {
		model.ModelLock.Lock()
		model.DocCount += 1
		model.ModelLock.Unlock()
		//fmt.Println(filePath)
		content := v.Content
		//fmt.Println(filePath, content)

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
			//fmt.Println(filePath, " => ", token, " => ", tf[token])
		}
		model.ModelLock.Lock()
		for token := range tf {
			model.TermCount += 1
			model.DF[token] += 1
		}
		model.ModelLock.Unlock()

		model.ModelLock.Lock()

		model.TFPD[filePath] = ConvertToDocData(tf)
		model.ModelLock.Unlock()
	}
}

func LoadCachedGobToModel(dirPath string, model *Model) {
	fmt.Println(dirPath)
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
		if fi.Name() == "url-files.gz" {
			readUrlFiles(dirPath, fi.Name(), model)
			break
		}
	}

	for _, fi := range fileInfos {
		if filepath.Ext(fi.Name()) == ".gz" && fi.Name() != "url-files.gz" {
			readCompressedFilesToModel(dirPath, fi.Name(), model)
			continue
		}
	}
	fmt.Println("------------------")
	fmt.Println(util.TerminalGreen + "FINISHED LOADING MODEL" + util.TerminalReset)
	fmt.Println("------------------")
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
	// fmt.Println(M, n, N)
	//If M is 0, set it to 1 to avoid division by zero errors.

	//using the log10 function to make the IDF values more readable

	//Calculate the IDF using the formula log(N/M), where N is the total number of documents in the collection and M is the number of documents that contain the term t.
	return float32(math.Log10(n / M))
}

// type FileWriter func([]byte, string)

// func ModelToJSON(m Model, createFile bool, filename string, fileWriter FileWriter) string {
// 	b, err := json.Marshal(m)
// 	if err != nil {
// 		fmt.Println("error:", err)
// 	}
// 	if createFile {
// 		fileWriter(b, filename)
// 	}
// 	return string(b)
// }
