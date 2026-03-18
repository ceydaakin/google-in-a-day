package crawler

// CrawlTask represents a URL to be fetched and indexed.
type CrawlTask struct {
	URL       string `json:"url"`
	OriginURL string `json:"origin_url"`
	Depth     int    `json:"depth"`
}
