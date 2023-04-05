package webcrawler

import "testing"

func TestShouldIgnoreLink(t *testing.T) {
	cases := []struct {
		link string
		want bool
	}{
		{"/", false},
		{"javascript.info/Learn#introduction", true},
		{"http://www.google.com", false},
		{"https://www.javascript.info", false},
		{"http://www.google.com/package.zip", true},
	}

	for _, v := range cases {
		got := shouldIgnoreLink(v.link)
		if got != v.want {
			t.Errorf("shouldIgnoreLink(%q) == %t, want %t", v.link, got, v.want)
		}
	}

}

func TestUrlToName(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"/article/js-animation/width/", "Article > Js Animation > Width"},
		{"/class-inheritance", "Class Inheritance"},
		{"/async-await", "Async Await"},
		{"/task/calculator-extendable", "Task > Calculator Extendable"},
	}

	for _, v := range cases {
		got := urlToName(v.url)
		if got != v.want {
			t.Errorf("urlToName(%q) == %q, want %q", v.url, got, v.want)
		}
	}
}

func TestExtractDomain(t *testing.T) {
	cases := []struct {
		url  string
		want string
	}{
		{"https://www.javascript.info", "www.javascript.info"},
		{"https://www.javascript.info/", "www.javascript.info"},
		{"https://www.javascript.info/async-await", "www.javascript.info"},
		{"https://www.javascript.info/async-await/", "www.javascript.info"},
		{"https://www.javascript.info/async-await/async-await", "www.javascript.info"},
	}

	for _, v := range cases {
		got := extractDomain(v.url)
		if got != v.want {
			t.Errorf("extractDomain(%q) == %q, want %q", v.url, got, v.want)
		}
	}
}
