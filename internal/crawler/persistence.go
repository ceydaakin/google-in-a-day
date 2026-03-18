package crawler

import (
	"encoding/json"
	"os"
	"time"

	"github.com/ceydaakin/google-in-a-day/internal/index"
)

// CrawlSnapshot captures the full state of a crawl for persistence.
type CrawlSnapshot struct {
	SeedURL     string           `json:"seed_url"`
	MaxDepth    int              `json:"max_depth"`
	Documents   []index.Document `json:"documents"`
	VisitedURLs []string         `json:"visited_urls"`
	QueuedTasks []CrawlTask      `json:"queued_tasks"`
	Timestamp   time.Time        `json:"timestamp"`
}

// SaveState writes a crawl snapshot to disk atomically.
// It writes to a temp file first, then renames to prevent corruption.
func SaveState(path string, snapshot CrawlSnapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

// LoadState reads a crawl snapshot from disk.
func LoadState(path string) (CrawlSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CrawlSnapshot{}, err
	}

	var snapshot CrawlSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return CrawlSnapshot{}, err
	}

	return snapshot, nil
}
