package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"unicode"

	"os"
)

type Lexer struct {
	content []rune
}

func newLexer(content string) *Lexer {
	return &Lexer{[]rune(content)}
}

func (l *Lexer) trimLeft() {
	for len(l.content) > 0 && unicode.IsSpace(rune(l.content[0])) {
		l.content = l.content[1:]
	}
}

func (l *Lexer) chop(n int) (token []rune) {
	token = l.content[:n]
	l.content = l.content[n:]
	return token
}

func (l *Lexer) chopWhile(f func(rune) bool) (token []rune) {
	n := 0
	for n < len(l.content) && f(l.content[n]) {
		n += 1
	}
	return l.chop(n)
}

func (l *Lexer) nextToken() []rune {

	l.trimLeft()

	if len(l.content) == 0 {
		fmt.Println("end of content")
		return nil
	}
	if unicode.IsNumber(l.content[0]) {
		return l.chopWhile(unicode.IsNumber)
	}
	if unicode.IsLetter(l.content[0]) {
		return l.chopWhile(func(r rune) bool {
			return unicode.IsLetter(r) || unicode.IsNumber(r)
		})

	}
	return l.chop(1)
}

func (l *Lexer) next() ([]rune, error) {
	token := l.nextToken()
	if token == nil {
		return nil, errors.New("no more tokens")
	}
	return token, nil
}

/*func indexDocument(content string) map[string]int {*/
/*return*/
/*}*/

func readEntireXMLFile(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	var content string

	d := xml.NewDecoder(f)
	for {
		t, err := d.Token()
		if err != nil {
			break
		}

		switch se := t.(type) {
		case xml.CharData:
			content += string(se)
		}
	}
	return content
}
func mapToSortedSlice(m map[string]int) (stats []struct {
	token string
	freq  int
}) {
	for k, v := range m {
		stats = append(stats, struct {
			token string
			freq  int
		}{k, v})
	}
	sort.Slice(stats, func(i, j int) bool { return stats[i].freq > stats[j].freq })

	return stats
}

func mapToJSON(m TermFreqIndex) string {
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println("error:", err)
	}
  JSONToFile(b)
	return string(b)
}

func JSONToFile(j []byte) {
	f, err := os.Create("index.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.Write(j)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
}

type TermFreq map[string]int
type TermFreqIndex map[string]TermFreq

func main() {

	/* filePath := "./docs.gl/gl4/glVertexAttribDivisor.xhtml"*/
	/*content := readEntireXMLFile(filePath)*/

	/*  allDocs := make(map[string]map[string]int)*/

	dirPath := "./docs.gl/gl4"
	dir, err := os.Open(dirPath)
	if err != nil {
		log.Fatal(err)

	}
	defer dir.Close()
	//topN := 20

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)

	}

	termFreqIndex := make(TermFreqIndex)

	for _, fi := range fileInfos {
		filePath := dirPath + "/" + fi.Name()
		fmt.Println("Indexing file: ", filePath)
		content := readEntireXMLFile(filePath)
		fileSize := len(content)

		fmt.Println(filePath, " => ", fileSize)
		tf := make(TermFreq)

		lexer := newLexer(content)
		for {
			token, err := lexer.next()
			if err != nil {
				fmt.Println("EOF")
				break
			}
			if _, ok := tf[strings.ToUpper(string(token))]; ok {
				tf[strings.ToUpper(string(token))] += 1
			} else {
				tf[strings.ToUpper(string(token))] = 1
			}
			//fmt.Println("token: ", strings.ToUpper(string(token)))
		}

		//stats := mapToSortedSlice(tf)
		termFreqIndex[filePath] = tf

		/*   fmt.Println(filePath)*/
		/*if len(stats) < topN {*/
		/*for t, v := range stats {*/
		/*fmt.Println(t, " => ", v)*/
		/*}*/
		/*} else {*/
		/*for t, v := range stats[:topN] {*/
		/*fmt.Println(t, " => ", v)*/
		/*}*/
		/*}*/

	}

  mapToJSON(termFreqIndex)

 /* for p, tf := range termFreqIndex {*/
		/*fmt.Printf("%v has %v unique terms", p, len(tf))*/
		/*fmt.Println()*/
	/*}*/

}
