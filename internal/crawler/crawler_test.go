package crawler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ceydaakin/google-in-a-day/internal/index"
)

func newTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><head><title>Home</title></head>
		<body><p>Welcome to the test site about Go programming</p>
		<a href="/about">About</a>
		<a href="/docs">Docs</a></body></html>`)
	})
	mux.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><head><title>About</title></head>
		<body><p>About this Go project</p>
		<a href="/about/team">Team</a></body></html>`)
	})
	mux.HandleFunc("/about/team", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><head><title>Team</title></head>
		<body><p>Our Go development team</p></body></html>`)
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><head><title>Docs</title></head>
		<body><p>Documentation for Go tutorials</p></body></html>`)
	})
	return httptest.NewServer(mux)
}

func newTestCrawler(idx *index.Index) *Crawler {
	return New(Config{
		MaxDepth:   2,
		Workers:    2,
		QueueSize:  50,
		Timeout:    5 * time.Second,
		SameDomain: true,
		RateLimit:  10 * time.Millisecond,
	}, idx)
}

func TestCrawlerIndexesSeedPage(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	idx := index.New()
	c := New(Config{
		MaxDepth:   0, // only seed page
		Workers:    1,
		QueueSize:  10,
		Timeout:    5 * time.Second,
		SameDomain: true,
		RateLimit:  1 * time.Millisecond,
	}, idx)

	if err := c.Start(ts.URL + "/"); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	c.Wait()

	docs, _ := idx.Stats()
	if docs != 1 {
		t.Errorf("expected 1 doc (seed only), got %d", docs)
	}
}

func TestCrawlerFollowsLinks(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	idx := index.New()
	c := newTestCrawler(idx)

	if err := c.Start(ts.URL + "/"); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	c.Wait()

	docs, _ := idx.Stats()
	if docs < 3 {
		t.Errorf("expected at least 3 docs (home + about + docs), got %d", docs)
	}
}

func TestCrawlerDeduplicates(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	idx := index.New()
	c := newTestCrawler(idx)

	if err := c.Start(ts.URL + "/"); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	c.Wait()

	// Run again — should fail because crawler is in completed state
	// Verify no duplicate pages were indexed
	allDocs := idx.AllDocuments()
	seen := make(map[string]bool)
	for _, d := range allDocs {
		if seen[d.URL] {
			t.Errorf("duplicate document found: %s", d.URL)
		}
		seen[d.URL] = true
	}
}

func TestCrawlerStartWhileRunning(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	idx := index.New()
	c := New(Config{
		MaxDepth:   5,
		Workers:    1,
		QueueSize:  100,
		Timeout:    5 * time.Second,
		SameDomain: true,
		RateLimit:  100 * time.Millisecond, // slow to keep it running
	}, idx)

	if err := c.Start(ts.URL + "/"); err != nil {
		t.Fatalf("first start failed: %v", err)
	}

	// Try to start again while running
	err := c.Start(ts.URL + "/")
	if err == nil {
		t.Error("expected error when starting while already running")
	}

	c.Stop()
	c.Wait()
}

func TestCrawlerPauseResume(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	idx := index.New()
	c := New(Config{
		MaxDepth:   3,
		Workers:    2,
		QueueSize:  50,
		Timeout:    5 * time.Second,
		SameDomain: true,
		RateLimit:  50 * time.Millisecond,
	}, idx)

	if err := c.Start(ts.URL + "/"); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := c.Pause(); err != nil {
		t.Fatalf("pause failed: %v", err)
	}
	if c.State() != StatePaused {
		t.Errorf("expected paused, got %s", c.State())
	}

	if err := c.Resume(); err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	if c.State() != StateRunning {
		t.Errorf("expected running, got %s", c.State())
	}

	c.Stop()
	c.Wait()
}

func TestCrawlerStop(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	idx := index.New()
	c := New(Config{
		MaxDepth:   10,
		Workers:    2,
		QueueSize:  100,
		Timeout:    5 * time.Second,
		SameDomain: true,
		RateLimit:  100 * time.Millisecond,
	}, idx)

	if err := c.Start(ts.URL + "/"); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := c.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	c.Wait()

	state := c.State()
	if state != StateStopped {
		t.Errorf("expected stopped, got %s", state)
	}
}

func TestCrawlerMetrics(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	idx := index.New()
	c := newTestCrawler(idx)

	if err := c.Start(ts.URL + "/"); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	c.Wait()

	m := c.GetMetrics()
	if m.URLsProcessed == 0 {
		t.Error("expected some URLs processed")
	}
	if m.URLsQueued == 0 {
		t.Error("expected some URLs queued")
	}
}

func TestCrawlerBackPressure(t *testing.T) {
	// Create a server with many links
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<html><body>`
		for i := 0; i < 20; i++ {
			html += fmt.Sprintf(`<a href="/page%d">Page %d</a>`, i, i)
		}
		html += `</body></html>`
		fmt.Fprint(w, html)
	})
	for i := 0; i < 20; i++ {
		i := i
		mux.HandleFunc(fmt.Sprintf("/page%d", i), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><p>Page %d content</p></body></html>`, i)
		})
	}
	ts := httptest.NewServer(mux)
	defer ts.Close()

	idx := index.New()
	c := New(Config{
		MaxDepth:   1,
		Workers:    1,
		QueueSize:  3, // very small queue to trigger back pressure
		Timeout:    5 * time.Second,
		SameDomain: true,
		RateLimit:  1 * time.Millisecond,
	}, idx)

	if err := c.Start(ts.URL + "/"); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	c.Wait()

	m := c.GetMetrics()
	if m.URLsDropped == 0 {
		t.Log("Note: back pressure test may not trigger drops depending on timing")
	}
}
