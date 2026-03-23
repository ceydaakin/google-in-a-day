package crawler

import (
	"net/url"
	"strings"
	"testing"
)

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

func TestParsePageComplete(t *testing.T) {
	html := `<html><head><title>Test Page</title></head>
	<body><p>Hello world</p><a href="/about">About</a></body></html>`

	base := mustParseURL("https://example.com/page")
	title, text, links := parsePage(strings.NewReader(html), base)

	if title != "Test Page" {
		t.Errorf("expected title 'Test Page', got '%s'", title)
	}
	if !strings.Contains(text, "Hello world") {
		t.Errorf("expected body to contain 'Hello world', got '%s'", text)
	}
	if len(links) == 0 {
		t.Fatal("expected at least one link")
	}
	if links[0] != "https://example.com/about" {
		t.Errorf("expected 'https://example.com/about', got '%s'", links[0])
	}
}

func TestParsePageNoTitle(t *testing.T) {
	html := `<html><body><p>No title here</p></body></html>`
	base := mustParseURL("https://example.com")
	title, _, _ := parsePage(strings.NewReader(html), base)
	if title != "" {
		t.Errorf("expected empty title, got '%s'", title)
	}
}

func TestExtractBetweenTags(t *testing.T) {
	html := `<html><title>Hello</title><body>World</body></html>`
	if got := extractBetweenTags(html, "title"); got != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", got)
	}
	if got := extractBetweenTags(html, "body"); got != "World" {
		t.Errorf("expected 'World', got '%s'", got)
	}
}

func TestExtractBetweenTagsMissing(t *testing.T) {
	html := `<html><body>World</body></html>`
	if got := extractBetweenTags(html, "title"); got != "" {
		t.Errorf("expected empty, got '%s'", got)
	}
}

func TestStripTags(t *testing.T) {
	html := `<p>Hello</p><script>var x=1;</script><p>World</p>`
	text := stripTags(html)
	if !strings.Contains(text, "Hello") || !strings.Contains(text, "World") {
		t.Errorf("expected 'Hello' and 'World', got '%s'", text)
	}
	if strings.Contains(text, "var") {
		t.Errorf("script content should be removed, got '%s'", text)
	}
}

func TestStripTagsStyle(t *testing.T) {
	html := `<style>.foo{color:red}</style><p>Content</p>`
	text := stripTags(html)
	if strings.Contains(text, "color") {
		t.Errorf("style content should be removed, got '%s'", text)
	}
	if !strings.Contains(text, "Content") {
		t.Errorf("expected 'Content', got '%s'", text)
	}
}

func TestExtractLinksAbsolute(t *testing.T) {
	html := `<a href="https://other.com/page">Link</a>`
	base := mustParseURL("https://example.com")
	links := extractLinks(html, base)
	if len(links) != 1 || links[0] != "https://other.com/page" {
		t.Errorf("expected absolute link, got %v", links)
	}
}

func TestExtractLinksRelative(t *testing.T) {
	html := `<a href="/docs/intro">Docs</a>`
	base := mustParseURL("https://example.com/page")
	links := extractLinks(html, base)
	if len(links) != 1 || links[0] != "https://example.com/docs/intro" {
		t.Errorf("expected resolved link, got %v", links)
	}
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"\r\n\tİT&#220; | Test\r", "İTÜ | Test"},
		{"  Hello   World  ", "Hello World"},
		{"&#252;niversite", "üniversite"},
		{"&amp; test", "& test"},
	}
	for _, tt := range tests {
		got := cleanText(tt.input)
		if got != tt.want {
			t.Errorf("cleanText(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractLinksFiltersJSTemplates(t *testing.T) {
	html := `<a href="' + item.url + '">JS</a><a href="https://real.com">Real</a>`
	base := mustParseURL("https://example.com")
	links := extractLinks(html, base)
	if len(links) != 1 || links[0] != "https://real.com" {
		t.Errorf("expected only real link, got %v", links)
	}
}

func TestExtractLinksFiltersFragments(t *testing.T) {
	html := `<a href="#section">Skip</a><a href="javascript:void(0)">JS</a><a href="mailto:a@b.com">Mail</a>`
	base := mustParseURL("https://example.com")
	links := extractLinks(html, base)
	if len(links) != 0 {
		t.Errorf("expected no links, got %v", links)
	}
}

func TestExtractLinksDeduplicate(t *testing.T) {
	html := `<a href="/page">A</a><a href="/page">B</a>`
	base := mustParseURL("https://example.com")
	links := extractLinks(html, base)
	if len(links) != 1 {
		t.Errorf("expected 1 deduplicated link, got %d", len(links))
	}
}

func TestResolveURL(t *testing.T) {
	base := mustParseURL("https://example.com/dir/page")

	tests := []struct {
		href string
		want string
	}{
		{"https://other.com", "https://other.com"},
		{"/absolute", "https://example.com/absolute"},
		{"relative", "https://example.com/dir/relative"},
		{"/about/", "https://example.com/about"},  // trailing slash normalized
		{"#frag", ""},
		{"javascript:alert(1)", ""},
		{"mailto:a@b.com", ""},
		{"ftp://files.com/a", ""},
	}

	for _, tt := range tests {
		got := resolveURL(tt.href, base)
		if got != tt.want {
			t.Errorf("resolveURL(%q) = %q, want %q", tt.href, got, tt.want)
		}
	}
}

func TestExtractAttr(t *testing.T) {
	tests := []struct {
		tag  string
		attr string
		want string
	}{
		{`<a href="https://example.com">`, "href", "https://example.com"},
		{`<a href='https://example.com'>`, "href", "https://example.com"},
		{`<a href=https://example.com>`, "href", "https://example.com"},
		{`<a class="link">`, "href", ""},
	}
	for _, tt := range tests {
		got := extractAttr(tt.tag, tt.attr)
		if got != tt.want {
			t.Errorf("extractAttr(%q, %q) = %q, want %q", tt.tag, tt.attr, got, tt.want)
		}
	}
}
