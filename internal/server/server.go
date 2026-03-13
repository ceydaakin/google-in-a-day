package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/ceydaakin/google-in-a-day/internal/index"
)

// Server serves the search HTTP interface.
type Server struct {
	idx  *index.Index
	addr string
}

// New creates a new search server.
func New(idx *index.Index, addr string) *Server {
	return &Server{idx: idx, addr: addr}
}

// Start begins listening. Call this in a goroutine.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/search", s.handleSearch)
	mux.HandleFunc("/api/search", s.handleAPISearch)
	mux.HandleFunc("/api/stats", s.handleStats)

	log.Printf("Search server listening on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
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

	// Format as triples
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

const homePage = `<!DOCTYPE html>
<html>
<head><title>Google in a Day</title>
<style>
  body { font-family: sans-serif; max-width: 700px; margin: 80px auto; text-align: center; }
  h1 { font-size: 2.5em; margin-bottom: 30px; }
  input[type=text] { width: 400px; padding: 12px; font-size: 16px; border: 1px solid #ddd; border-radius: 24px; outline: none; }
  input[type=text]:focus { border-color: #4285f4; }
  button { padding: 12px 24px; margin: 20px 8px; font-size: 14px; border: 1px solid #ddd; border-radius: 4px; background: #f8f8f8; cursor: pointer; }
  button:hover { border-color: #999; }
</style>
</head>
<body>
  <h1>Google in a Day</h1>
  <form action="/search" method="GET">
    <input type="text" name="q" placeholder="Search the crawled web..." autofocus>
    <br>
    <button type="submit">Search</button>
  </form>
</body>
</html>`

const resultsPage = `<!DOCTYPE html>
<html>
<head><title>{{.Query}} - Search</title>
<style>
  body { font-family: sans-serif; max-width: 700px; margin: 20px auto; }
  .header { margin-bottom: 20px; }
  .header a { text-decoration: none; font-size: 1.5em; font-weight: bold; }
  .header form { display: inline; margin-left: 20px; }
  .header input[type=text] { width: 350px; padding: 8px; font-size: 14px; border: 1px solid #ddd; border-radius: 20px; }
  .stats { color: #666; font-size: 13px; margin-bottom: 20px; }
  .result { margin-bottom: 20px; }
  .result .url { color: #006621; font-size: 13px; }
  .result .title { font-size: 18px; }
  .result .title a { color: #1a0dab; text-decoration: none; }
  .result .title a:hover { text-decoration: underline; }
  .result .meta { color: #666; font-size: 12px; }
  .no-results { color: #666; font-size: 16px; }
</style>
</head>
<body>
  <div class="header">
    <a href="/">Google in a Day</a>
    <form action="/search" method="GET">
      <input type="text" name="q" value="{{.Query}}">
      <button type="submit">Search</button>
    </form>
  </div>
  <div class="stats">{{.Count}} results found ({{.Docs}} pages indexed, {{.Keywords}} keywords)</div>
  {{if .Results}}
    {{range .Results}}
    <div class="result">
      <div class="url">{{.RelevantURL}}</div>
      <div class="title"><a href="{{.RelevantURL}}">{{if .Title}}{{.Title}}{{else}}{{.RelevantURL}}{{end}}</a></div>
      <div class="meta">Origin: {{.OriginURL}} | Depth: {{.Depth}} | Score: {{printf "%.1f" .Score}}</div>
    </div>
    {{end}}
  {{else}}
    <div class="no-results">No results found for "{{.Query}}"</div>
  {{end}}
</body>
</html>`
