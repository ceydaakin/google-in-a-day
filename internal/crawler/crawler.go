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
}

// Crawler is the concurrent web crawler.
type Crawler struct {
	config  Config
	idx     *index.Index
	visited sync.Map // map[string]bool
	client  *http.Client
	active  int64 // atomic counter for in-flight tasks
}

// New creates a new Crawler.
func New(cfg Config, idx *index.Index) *Crawler {
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
	}
}

// Run starts crawling from the seed URL and blocks until complete.
func (c *Crawler) Run(seedURL string) {
	frontier := make(chan CrawlTask, c.config.QueueSize)

	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < c.config.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.worker(id, frontier)
		}(i)
	}

	// Parse seed URL for domain restriction
	parsedSeed, err := url.Parse(seedURL)
	if err != nil {
		log.Fatalf("invalid seed URL: %v", err)
	}
	_ = parsedSeed

	// Enqueue seed
	c.visited.Store(seedURL, true)
	atomic.AddInt64(&c.active, 1)
	frontier <- CrawlTask{URL: seedURL, OriginURL: seedURL, Depth: 0}

	// Monitor: close frontier when all work is done
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			if atomic.LoadInt64(&c.active) == 0 {
				close(frontier)
				return
			}
		}
	}()

	wg.Wait()
	docs, keywords := c.idx.Stats()
	log.Printf("Crawling complete. Indexed %d pages, %d keywords.", docs, keywords)
}

func (c *Crawler) worker(id int, frontier chan CrawlTask) {
	for task := range frontier {
		c.processTask(id, task, frontier)
		atomic.AddInt64(&c.active, -1)
	}
}

func (c *Crawler) processTask(id int, task CrawlTask, frontier chan CrawlTask) {
	log.Printf("[worker %d] crawling (depth=%d): %s", id, task.Depth, task.URL)

	resp, err := c.client.Get(task.URL)
	if err != nil {
		log.Printf("[worker %d] error fetching %s: %v", id, task.URL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[worker %d] non-200 status %d for %s", id, resp.StatusCode, task.URL)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return
	}

	pageURL, err := url.Parse(task.URL)
	if err != nil {
		return
	}

	title, bodyText, links := parsePage(resp.Body, pageURL)
	wordFreq := wordFrequency(bodyText)

	// Index the document
	doc := &index.Document{
		URL:       task.URL,
		OriginURL: task.OriginURL,
		Depth:     task.Depth,
		Title:     title,
		Body:      bodyText,
		WordFreq:  wordFreq,
	}
	c.idx.Add(doc)

	// Enqueue child links if not at max depth
	if task.Depth >= c.config.MaxDepth {
		return
	}

	seedURL, _ := url.Parse(task.OriginURL)
	for _, link := range links {
		// Same-domain check
		if c.config.SameDomain {
			parsed, err := url.Parse(link)
			if err != nil {
				continue
			}
			if parsed.Host != seedURL.Host {
				continue
			}
		}

		// Check visited (atomic check-and-set)
		if _, loaded := c.visited.LoadOrStore(link, true); loaded {
			continue
		}

		atomic.AddInt64(&c.active, 1)
		// Non-blocking send: drop URLs if frontier is full (back pressure)
		select {
		case frontier <- CrawlTask{URL: link, OriginURL: task.OriginURL, Depth: task.Depth + 1}:
		default:
			// Frontier full — drop this URL (back pressure)
			atomic.AddInt64(&c.active, -1)
			log.Printf("[worker %d] frontier full, dropping: %s", id, link)
		}
	}
}
