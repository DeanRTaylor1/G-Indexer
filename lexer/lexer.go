package lexer

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/net/html"

	"github.com/tebeka/snowball"
)

type Lexer struct {
	content []rune
}

type stat struct {
	token string
	freq  int
}

// NewLexer creates a new Lexer
func NewLexer(content string) *Lexer {
	return &Lexer{[]rune(content)}
}

// TrimLeft trims empty spaces from the left of the content
func (l *Lexer) TrimLeft() {
	for len(l.content) > 0 && unicode.IsSpace(rune(l.content[0])) {
		l.content = l.content[1:]
	}
}

// Chop chops the content by n and returns the chopped content
func (l *Lexer) Chop(n int) (token []rune) {
	token = l.content[:n]
	l.content = l.content[n:]
	return token
}

// ChopWhile chops the content while the predicate f returns true
func (l *Lexer) ChopWhile(f func(rune) bool) (token []rune) {
	n := 0
	for n < len(l.content) && f(l.content[n]) {
		n += 1
	}
	return l.Chop(n)
}

// NextToken returns the next token
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
		stemmer, err := snowball.New("english")
		if err != nil {
			fmt.Println(err)
		}
		defer stemmer.Close()

		term := l.ChopWhile(func(r rune) bool {
			return unicode.IsLetter(r) || unicode.IsNumber(r)
		})

		return []rune(stemmer.Stem(strings.ToLower(string(term))))

	}
	return l.Chop(1)
}

// Next returns the next token as a string
func (l *Lexer) Next() (string, error) {

	token := l.NextToken()
	if token == nil {
		return "EOF", errors.New("no more tokens")
	}
	return (string(token)), nil
}

// Tokenize parses a html string and returns all the links as a slice of strings
func ParseLinks(htmlContent string) []string {
	links := []string{}
	nodes, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Println(err)
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					links = append(links, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(nodes)
	return links
}

// Tokenize parses a html string and returns all the words in the document as a slice of strings
func ParseHtmlTextContent(htmlContent string) string {
	var content string

	d := html.NewTokenizer(strings.NewReader(htmlContent))
	for {
		tt := d.Next()
		switch tt {
		case html.ErrorToken:
			return content
		case html.TextToken:
			content += string(d.Text())
		}
	}
}

// Utility function to sort a map by value
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
