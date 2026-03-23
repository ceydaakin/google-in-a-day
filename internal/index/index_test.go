package index

import (
	"sync"
	"testing"
)

func newTestDoc(url, title string, wordFreq map[string]int) *Document {
	return &Document{
		URL:       url,
		OriginURL: "https://origin.com",
		Depth:     1,
		MaxDepth:  3,
		Title:     title,
		WordFreq:  wordFreq,
	}
}

func TestNewIndexEmpty(t *testing.T) {
	idx := New()
	docs, keywords := idx.Stats()
	if docs != 0 || keywords != 0 {
		t.Errorf("expected 0/0, got %d/%d", docs, keywords)
	}
}

func TestAddAndSearchSingle(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://example.com", "Go Tutorial", map[string]int{
		"go": 5, "tutorial": 3,
	}))

	results := idx.Search("go")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].RelevantURL != "https://example.com" {
		t.Errorf("expected example.com, got %s", results[0].RelevantURL)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://example.com", "Test", map[string]int{"test": 1}))

	results := idx.Search("")
	if results != nil {
		t.Errorf("expected nil, got %v", results)
	}
}

func TestSearchNonExistentKeyword(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://example.com", "Test", map[string]int{"test": 1}))

	results := idx.Search("nonexistent")
	if results != nil {
		t.Errorf("expected nil, got %v", results)
	}
}

func TestSearchSortedByScore(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://low.com", "Page", map[string]int{"go": 1}))
	idx.Add(newTestDoc("https://high.com", "Go Page", map[string]int{"go": 10}))

	results := idx.Search("go")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].RelevantURL != "https://high.com" {
		t.Errorf("expected high.com first (higher score), got %s", results[0].RelevantURL)
	}
}

func TestSearchTitleBonus(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://notitle.com", "Other", map[string]int{"go": 5}))
	idx.Add(newTestDoc("https://titled.com", "Go Programming", map[string]int{"go": 1}))

	results := idx.Search("go")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// titled.com: score = 1 + 10 (title bonus) = 11
	// notitle.com: score = 5
	if results[0].RelevantURL != "https://titled.com" {
		t.Errorf("expected titled.com first (title bonus), got %s", results[0].RelevantURL)
	}
}

func TestMultiWordSearchAND(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://both.com", "Both", map[string]int{"go": 3, "tutorial": 2}))
	idx.Add(newTestDoc("https://goonly.com", "Go Only", map[string]int{"go": 5}))
	idx.Add(newTestDoc("https://tutonly.com", "Tutorial Only", map[string]int{"tutorial": 4}))

	results := idx.Search("go tutorial")
	if len(results) != 1 {
		t.Fatalf("expected 1 result (AND semantics), got %d", len(results))
	}
	if results[0].RelevantURL != "https://both.com" {
		t.Errorf("expected both.com, got %s", results[0].RelevantURL)
	}
}

func TestMultiWordSearchCombinedScoring(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://low.com", "Page", map[string]int{"go": 1, "web": 1}))
	idx.Add(newTestDoc("https://high.com", "Go Web", map[string]int{"go": 5, "web": 3}))

	results := idx.Search("go web")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].RelevantURL != "https://high.com" {
		t.Errorf("expected high.com first, got %s", results[0].RelevantURL)
	}
}

func TestSingleWordFallback(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://example.com", "Test", map[string]int{"test": 1}))

	results := idx.Search("test")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestStats(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://a.com", "A", map[string]int{"go": 1, "web": 1}))
	idx.Add(newTestDoc("https://b.com", "B", map[string]int{"go": 1, "api": 1}))

	docs, keywords := idx.Stats()
	if docs != 2 {
		t.Errorf("expected 2 docs, got %d", docs)
	}
	if keywords != 3 { // go, web, api
		t.Errorf("expected 3 keywords, got %d", keywords)
	}
}

func TestConcurrentSafety(t *testing.T) {
	idx := New()
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				idx.Add(newTestDoc("https://example.com/page", "Test", map[string]int{"go": 1}))
			}
		}(i)
	}

	// Readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				idx.Search("go")
				idx.Stats()
			}
		}()
	}

	wg.Wait()
}

func TestAllDocumentsDeepCopy(t *testing.T) {
	idx := New()
	idx.Add(newTestDoc("https://example.com", "Test", map[string]int{"go": 5}))

	docs := idx.AllDocuments()
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(docs))
	}

	// Modify the copy — should not affect the index
	docs[0].WordFreq["go"] = 999
	docs[0].Title = "Modified"

	results := idx.Search("go")
	if results[0].Score == 999 {
		t.Error("modifying copy should not affect index")
	}
}

func TestSearchReturnsMaxDepthInTriple(t *testing.T) {
	idx := New()
	idx.Add(&Document{
		URL: "https://example.com", OriginURL: "https://seed.com",
		Depth: 1, MaxDepth: 5, Title: "Test",
		WordFreq: map[string]int{"test": 1},
	})

	results := idx.Search("test")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Depth != 5 {
		t.Errorf("expected Depth=5 (MaxDepth/k parameter), got %d", results[0].Depth)
	}
}

func TestLoadDocuments(t *testing.T) {
	idx := New()
	docs := []Document{
		{URL: "https://a.com", OriginURL: "https://seed.com", Depth: 0, MaxDepth: 2, Title: "A", WordFreq: map[string]int{"hello": 1}},
		{URL: "https://b.com", OriginURL: "https://seed.com", Depth: 1, MaxDepth: 2, Title: "B", WordFreq: map[string]int{"world": 1}},
	}
	idx.LoadDocuments(docs)

	docCount, _ := idx.Stats()
	if docCount != 2 {
		t.Errorf("expected 2 docs after load, got %d", docCount)
	}

	results := idx.Search("hello")
	if len(results) != 1 || results[0].RelevantURL != "https://a.com" {
		t.Errorf("search after load failed: %v", results)
	}
}
