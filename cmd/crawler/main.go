package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ceydaakin/google-in-a-day/internal/crawler"
	"github.com/ceydaakin/google-in-a-day/internal/index"
	"github.com/ceydaakin/google-in-a-day/internal/server"
)

func main() {
	seedURL := flag.String("url", "", "Seed URL to start crawling (optional — can also start via dashboard)")
	maxDepth := flag.Int("depth", 2, "Maximum crawl depth")
	workers := flag.Int("workers", 5, "Number of concurrent crawler workers")
	queueSize := flag.Int("queue", 100, "Maximum frontier queue size (back pressure)")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTP request timeout per page")
	sameDomain := flag.Bool("same-domain", true, "Only crawl pages on the same domain")
	port := flag.Int("port", 3600, "Search server port")
	rateLimit := flag.Duration("rate-limit", 500*time.Millisecond, "Minimum delay between requests to the same domain")
	saveState := flag.String("save-state", "", "Path to save crawl state on shutdown")
	loadState := flag.String("load-state", "", "Path to load previous crawl state from")

	flag.Parse()

	log.Printf("Starting Google in a Day")

	// Create shared index
	idx := index.New()

	// Create crawler
	c := crawler.New(crawler.Config{
		MaxDepth:   *maxDepth,
		Workers:    *workers,
		QueueSize:  *queueSize,
		Timeout:    *timeout,
		SameDomain: *sameDomain,
		RateLimit:  *rateLimit,
	}, idx)

	// Restore from saved state if requested
	if *loadState != "" {
		snapshot, err := crawler.LoadState(*loadState)
		if err != nil {
			log.Fatalf("Failed to load state from %s: %v", *loadState, err)
		}
		c.RestoreFrom(snapshot)
		log.Printf("Loaded state from %s", *loadState)
	}

	// Start search server
	srv := server.New(c, idx, fmt.Sprintf(":%d", *port))
	go func() {
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("search server error: %v", err)
		}
	}()

	// If seed URL provided, start crawling immediately (CLI mode)
	if *seedURL != "" {
		log.Printf("Seed URL: %s", *seedURL)
		log.Printf("Config: depth=%d workers=%d queue=%d rate=%s", *maxDepth, *workers, *queueSize, *rateLimit)
		if err := c.Start(*seedURL); err != nil {
			log.Fatalf("Failed to start crawler: %v", err)
		}
	} else {
		log.Printf("No seed URL provided — use the dashboard at http://localhost:%d/dashboard to start crawling", *port)
	}

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutdown signal received...")

	// Save state if requested
	if *saveState != "" {
		state := c.State()
		if state == crawler.StateRunning {
			c.Pause()
			time.Sleep(500 * time.Millisecond) // let workers finish current tasks
		}
		snapshot := c.Snapshot()
		if err := crawler.SaveState(*saveState, snapshot); err != nil {
			log.Printf("Warning: failed to save state: %v", err)
		} else {
			log.Printf("State saved to %s", *saveState)
		}
	}

	// Always save p.data (raw storage format: word url origin depth frequency)
	docs := idx.AllDocuments()
	if len(docs) > 0 {
		pdataPath := "data/storage/p.data"
		if err := crawler.SavePData(pdataPath, docs); err != nil {
			log.Printf("Warning: failed to save p.data: %v", err)
		} else {
			log.Printf("Raw storage saved to %s (%d documents)", pdataPath, len(docs))
		}
	}

	// Stop crawler gracefully
	if c.State() == crawler.StateRunning || c.State() == crawler.StatePaused {
		c.Stop()
	}

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}
