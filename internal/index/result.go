package index

import "sort"

// SearchResult represents a single search result triple with score.
type SearchResult struct {
	RelevantURL string
	OriginURL   string
	Depth       int
	Score       float64
	Title       string
}

func sortResults(results []SearchResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}
