package server

import (
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
		URL: "https://example.com", Title: "Test",
		WordFreq: map[string]int{"test": 1},
	})

	req := httptest.NewRequest("GET", "/search?q=test", nil)
	w := httptest.NewRecorder()
	srv.handleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "1 results found") {
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
		Depth: 1, Title: "Go Tutorial",
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
		URL: "https://a.com", WordFreq: map[string]int{"go": 1},
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
	body := strings.NewReader(`{"url":""}`)
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

	body := strings.NewReader(`{"url":"` + ts.URL + `","depth":0}`)
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
