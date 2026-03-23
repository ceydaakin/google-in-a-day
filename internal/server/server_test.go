package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ceydaakin/google-in-a-day/internal/crawler"
	"github.com/ceydaakin/google-in-a-day/internal/index"
)

func newTestServer() *Server {
	idx := index.New()
	c := crawler.New(crawler.Config{
		MaxDepth:   2,
		Workers:    2,
		QueueSize:  10,
		Timeout:    5 * time.Second,
		SameDomain: true,
		RateLimit:  10 * time.Millisecond,
	}, idx)
	return New(c, idx, ":0")
}

func TestHomeHandler(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.handleHome(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Google in a Day") {
		t.Error("expected home page content")
	}
}

func TestHomeHandler404(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.handleHome(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestSearchRedirect(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()
	srv.handleSearch(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", w.Code)
	}
}

func TestSearchWithQuery(t *testing.T) {
	srv := newTestServer()
	srv.idx.Add(&index.Document{
		URL: "https://example.com", Title: "Test", MaxDepth: 2,
		WordFreq: map[string]int{"test": 1},
	})

	req := httptest.NewRequest("GET", "/search?q=test", nil)
	w := httptest.NewRecorder()
	srv.handleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "1 results") {
		t.Error("expected results in page")
	}
}

func TestAPISearchMissingQuery(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/api/search", nil)
	w := httptest.NewRecorder()
	srv.handleAPISearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAPISearchJSON(t *testing.T) {
	srv := newTestServer()
	srv.idx.Add(&index.Document{
		URL: "https://example.com", OriginURL: "https://seed.com",
		Depth: 1, MaxDepth: 2, Title: "Go Tutorial",
		WordFreq: map[string]int{"go": 5},
	})

	req := httptest.NewRequest("GET", "/api/search?q=go", nil)
	w := httptest.NewRecorder()
	srv.handleAPISearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var results []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &results)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["relevant_url"] != "https://example.com" {
		t.Errorf("unexpected URL: %v", results[0]["relevant_url"])
	}
}

func TestAPIStats(t *testing.T) {
	srv := newTestServer()
	srv.idx.Add(&index.Document{
		URL: "https://a.com", MaxDepth: 2, WordFreq: map[string]int{"go": 1},
	})

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()
	srv.handleStats(w, req)

	var stats map[string]int
	json.Unmarshal(w.Body.Bytes(), &stats)
	if stats["documents"] != 1 {
		t.Errorf("expected 1 document, got %d", stats["documents"])
	}
}

func TestAPIStatus(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	srv.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var status map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &status)
	if status["state"] != "idle" {
		t.Errorf("expected idle state, got %v", status["state"])
	}
}

func TestAPIPauseNotRunning(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("POST", "/api/pause", nil)
	w := httptest.NewRecorder()
	srv.handlePause(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestAPIStopNotRunning(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("POST", "/api/stop", nil)
	w := httptest.NewRecorder()
	srv.handleStop(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestAPIStartMissingURL(t *testing.T) {
	srv := newTestServer()
	body := strings.NewReader(`{"origin":""}`)
	req := httptest.NewRequest("POST", "/api/index", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleStartCrawl(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAPIDashboard(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()
	srv.handleDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Dashboard") {
		t.Error("expected dashboard content")
	}
}

func TestAPIResumeNotPaused(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("POST", "/api/resume", nil)
	w := httptest.NewRecorder()
	srv.handleResume(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestAPISaveStateNotReady(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("POST", "/api/save", nil)
	w := httptest.NewRecorder()
	srv.handleSaveState(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestAPISaveStateMethodNotAllowed(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/api/save", nil)
	w := httptest.NewRecorder()
	srv.handleSaveState(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestAPIStartInvalidJSON(t *testing.T) {
	srv := newTestServer()
	body := strings.NewReader("not json")
	req := httptest.NewRequest("POST", "/api/index", body)
	w := httptest.NewRecorder()
	srv.handleStartCrawl(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestAPIStartValidURL(t *testing.T) {
	srv := newTestServer()
	// Create a test HTTP server that serves some content
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>test</body></html>"))
	}))
	defer ts.Close()

	body := strings.NewReader(`{"origin":"` + ts.URL + `","k":0}`)
	req := httptest.NewRequest("POST", "/api/index", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleStartCrawl(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp apiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Errorf("expected success=true, got error: %s", resp.Error)
	}

	// Wait for crawler to finish
	time.Sleep(500 * time.Millisecond)
}

func TestAPIStartInvalidURLScheme(t *testing.T) {
	srv := newTestServer()
	body := strings.NewReader(`{"origin":"not-a-url"}`)
	req := httptest.NewRequest("POST", "/api/index", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleStartCrawl(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for URL without scheme, got %d", w.Code)
	}
}

func TestAPIStartFTPScheme(t *testing.T) {
	srv := newTestServer()
	body := strings.NewReader(`{"origin":"ftp://example.com"}`)
	req := httptest.NewRequest("POST", "/api/index", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleStartCrawl(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for ftp scheme, got %d", w.Code)
	}
}

func TestAPISearchReturnsMaxDepth(t *testing.T) {
	srv := newTestServer()
	srv.idx.Add(&index.Document{
		URL: "https://example.com", OriginURL: "https://seed.com",
		Depth: 1, MaxDepth: 5, Title: "Go",
		WordFreq: map[string]int{"go": 3},
	})

	req := httptest.NewRequest("GET", "/api/search?q=go", nil)
	w := httptest.NewRecorder()
	srv.handleAPISearch(w, req)

	var results []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &results)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	depth := int(results[0]["depth"].(float64))
	if depth != 5 {
		t.Errorf("expected depth=5 (k parameter), got %d", depth)
	}
}

func TestAPISearchNoResults(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/api/search?q=nonexistent", nil)
	w := httptest.NewRecorder()
	srv.handleAPISearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSearchPageNoResults(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/search?q=nothing", nil)
	w := httptest.NewRecorder()
	srv.handleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "No results found") {
		t.Error("expected 'No results found' message")
	}
}

func TestStartAndShutdown(t *testing.T) {
	srv := newTestServer()
	srv.addr = ":0" // random port

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("shutdown error: %v", err)
	}

	err := <-errCh
	if err != nil && err.Error() != "http: Server closed" {
		t.Errorf("unexpected start error: %v", err)
	}
}

func TestAPISaveStateSuccess(t *testing.T) {
	srv := newTestServer()

	// Start and complete a crawl against a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>test page</body></html>"))
	}))
	defer ts.Close()

	// Start crawler and wait for completion (depth=0 = seed only)
	srv.crawler.SetMaxDepth(0)
	if err := srv.crawler.Start(ts.URL); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	srv.crawler.Wait()

	// Now save state (crawler should be in completed state)
	req := httptest.NewRequest("POST", "/api/save", nil)
	w := httptest.NewRecorder()
	srv.handleSaveState(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Errorf("expected success=true, got error: %s", resp.Error)
	}
}

func TestAPIStartAlreadyRunning(t *testing.T) {
	srv := newTestServer()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte("<html><body><a href=\"/page1\">link</a></body></html>"))
	}))
	defer ts.Close()

	// Start first crawl
	body1 := strings.NewReader(`{"origin":"` + ts.URL + `","k":2}`)
	req1 := httptest.NewRequest("POST", "/api/index", body1)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	srv.handleStartCrawl(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("first start expected 200, got %d", w1.Code)
	}

	// Try to start again — should get 409
	body2 := strings.NewReader(`{"origin":"` + ts.URL + `","k":1}`)
	req2 := httptest.NewRequest("POST", "/api/index", body2)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	srv.handleStartCrawl(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate start, got %d", w2.Code)
	}

	srv.crawler.Stop()
	srv.crawler.Wait()
}

func TestAPIMethodNotAllowed(t *testing.T) {
	srv := newTestServer()

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/index"},
		{"GET", "/api/pause"},
		{"GET", "/api/resume"},
		{"GET", "/api/stop"},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, nil)
		w := httptest.NewRecorder()

		switch ep.path {
		case "/api/index":
			srv.handleStartCrawl(w, req)
		case "/api/pause":
			srv.handlePause(w, req)
		case "/api/resume":
			srv.handleResume(w, req)
		case "/api/stop":
			srv.handleStop(w, req)
		}

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: expected 405, got %d", ep.method, ep.path, w.Code)
		}
	}
}
