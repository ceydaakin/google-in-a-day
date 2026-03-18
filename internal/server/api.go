package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ceydaakin/google-in-a-day/internal/crawler"
)

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, resp apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardPage)
}

func (s *Server) handleStartCrawl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "method not allowed"})
		return
	}

	var req struct {
		URL       string `json:"url"`
		Depth     int    `json:"depth"`
		Workers   int    `json:"workers"`
		QueueSize int    `json:"queue_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "invalid JSON body"})
		return
	}

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "url is required"})
		return
	}

	if _, err := url.Parse(req.URL); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Error: "invalid URL"})
		return
	}

	// Apply overrides if provided; otherwise use server defaults
	if req.Depth > 0 {
		s.crawler.SetMaxDepth(req.Depth)
	}

	if err := s.crawler.Start(req.URL); err != nil {
		writeJSON(w, http.StatusConflict, apiResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Data:    map[string]string{"status": "started", "seed_url": req.URL},
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	metrics := s.crawler.GetMetrics()
	state := s.crawler.State()
	docs, keywords := s.idx.Stats()

	resp := map[string]interface{}{
		"state":    state.String(),
		"seed_url": s.crawler.SeedURL(),
		"metrics":  metrics,
		"docs":     docs,
		"keywords": keywords,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handlePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "method not allowed"})
		return
	}
	if err := s.crawler.Pause(); err != nil {
		writeJSON(w, http.StatusConflict, apiResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: "paused"})
}

func (s *Server) handleResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "method not allowed"})
		return
	}
	if err := s.crawler.Resume(); err != nil {
		writeJSON(w, http.StatusConflict, apiResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: "resumed"})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "method not allowed"})
		return
	}
	if err := s.crawler.Stop(); err != nil {
		writeJSON(w, http.StatusConflict, apiResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: "stopped"})
}

func (s *Server) handleSaveState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Error: "method not allowed"})
		return
	}

	state := s.crawler.State()
	if state != crawler.StatePaused && state != crawler.StateStopped && state != crawler.StateCompleted {
		writeJSON(w, http.StatusConflict, apiResponse{
			Error: "crawler must be paused, stopped, or completed to save state",
		})
		return
	}

	snapshot := s.crawler.Snapshot()
	path := fmt.Sprintf("crawl_state_%d.json", time.Now().Unix())
	if err := crawler.SaveState(path, snapshot); err != nil {
		writeJSON(w, http.StatusInternalServerError, apiResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Data:    map[string]string{"path": path},
	})
}
