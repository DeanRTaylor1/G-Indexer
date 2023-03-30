package Lexer

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/deanrtaylor1/gosearch/src/Types"
)

type Lexer struct {
	content []rune
}

type stat struct {
	token string
	freq  int
}

func NewLexer(content string) *Lexer {
	return &Lexer{[]rune(content)}
}

func (l *Lexer) TrimLeft() {
	for len(l.content) > 0 && unicode.IsSpace(rune(l.content[0])) {
		l.content = l.content[1:]
	}
}

func (l *Lexer) Chop(n int) (token []rune) {
	token = l.content[:n]
	l.content = l.content[n:]
	return token
}

func (l *Lexer) ChopWhile(f func(rune) bool) (token []rune) {
	n := 0
	for n < len(l.content) && f(l.content[n]) {
		n += 1
	}
	return l.Chop(n)
}

func (l *Lexer) NextToken() []rune {

	l.TrimLeft()

	if len(l.content) == 0 {
		//fmt.Println("end of content")
		return nil
	}
	if unicode.IsNumber(l.content[0]) {
		return l.ChopWhile(unicode.IsNumber)
	}
	if unicode.IsLetter(l.content[0]) {
		return l.ChopWhile(func(r rune) bool {
			return unicode.IsLetter(r) || unicode.IsNumber(r)
		})

	}
	return l.Chop(1)
}

func (l *Lexer) Next() (string, error) {
	token := l.NextToken()
	if token == nil {
		return "EOF", errors.New("no more tokens")
	}
	return strings.ToUpper(string(token)), nil
}

/*func indexDocument(content string) map[string]int {*/
/*return*/
/*}*/

func ReadEntireXMLFile(filePath string) string {
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
func MapToSortedSlice(m map[string]int) (stats []stat) {
	for k, v := range m {
		stats = append(stats, struct {
			token string
			freq  int
		}{k, v})
	}
	sort.Slice(stats, func(i, j int) bool { return stats[i].freq > stats[j].freq })

	return stats
}

func MapToJSON(m Types.TermFreqIndex, createFile bool, filename string) string {
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println("error:", err)
	}
	if createFile {
		JSONToFile(b, filename)
	}
	return string(b)
}

func JSONToFile(j []byte, filename string) {
	f, err := os.Create(filename)
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

func LogStats(filePath string, stats []stat, topN int) {
	fmt.Println(filePath)
	if len(stats) < topN {
		for _, v := range stats {
			fmt.Println(v.token, " => ", v.freq)
		}
	} else {
		for _, v := range stats[:topN] {
			fmt.Println(v.token, " => ", v.freq)
		}
	}
}

func CheckIndex(path string) (Types.TermFreqIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer f.Close()

	var index Types.TermFreqIndex

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
