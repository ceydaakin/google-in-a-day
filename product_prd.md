# Product Requirement Document: Google in a Day

## 1. Overview

A concurrent web crawler and real-time search engine built in Go. The system crawls web pages starting from a seed URL up to a configurable depth, indexes their content in memory, and serves keyword-based search queries — all concurrently and with thread safety.

## 2. Problem Statement

Build a functional mini search engine that demonstrates:
- Recursive web crawling with depth control
- Real-time search over a live, growing index
- Proper concurrency management and back pressure handling
- Language-native implementation (Go standard library only)

## 3. Target Users

- Developers or students exploring search engine internals
- Course evaluators assessing architectural sensibility and concurrency patterns

## 4. Functional Requirements

### 4.1 Indexer (Crawler)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-1 | Initiate crawling from a user-specified origin URL | Must |
| FR-2 | Crawl recursively up to a configurable maximum depth `k` | Must |
| FR-3 | Extract and follow hyperlinks (`<a href="...">`) from crawled pages | Must |
| FR-4 | Store page content (title, body text, URL, origin URL, depth) in an in-memory index | Must |
| FR-5 | Skip already-visited URLs (uniqueness guarantee) | Must |
| FR-6 | Respect back pressure: limit concurrent workers and queue depth | Must |
| FR-7 | Use only Go standard library (`net/http`, `html` tokenizer) — no third-party crawling/scraping libraries | Must |
| FR-8 | Handle HTTP errors and timeouts gracefully | Should |
| FR-9 | Restrict crawling to same-domain or configurable domain scope | Should |

### 4.2 Searcher (Query Engine)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-10 | Accept a keyword query from the user | Must |
| FR-11 | Return a list of triples: `(relevant_url, origin_url, depth)` | Must |
| FR-12 | Search must work while the indexer is still actively crawling (live indexing) | Must |
| FR-13 | Rank results using a relevancy heuristic (keyword frequency + title match bonus) | Must |
| FR-14 | Thread-safe read access to the index concurrent with write access from the crawler | Must |
| FR-15 | Return results within reasonable latency (<500ms for indexed content) | Should |

### 4.3 Interface

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-16 | CLI interface to start crawling with `--url`, `--depth`, `--workers` flags | Must |
| FR-17 | CLI or HTTP endpoint to submit search queries | Must |
| FR-18 | Simple web UI for search (HTML form + results page) | Nice-to-have |

## 5. Non-Functional Requirements

| ID | Requirement |
|----|-------------|
| NFR-1 | **Concurrency**: Use goroutines and channels for parallel crawling |
| NFR-2 | **Thread Safety**: Use `sync.RWMutex` or `sync.Map` for shared data structures |
| NFR-3 | **Back Pressure**: Bounded worker pool (semaphore pattern) + bounded work queue |
| NFR-4 | **No External Dependencies**: Only Go standard library packages |
| NFR-5 | **Graceful Shutdown**: Handle SIGINT, drain in-flight requests |
| NFR-6 | **Logging**: Structured log output for crawl progress |

## 6. System Architecture

```
                    ┌─────────────┐
                    │  Seed URL    │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Frontier    │  (bounded channel)
                    │  (URL Queue) │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
        ┌─────▼───┐  ┌────▼────┐  ┌────▼────┐
        │ Worker 1 │  │ Worker 2│  │ Worker N│   (goroutine pool)
        └─────┬───┘  └────┬────┘  └────┬────┘
              │            │            │
              └────────────┼────────────┘
                           │
                    ┌──────▼──────┐
                    │ Visited Set  │  (sync.Map)
                    │ + Index      │  (sync.RWMutex)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Search API   │  (HTTP server / CLI)
                    └─────────────┘
```

### Core Components

1. **Frontier** — Bounded channel of `CrawlTask{URL, OriginURL, Depth}`. Acts as the work queue with natural back pressure (channel capacity).

2. **Worker Pool** — Fixed number of goroutines consuming from the frontier. Each worker fetches a page, parses links, and sends new tasks back to the frontier.

3. **Visited Set** — `sync.Map` storing visited URLs. Checked before enqueuing to guarantee uniqueness.

4. **Index** — In-memory inverted index mapping keywords to document entries. Protected by `sync.RWMutex` (writers acquire write lock, searchers acquire read lock).

5. **Search Server** — HTTP server running concurrently with the crawler. Reads from the index using read locks. Returns ranked results as `(relevant_url, origin_url, depth)`.

## 7. Data Structures

```go
// CrawlTask represents a URL to be crawled
type CrawlTask struct {
    URL       string
    OriginURL string
    Depth     int
}

// Document represents a crawled page
type Document struct {
    URL       string
    OriginURL string
    Depth     int
    Title     string
    Body      string
    WordFreq  map[string]int
}

// Index is the inverted index: keyword -> []Document
type Index struct {
    mu   sync.RWMutex
    docs map[string][]*Document  // keyword -> documents containing it
}
```

## 8. Relevancy Heuristic

Score for a document given a query keyword:
```
score = word_frequency(keyword) + title_match_bonus
```
- `word_frequency`: count of keyword occurrences in the page body (normalized by total words)
- `title_match_bonus`: +10 if keyword appears in the page title

Results are sorted by score descending.

## 9. Back Pressure Strategy

1. **Bounded frontier channel**: capacity = `workers * 10`. When full, producers block — this naturally throttles discovery.
2. **Fixed worker pool**: `--workers` flag (default: 5). No more than N concurrent HTTP requests.
3. **HTTP client timeout**: 10 seconds per request to prevent slow pages from blocking workers.
4. **Optional rate limiter**: Configurable delay between requests to the same domain.

## 10. Success Criteria

- Crawler correctly discovers and indexes pages up to depth k
- Search returns relevant results while crawling is in progress
- No data races (verifiable with `go run -race`)
- System handles back pressure without memory blowup or deadlock
- All code can be explained and design choices justified

## 11. Out of Scope

- Persistent storage (disk/database)
- Distributed crawling across multiple machines
- JavaScript rendering (SPA pages)
- robots.txt compliance (for this prototype)
- Full-text search / stemming / NLP
