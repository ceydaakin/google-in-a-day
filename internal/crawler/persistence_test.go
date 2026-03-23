package crawler

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ceydaakin/google-in-a-day/internal/index"
)

func TestSaveAndLoadState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	snapshot := CrawlSnapshot{
		SeedURL:  "https://example.com",
		MaxDepth: 2,
		Documents: []index.Document{
			{
				URL:       "https://example.com",
				OriginURL: "https://example.com",
				Depth:     0,
				MaxDepth:  2,
				Title:     "Test",
				Body:      "Hello world",
				WordFreq:  map[string]int{"hello": 1, "world": 1},
			},
		},
		VisitedURLs: []string{"https://example.com", "https://example.com/about"},
		QueuedTasks: []CrawlTask{
			{URL: "https://example.com/docs", OriginURL: "https://example.com", Depth: 1},
		},
		Timestamp: time.Now(),
	}

	if err := SaveState(path, snapshot); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	loaded, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.SeedURL != snapshot.SeedURL {
		t.Errorf("SeedURL: got %s, want %s", loaded.SeedURL, snapshot.SeedURL)
	}
	if loaded.MaxDepth != snapshot.MaxDepth {
		t.Errorf("MaxDepth: got %d, want %d", loaded.MaxDepth, snapshot.MaxDepth)
	}
	if len(loaded.Documents) != 1 {
		t.Errorf("Documents: got %d, want 1", len(loaded.Documents))
	}
	if len(loaded.VisitedURLs) != 2 {
		t.Errorf("VisitedURLs: got %d, want 2", len(loaded.VisitedURLs))
	}
	if len(loaded.QueuedTasks) != 1 {
		t.Errorf("QueuedTasks: got %d, want 1", len(loaded.QueuedTasks))
	}
}

func TestLoadStateNonexistent(t *testing.T) {
	_, err := LoadState("/nonexistent/path/state.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadStateInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, err := LoadState(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSaveStateAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	snapshot := CrawlSnapshot{SeedURL: "https://example.com", Timestamp: time.Now()}
	if err := SaveState(path, snapshot); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// Verify temp file is cleaned up
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should be cleaned up after rename")
	}

	// Verify final file exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("final file should exist: %v", err)
	}
}
