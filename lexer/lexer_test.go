package lexer

import (
	"reflect"
	"regexp"
	"testing"
	"unicode"
)

func TestNewLexer(t *testing.T) {
	l := NewLexer("Hello World!")
	if l == nil {
		t.Error("NewLexer() returned nil")
	} else {

		if len(l.content) != 12 {
			t.Error("NewLexer() returned wrong length")
		}

		if string(l.content) != "Hello World!" {
			t.Error("NewLexer() returned wrong content")
		}
	}

}

func TestTrimLeft(t *testing.T) {
	l := NewLexer(" Hello World!")
	l.TrimLeft()
	if string(l.content) != "Hello World!" {
		t.Error("TrimLeft() failed")
	}

}

func TestChop(t *testing.T) {
	l := NewLexer("Hello World!")
	l.Chop(5)
	if string(l.content) != " World!" {
		t.Error("Chop() failed")
	}
}

func TestChopWhile(t *testing.T) {
	l := NewLexer("Hello World!")

	f := func(x rune) bool {
		return unicode.IsLetter(x)
	}

	l.ChopWhile(f)
	expected := " World!"
	if string(l.content) != expected {
		t.Errorf("ChopWhile() Failed, expected %v, got %v", expected, l.content)
	}
}

func TestNextToken(t *testing.T) {

	l := NewLexer("Hello World!")

	expected := "hello"
	nextToken := l.NextToken()

	if string(nextToken) != expected {
		t.Errorf("NextToken() Failed, expected %v, got %v", expected, string(nextToken))
	}

}

func TestNext(t *testing.T) {
	l := NewLexer("Hello World!")

	expected := "hello"
	nextToken, err := l.Next()

	if err != nil {
		t.Errorf("Next() Failed, expected %v, got %v", nil, err)
	}

	if nextToken != expected {
		t.Errorf("NextToken() Failed, expected %v, got %v", expected, nextToken)
	}

	nextToken2, err := l.Next()

	if err != nil {
		t.Errorf("Next() Failed, expected %v, got %v", nil, err)
	}

	expected2 := "world"

	if nextToken2 != expected2 {
		t.Errorf("NextToken() Failed, expected %v, got %v", expected2, nextToken2)
	}

	expected3 := "!"

	nextToken3, err := l.Next()

	if err != nil {
		t.Errorf("Next() Failed, expected %v, got %v", nil, err)
	}

	if nextToken3 != expected3 {
		t.Errorf("NextToken() Failed, expected %v, got %v", expected3, nextToken3)
	}

	EOF, err := l.Next()

	if err == nil {
		t.Errorf("Next() Failed, expected %v, got %v", expected, err)
	}

	if EOF != "EOF" {
		t.Errorf("NextToken() Failed, expected %v, got %v", "EOF", EOF)
	}
}

func TestParseLinks(t *testing.T) {
	testCases := []struct {
		name          string
		htmlContent   string
		expectedLinks []string
	}{
		{
			name: "Basic test",
			htmlContent: `
<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
</head>
<body>
  <h1>Sample Links</h1>
  <a href="https://example.com/page1">Link 1</a>
  <a href="https://example.com/page2">Link 2</a>
  <a href="/page3">Link 3</a>
  <a href="/page4">Link 4</a>
</body>
</html>
`,
			expectedLinks: []string{
				"https://example.com/page1",
				"https://example.com/page2",
				"/page3",
				"/page4",
			},
		},
		// Add more test cases here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			links := ParseLinks(tc.htmlContent)
			if !reflect.DeepEqual(links, tc.expectedLinks) {
				t.Errorf("Expected: %v, got: %v", tc.expectedLinks, links)
			}
		})
	}
}

func TestParseHtmlTextContent(t *testing.T) {
	testCases := []struct {
		name                string
		htmlContent         string
		expectedTextContent string
	}{
		{
			name: "Basic test",
			htmlContent: `
<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
</head>
<body>
  <h1>Sample Links</h1>
  <a href="https://example.com/page1">Link 1</a>
  <a href="https://example.com/page2">Link 2</a>
  <a href="/page3">Link 3</a>
  <a href="/page4">Link 4</a>
</body>
</html>`,
			expectedTextContent: "Test Page Sample Links Link 1 Link 2 Link 3 Link 4",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			textContent := ParseHtmlTextContent(tc.htmlContent)
			// Remove all whitespaces and newlines from both textContent and expectedTextContent
			re := regexp.MustCompile(`\s`)
			expected := re.ReplaceAllString(tc.expectedTextContent, "")
			actual := re.ReplaceAllString(textContent, "")
			if actual != expected {
				t.Errorf("Expected: %v, got: %v", expected, actual)
			}
		})

	}

}

func TestMapToSortedSlice(t *testing.T) {

	testCases := []struct {
		name     string
		input    map[string]int
		expected []stat
	}{
		{
			name: "Basic test",
			input: map[string]int{
				"one":   1,
				"two":   2,
				"three": 3,
			},
			expected: []stat{
				{token: "three", freq: 3},
				{token: "two", freq: 2},
				{token: "one", freq: 1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sortedSlice := MapToSortedSlice(tc.input)
			if !reflect.DeepEqual(sortedSlice, tc.expected) {
				t.Errorf("Expected: %v, got: %v", tc.expected, sortedSlice)
			}
		})
	}

}
