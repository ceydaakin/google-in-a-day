package crawler

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects crawler performance data in a thread-safe manner.
type Metrics struct {
	urlsProcessed int64
	urlsQueued    int64
	urlsDropped   int64
	urlsErrored   int64
	activeWorkers int64

	mu            sync.RWMutex
	workerStatus  map[int]WorkerInfo
	history       []HistoryEntry
	maxHistory    int
	startTime     time.Time
	backPressures int64 // number of times back pressure was applied
}

// WorkerInfo describes what a worker is currently doing.
type WorkerInfo struct {
	ID    int    `json:"id"`
	State string `json:"state"` // "idle", "fetching", "parsing", "paused"
	URL   string `json:"url"`
}

// HistoryEntry records a single crawl event.
type HistoryEntry struct {
	URL        string        `json:"url"`
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"duration_ms"`
	Timestamp  time.Time     `json:"timestamp"`
	Error      string        `json:"error,omitempty"`
}

// MetricsSnapshot is an immutable point-in-time copy of all metrics.
type MetricsSnapshot struct {
	URLsProcessed int64         `json:"urls_processed"`
	URLsQueued    int64         `json:"urls_queued"`
	URLsDropped   int64         `json:"urls_dropped"`
	URLsErrored   int64         `json:"urls_errored"`
	ActiveWorkers int64         `json:"active_workers"`
	BackPressures int64         `json:"back_pressures"`
	Workers       []WorkerInfo  `json:"workers"`
	History       []HistoryEntry `json:"history"`
	Uptime        string        `json:"uptime"`
}

// NewMetrics creates a new metrics collector.
func NewMetrics(workerCount int) *Metrics {
	workers := make(map[int]WorkerInfo, workerCount)
	for i := 0; i < workerCount; i++ {
		workers[i] = WorkerInfo{ID: i, State: "idle"}
	}
	return &Metrics{
		workerStatus: workers,
		maxHistory:   100,
		startTime:    time.Now(),
	}
}

func (m *Metrics) IncrProcessed()   { atomic.AddInt64(&m.urlsProcessed, 1) }
func (m *Metrics) IncrQueued()      { atomic.AddInt64(&m.urlsQueued, 1) }
func (m *Metrics) IncrDropped()     { atomic.AddInt64(&m.urlsDropped, 1); atomic.AddInt64(&m.backPressures, 1) }
func (m *Metrics) IncrErrored()     { atomic.AddInt64(&m.urlsErrored, 1) }
func (m *Metrics) IncrActiveWorkers() { atomic.AddInt64(&m.activeWorkers, 1) }
func (m *Metrics) DecrActiveWorkers() { atomic.AddInt64(&m.activeWorkers, -1) }

// SetWorkerState updates a worker's current activity.
func (m *Metrics) SetWorkerState(id int, state, url string) {
	m.mu.Lock()
	m.workerStatus[id] = WorkerInfo{ID: id, State: state, URL: url}
	m.mu.Unlock()
}

// RecordHistory appends a crawl event to the bounded history ring.
func (m *Metrics) RecordHistory(entry HistoryEntry) {
	m.mu.Lock()
	if len(m.history) >= m.maxHistory {
		m.history = m.history[1:]
	}
	m.history = append(m.history, entry)
	m.mu.Unlock()
}

// Snapshot returns an immutable copy of all metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	workers := make([]WorkerInfo, 0, len(m.workerStatus))
	for _, w := range m.workerStatus {
		workers = append(workers, w)
	}
	history := make([]HistoryEntry, len(m.history))
	copy(history, m.history)
	m.mu.RUnlock()

	return MetricsSnapshot{
		URLsProcessed: atomic.LoadInt64(&m.urlsProcessed),
		URLsQueued:    atomic.LoadInt64(&m.urlsQueued),
		URLsDropped:   atomic.LoadInt64(&m.urlsDropped),
		URLsErrored:   atomic.LoadInt64(&m.urlsErrored),
		ActiveWorkers: atomic.LoadInt64(&m.activeWorkers),
		BackPressures: atomic.LoadInt64(&m.backPressures),
		Workers:       workers,
		History:       history,
		Uptime:        time.Since(m.startTime).Round(time.Second).String(),
	}
}
