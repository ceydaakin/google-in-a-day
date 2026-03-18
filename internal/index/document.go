package index

// Document represents a crawled and parsed web page.
type Document struct {
	URL       string         `json:"url"`
	OriginURL string         `json:"origin_url"`
	Depth     int            `json:"depth"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	WordFreq  map[string]int `json:"word_freq"`
}
