package index

// Document represents a crawled and parsed web page.
type Document struct {
	URL       string
	OriginURL string
	Depth     int
	Title     string
	Body      string
	WordFreq  map[string]int
}
