package main

import (
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
	seedURL := flag.String("url", "", "Seed URL to start crawling from (required)")
	maxDepth := flag.Int("depth", 2, "Maximum crawl depth")
	workers := flag.Int("workers", 5, "Number of concurrent crawler workers")
	queueSize := flag.Int("queue", 100, "Maximum frontier queue size (back pressure)")
	timeout := flag.Duration("timeout", 10*time.Second, "HTTP request timeout per page")
	sameDomain := flag.Bool("same-domain", true, "Only crawl pages on the same domain as the seed URL")
	port := flag.Int("port", 8080, "Search server port")

	flag.Parse()

	if *seedURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --url is required")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Starting Google in a Day")
	log.Printf("Seed URL: %s", *seedURL)
	log.Printf("Max depth: %d, Workers: %d, Queue: %d", *maxDepth, *workers, *queueSize)

	// Create shared index
	idx := index.New()

	// Start search server (runs concurrently with crawler)
	srv := server.New(idx, fmt.Sprintf(":%d", *port))
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("search server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	// Start crawler
	c := crawler.New(crawler.Config{
		MaxDepth:   *maxDepth,
		Workers:    *workers,
		QueueSize:  *queueSize,
		Timeout:    *timeout,
		SameDomain: *sameDomain,
	}, idx)

	c.Run(*seedURL)

	// Crawler is done, keep search server running
	log.Printf("Crawling finished. Search server still running on http://localhost:%d", *port)
	log.Printf("Press Ctrl+C to stop.")
	select {} // block forever
}
