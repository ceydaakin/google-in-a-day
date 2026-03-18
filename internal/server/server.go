package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/ceydaakin/google-in-a-day/internal/crawler"
	"github.com/ceydaakin/google-in-a-day/internal/index"
)

// Server serves the search HTTP interface and dashboard.
type Server struct {
	crawler *crawler.Crawler
	idx     *index.Index
	addr    string
	httpSrv *http.Server
}

// New creates a new search server.
func New(c *crawler.Crawler, idx *index.Index, addr string) *Server {
	return &Server{crawler: c, idx: idx, addr: addr}
}

// Start begins listening. Call this in a goroutine.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Search UI
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/search", s.handleSearch)
	mux.HandleFunc("/dashboard", s.handleDashboard)

	// Search API
	mux.HandleFunc("/api/search", s.handleAPISearch)
	mux.HandleFunc("/api/stats", s.handleStats)

	// Crawler lifecycle API
	mux.HandleFunc("/api/index", s.handleStartCrawl)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/pause", s.handlePause)
	mux.HandleFunc("/api/resume", s.handleResume)
	mux.HandleFunc("/api/stop", s.handleStop)
	mux.HandleFunc("/api/save", s.handleSaveState)

	s.httpSrv = &http.Server{Addr: s.addr, Handler: mux}
	log.Printf("Search server listening on %s", s.addr)
	log.Printf("Dashboard: http://localhost%s/dashboard", s.addr)
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.New("home").Parse(homePage))
	tmpl.Execute(w, nil)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	results := s.idx.Search(query)
	docs, keywords := s.idx.Stats()

	data := struct {
		Query    string
		Results  []index.SearchResult
		Count    int
		Docs     int
		Keywords int
	}{
		Query:    query,
		Results:  results,
		Count:    len(results),
		Docs:     docs,
		Keywords: keywords,
	}

	tmpl := template.Must(template.New("results").Parse(resultsPage))
	tmpl.Execute(w, data)
}

func (s *Server) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"missing query parameter 'q'"}`, http.StatusBadRequest)
		return
	}

	results := s.idx.Search(query)

	type Triple struct {
		RelevantURL string  `json:"relevant_url"`
		OriginURL   string  `json:"origin_url"`
		Depth       int     `json:"depth"`
		Score       float64 `json:"score"`
		Title       string  `json:"title"`
	}

	triples := make([]Triple, len(results))
	for i, r := range results {
		triples[i] = Triple{
			RelevantURL: r.RelevantURL,
			OriginURL:   r.OriginURL,
			Depth:       r.Depth,
			Score:       r.Score,
			Title:       r.Title,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triples)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	docs, keywords := s.idx.Stats()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"documents":%d,"keywords":%d}`, docs, keywords)
}
