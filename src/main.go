package main

import (
	"encoding/xml"
	"errors"
	"fmt"
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

func (l *Lexer) nextToken() []rune {

	l.trimLeft()

	if len(l.content) == 0 {
		fmt.Println("end of content")
		return nil
	}
	if unicode.IsNumber(l.content[0]) {
		n := 0
		for n < len(l.content) && (unicode.IsNumber(l.content[n])) {
			n += 1
		}
		token := l.content[:n]
		l.content = l.content[n:]

		return token
	}
	if unicode.IsLetter(l.content[0]) {
		n := 0
		for n < len(l.content) && (unicode.IsLetter(l.content[n]) || unicode.IsNumber(l.content[n])) {
			n += 1
		}
		token := l.content[:n]
		l.content = l.content[n:]

		return token
	}
	token := l.content[:1]
	l.content = l.content[1:]
	return token
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

func main() {

	/*  allDocs := make(map[string]map[string]int)*/

	/*dirPath := "./docs.gl/gl4"*/
	/*dir, err := os.Open(dirPath)*/
	/*if err != nil {*/
	/*log.Fatal(err)*/

	/*}*/
	/*defer dir.Close()*/

	/*fileInfos, err := dir.Readdir(-1)*/
	/*if err != nil {*/
	/*log.Fatal(err)*/

	/*}*/

	/*for _, fi := range fileInfos {*/
	/*filePath := dirPath + "/" + fi.Name()*/
	/*content := readEntireXMLFile(filePath)*/
	/*fileSize := len(content)*/

	/*fmt.Println(filePath,  " => ",  fileSize)*/
	/*}*/

	filePath := "./docs.gl/gl4/glVertexAttribDivisor.xhtml"
	content := readEntireXMLFile(filePath)
	lexer := newLexer(content)
	for {
		token, err := lexer.next()
		if err != nil {
			fmt.Println("EOF")
			break
		}
		fmt.Println("token: ", string(token))
	}

}
