package index

import (
	"strings"
	"sync"
)

// Index is a thread-safe inverted index mapping keywords to documents.
type Index struct {
	mu   sync.RWMutex
	docs map[string][]*Document // keyword -> documents
	all  []*Document            // all documents in insertion order
}

// New creates a new empty Index.
func New() *Index {
	return &Index{
		docs: make(map[string][]*Document),
	}
}

// Add indexes a document. Thread-safe for concurrent writes.
func (idx *Index) Add(doc *Document) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.all = append(idx.all, doc)

	// Index each unique word
	seen := make(map[string]bool)
	for word := range doc.WordFreq {
		w := strings.ToLower(word)
		if !seen[w] {
			idx.docs[w] = append(idx.docs[w], doc)
			seen[w] = true
		}
	}
}

// Search returns documents matching the query keyword, sorted by relevance.
// Thread-safe for concurrent reads while the crawler is writing.
func (idx *Index) Search(query string) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	keyword := strings.ToLower(strings.TrimSpace(query))
	if keyword == "" {
		return nil
	}

	docs := idx.docs[keyword]
	if len(docs) == 0 {
		return nil
	}

	results := make([]SearchResult, 0, len(docs))
	for _, doc := range docs {
		score := float64(doc.WordFreq[keyword])
		// Title match bonus
		if strings.Contains(strings.ToLower(doc.Title), keyword) {
			score += 10
		}
		results = append(results, SearchResult{
			RelevantURL: doc.URL,
			OriginURL:   doc.OriginURL,
			Depth:       doc.Depth,
			Score:       score,
			Title:       doc.Title,
		})
	}

	// Sort by score descending
	sortResults(results)
	return results
}

// Stats returns the number of indexed documents and unique keywords.
func (idx *Index) Stats() (docCount, keywordCount int) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.all), len(idx.docs)
}
