package crawler

// CrawlTask represents a URL to be fetched and indexed.
type CrawlTask struct {
	URL       string
	OriginURL string
	Depth     int
}
