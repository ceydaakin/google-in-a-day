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

	seen := make(map[string]bool)
	for word := range doc.WordFreq {
		w := strings.ToLower(word)
		if !seen[w] {
			idx.docs[w] = append(idx.docs[w], doc)
			seen[w] = true
		}
	}
}

// Search returns documents matching the query, sorted by relevance.
// Supports multi-word queries with AND semantics.
// Thread-safe for concurrent reads while the crawler is writing.
func (idx *Index) Search(query string) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	keywords := splitQuery(query)
	if len(keywords) == 0 {
		return nil
	}

	if len(keywords) == 1 {
		return idx.searchSingle(keywords[0])
	}

	return idx.searchMulti(keywords)
}

func (idx *Index) searchSingle(keyword string) []SearchResult {
	docs := idx.docs[keyword]
	if len(docs) == 0 {
		return nil
	}

	results := make([]SearchResult, 0, len(docs))
	for _, doc := range docs {
		score := float64(doc.WordFreq[keyword])
		if strings.Contains(strings.ToLower(doc.Title), keyword) {
			score += 10
		}
		results = append(results, SearchResult{
			RelevantURL: doc.URL,
			OriginURL:   doc.OriginURL,
			Depth:       doc.MaxDepth,
			Score:       score,
			Title:       doc.Title,
		})
	}

	sortResults(results)
	return results
}

func (idx *Index) searchMulti(keywords []string) []SearchResult {
	// Find documents containing ALL keywords (AND semantics)
	// Start with the smallest posting list for efficiency
	smallest := keywords[0]
	smallestLen := len(idx.docs[smallest])
	for _, kw := range keywords[1:] {
		if l := len(idx.docs[kw]); l < smallestLen {
			smallest = kw
			smallestLen = l
		}
	}

	if smallestLen == 0 {
		return nil
	}

	// Build sets for each keyword's documents
	keywordSets := make([]map[*Document]bool, len(keywords))
	for i, kw := range keywords {
		set := make(map[*Document]bool)
		for _, doc := range idx.docs[kw] {
			set[doc] = true
		}
		keywordSets[i] = set
	}

	// Intersect: only keep docs present in ALL sets
	var results []SearchResult
	for _, doc := range idx.docs[smallest] {
		inAll := true
		for _, set := range keywordSets {
			if !set[doc] {
				inAll = false
				break
			}
		}
		if !inAll {
			continue
		}

		// Combined score across all keywords
		var score float64
		for _, kw := range keywords {
			score += float64(doc.WordFreq[kw])
			if strings.Contains(strings.ToLower(doc.Title), kw) {
				score += 10
			}
		}

		results = append(results, SearchResult{
			RelevantURL: doc.URL,
			OriginURL:   doc.OriginURL,
			Depth:       doc.MaxDepth,
			Score:       score,
			Title:       doc.Title,
		})
	}

	sortResults(results)
	return results
}

// Stats returns the number of indexed documents and unique keywords.
func (idx *Index) Stats() (docCount, keywordCount int) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.all), len(idx.docs)
}

// AllDocuments returns a deep copy of all indexed documents.
func (idx *Index) AllDocuments() []Document {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	result := make([]Document, len(idx.all))
	for i, doc := range idx.all {
		wordFreq := make(map[string]int, len(doc.WordFreq))
		for k, v := range doc.WordFreq {
			wordFreq[k] = v
		}
		result[i] = Document{
			URL:       doc.URL,
			OriginURL: doc.OriginURL,
			Depth:     doc.Depth,
			MaxDepth:  doc.MaxDepth,
			Title:     doc.Title,
			Body:      doc.Body,
			WordFreq:  wordFreq,
		}
	}
	return result
}

// LoadDocuments restores documents from a persistence snapshot.
func (idx *Index) LoadDocuments(docs []Document) {
	for i := range docs {
		idx.Add(&docs[i])
	}
}

func splitQuery(query string) []string {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	var keywords []string
	for _, w := range words {
		if len(w) > 1 {
			keywords = append(keywords, w)
		}
	}
	return keywords
}
