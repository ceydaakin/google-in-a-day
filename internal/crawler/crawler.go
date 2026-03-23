package crawler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ceydaakin/google-in-a-day/internal/index"
)

// Config holds crawler configuration.
type Config struct {
	MaxDepth   int
	Workers    int
	QueueSize  int
	Timeout    time.Duration
	SameDomain bool
	RateLimit  time.Duration
}

// Crawler is the concurrent web crawler with lifecycle controls.
type Crawler struct {
	config      Config
	idx         *index.Index
	visited     sync.Map
	client      *http.Client
	rateLimiter *RateLimiter
	metrics     *Metrics

	state    atomicState
	active   int64 // atomic: in-flight tasks
	frontier chan CrawlTask
	stopCh   chan struct{}
	doneCh   chan struct{}
	seedURL  string

	mu sync.Mutex // protects pause/resume coordination
}

// New creates a new Crawler.
func New(cfg Config, idx *index.Index) *Crawler {
	rateInterval := cfg.RateLimit
	if rateInterval == 0 {
		rateInterval = 500 * time.Millisecond
	}

	return &Crawler{
		config: cfg,
		idx:    idx,
		client: &http.Client{
			Timeout: cfg.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		rateLimiter: NewRateLimiter(rateInterval),
		metrics:     NewMetrics(cfg.Workers),
	}
}

// Start begins crawling from the seed URL. Non-blocking — returns immediately.
func (c *Crawler) Start(seedURL string) error {
	if c.state.Load() == StateRunning || c.state.Load() == StatePaused {
		return fmt.Errorf("crawler is already active (state: %s)", c.state.Load())
	}

	c.seedURL = seedURL
	c.visited = sync.Map{} // Reset visited set for each new crawl
	c.frontier = make(chan CrawlTask, c.config.QueueSize)
	c.stopCh = make(chan struct{})
	c.doneCh = make(chan struct{})
	c.metrics = NewMetrics(c.config.Workers)
	c.state.Store(StateRunning)

	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < c.config.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.worker(id)
		}(i)
	}

	// Enqueue seed
	c.visited.Store(seedURL, true)
	atomic.AddInt64(&c.active, 1)
	c.metrics.IncrQueued()
	c.frontier <- CrawlTask{URL: seedURL, OriginURL: seedURL, Depth: 0}

	// Monitor: wait for all work to be done or stop to be requested
	go func() {
		defer close(c.doneCh)
		for {
			select {
			case <-c.stopCh:
				// Wait for all workers to exit (they check stopCh)
				wg.Wait()
				c.state.Store(StateStopped)
				return
			default:
				if atomic.LoadInt64(&c.active) == 0 && c.state.Load() == StateRunning {
					close(c.frontier)
					wg.Wait()
					docs, keywords := c.idx.Stats()
					log.Printf("Crawling complete. Indexed %d pages, %d keywords.", docs, keywords)
					c.state.Store(StateCompleted)
					return
				}
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	log.Printf("Crawler started: seed=%s depth=%d workers=%d queue=%d rate=%s",
		seedURL, c.config.MaxDepth, c.config.Workers, c.config.QueueSize, c.config.RateLimit)

	return nil
}

// Pause pauses the crawler. Workers finish their current task then wait.
func (c *Crawler) Pause() error {
	if !c.state.CompareAndSwap(StateRunning, StatePaused) {
		return fmt.Errorf("can only pause a running crawler (state: %s)", c.state.Load())
	}
	log.Println("Crawler paused")
	return nil
}

// Resume resumes a paused crawler.
func (c *Crawler) Resume() error {
	if !c.state.CompareAndSwap(StatePaused, StateRunning) {
		return fmt.Errorf("can only resume a paused crawler (state: %s)", c.state.Load())
	}
	log.Println("Crawler resumed")
	return nil
}

// Stop signals all workers to terminate and waits for completion.
func (c *Crawler) Stop() error {
	current := c.state.Load()
	if current != StateRunning && current != StatePaused {
		return fmt.Errorf("crawler is not active (state: %s)", current)
	}
	c.state.Store(StateStopped)
	close(c.stopCh)
	log.Println("Crawler stop requested")
	return nil
}

// Wait blocks until the crawler finishes or is stopped.
func (c *Crawler) Wait() {
	if c.doneCh != nil {
		<-c.doneCh
	}
}

// State returns the current crawler state.
func (c *Crawler) State() CrawlState {
	return c.state.Load()
}

// GetMetrics returns an immutable snapshot of current metrics.
func (c *Crawler) GetMetrics() MetricsSnapshot {
	return c.metrics.Snapshot()
}

// SeedURL returns the current seed URL.
func (c *Crawler) SeedURL() string {
	return c.seedURL
}

// SetMaxDepth updates the max crawl depth (only effective before Start).
func (c *Crawler) SetMaxDepth(depth int) {
	c.config.MaxDepth = depth
}

// Snapshot creates a persistence snapshot. Only call when paused or stopped.
func (c *Crawler) Snapshot() CrawlSnapshot {
	var visitedURLs []string
	c.visited.Range(func(key, value interface{}) bool {
		visitedURLs = append(visitedURLs, key.(string))
		return true
	})

	return CrawlSnapshot{
		SeedURL:     c.seedURL,
		MaxDepth:    c.config.MaxDepth,
		Documents:   c.idx.AllDocuments(),
		VisitedURLs: visitedURLs,
		Timestamp:   time.Now(),
	}
}

// RestoreFrom loads a previously saved snapshot into the crawler and index.
func (c *Crawler) RestoreFrom(snapshot CrawlSnapshot) {
	c.idx.LoadDocuments(snapshot.Documents)
	for _, u := range snapshot.VisitedURLs {
		c.visited.Store(u, true)
	}
	c.seedURL = snapshot.SeedURL
	log.Printf("Restored %d documents, %d visited URLs from snapshot",
		len(snapshot.Documents), len(snapshot.VisitedURLs))
}

func (c *Crawler) worker(id int) {
	for {
		// Check stop first
		select {
		case <-c.stopCh:
			return
		default:
		}

		// Wait if paused
		c.waitIfPaused(id)

		// Get next task from frontier (or stop)
		select {
		case <-c.stopCh:
			return
		case task, ok := <-c.frontier:
			if !ok {
				return // frontier closed, crawl complete
			}
			c.metrics.IncrActiveWorkers()
			c.processTask(id, task)
			c.metrics.DecrActiveWorkers()
			atomic.AddInt64(&c.active, -1)
		}
	}
}

func (c *Crawler) waitIfPaused(workerID int) {
	for c.state.Load() == StatePaused {
		c.metrics.SetWorkerState(workerID, "paused", "")
		select {
		case <-c.stopCh:
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *Crawler) processTask(id int, task CrawlTask) {
	c.metrics.SetWorkerState(id, "fetching", task.URL)
	start := time.Now()

	parsedURL, err := url.Parse(task.URL)
	if err != nil {
		c.metrics.IncrErrored()
		c.metrics.SetWorkerState(id, "idle", "")
		return
	}

	// Rate limit per domain
	if !c.rateLimiter.Wait(parsedURL.Host, c.stopCh) {
		return // stop requested during rate limit wait
	}

	resp, err := c.client.Get(task.URL)
	if err != nil {
		c.metrics.IncrErrored()
		c.metrics.RecordHistory(HistoryEntry{
			URL: task.URL, Duration: time.Since(start),
			Timestamp: time.Now(), Error: err.Error(),
		})
		c.metrics.SetWorkerState(id, "idle", "")
		log.Printf("[worker %d] error fetching %s: %v", id, task.URL, err)
		return
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		c.metrics.IncrErrored()
		c.metrics.RecordHistory(HistoryEntry{
			URL: task.URL, StatusCode: statusCode,
			Duration: time.Since(start), Timestamp: time.Now(),
		})
		c.metrics.SetWorkerState(id, "idle", "")
		log.Printf("[worker %d] non-200 status %d for %s", id, statusCode, task.URL)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		c.metrics.SetWorkerState(id, "idle", "")
		return
	}

	// Parse page
	c.metrics.SetWorkerState(id, "parsing", task.URL)
	title, bodyText, links := parsePage(resp.Body, parsedURL)
	wordFreq := wordFrequency(bodyText)

	doc := &index.Document{
		URL:       task.URL,
		OriginURL: task.OriginURL,
		Depth:     task.Depth,
		MaxDepth:  c.config.MaxDepth,
		Title:     title,
		WordFreq:  wordFreq,
	}
	c.idx.Add(doc)
	c.metrics.IncrProcessed()
	c.metrics.RecordHistory(HistoryEntry{
		URL: task.URL, StatusCode: statusCode,
		Duration: time.Since(start), Timestamp: time.Now(),
	})

	log.Printf("[worker %d] indexed (depth=%d): %s", id, task.Depth, task.URL)

	// Enqueue child links
	if task.Depth >= c.config.MaxDepth {
		c.metrics.SetWorkerState(id, "idle", "")
		return
	}

	c.enqueueLinks(id, task, links)
	c.metrics.SetWorkerState(id, "idle", "")
}

func (c *Crawler) enqueueLinks(workerID int, parent CrawlTask, links []string) {
	seedURL, _ := url.Parse(parent.OriginURL)

	for _, link := range links {
		if c.config.SameDomain {
			parsed, err := url.Parse(link)
			if err != nil {
				continue
			}
			if parsed.Host != seedURL.Host {
				continue
			}
		}

		if _, loaded := c.visited.LoadOrStore(link, true); loaded {
			continue
		}

		// Check if stop was requested before trying to enqueue
		select {
		case <-c.stopCh:
			return
		default:
		}

		atomic.AddInt64(&c.active, 1)
		c.metrics.IncrQueued()

		newTask := CrawlTask{
			URL:       link,
			OriginURL: parent.OriginURL,
			Depth:     parent.Depth + 1,
		}

		select {
		case c.frontier <- newTask:
		case <-c.stopCh:
			atomic.AddInt64(&c.active, -1)
			return
		default:
			atomic.AddInt64(&c.active, -1)
			c.metrics.IncrDropped()
			log.Printf("[worker %d] frontier full, dropping: %s", workerID, link)
		}
	}
}
