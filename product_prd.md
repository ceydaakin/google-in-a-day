# Product Requirement Document: Google in a Day

## 1. Overview

A concurrent web crawler and real-time search engine built in Go. The system crawls web pages starting from a seed URL up to a configurable depth, indexes their content in memory, and serves keyword-based search queries — all concurrently and with thread safety. It features a real-time dashboard, crawler lifecycle control, per-domain rate limiting, multi-word search, and persistence for resuming interrupted crawls.

## 2. Problem Statement

Build a functional mini search engine that demonstrates:
- Recursive web crawling with depth control
- Real-time search over a live, growing index
- Proper concurrency management and multi-layered back pressure
- Language-native implementation (Go standard library only)
- System observability with real-time metrics
- Fault tolerance through state persistence

## 3. Target Users

- Developers or students exploring search engine internals
- Course evaluators assessing architectural sensibility and concurrency patterns

## 4. Functional Requirements

### 4.1 Indexer (Crawler)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-1 | Initiate crawling from a user-specified origin URL | Must | Done |
| FR-2 | Crawl recursively up to a configurable maximum depth `k` | Must | Done |
| FR-3 | Extract and follow hyperlinks (`<a href="...">`) from crawled pages | Must | Done |
| FR-4 | Store page content (title, body text, URL, origin URL, depth) in an in-memory index | Must | Done |
| FR-5 | Skip already-visited URLs (uniqueness guarantee) | Must | Done |
| FR-6 | Respect back pressure: limit concurrent workers, queue depth, and per-domain rate | Must | Done |
| FR-7 | Use only Go standard library — no third-party crawling/scraping libraries | Must | Done |
| FR-8 | Handle HTTP errors and timeouts gracefully | Should | Done |
| FR-9 | Restrict crawling to same-domain or configurable domain scope | Should | Done |
| FR-10 | Support pause, resume, and stop controls | Must | Done |
| FR-11 | Per-domain rate limiting (configurable interval) | Must | Done |

### 4.2 Searcher (Query Engine)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-12 | Accept a keyword query from the user | Must | Done |
| FR-13 | Return a list of triples: `(relevant_url, origin_url, depth)` where depth is the actual crawl depth (hops from origin) | Must | Done |
| FR-14 | Search must work while the indexer is still actively crawling (live indexing) | Must | Done |
| FR-15 | Rank results using a relevancy heuristic (keyword frequency + title match bonus) | Must | Done |
| FR-16 | Thread-safe read access to the index concurrent with write access from the crawler | Must | Done |
| FR-17 | Multi-word search with AND semantics and combined scoring | Must | Done |
| FR-18 | Return results within reasonable latency (<500ms for indexed content) | Should | Done |

### 4.3 Interface

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-19 | CLI interface to start crawling with flags | Must | Done |
| FR-20 | HTTP endpoint to submit search queries | Must | Done |
| FR-21 | Simple web UI for search (HTML form + results page) | Must | Done |
| FR-22 | Real-time dashboard showing indexing progress, queue depth, and back pressure | Must | Done |
| FR-23 | Dashboard controls to start, pause, resume, stop crawling | Must | Done |
| FR-24 | REST API for crawler lifecycle management | Must | Done |

### 4.4 Persistence (Bonus)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-25 | Save crawl state to disk on shutdown | Nice-to-have | Done |
| FR-26 | Resume from saved state without recrawling | Nice-to-have | Done |

## 5. Non-Functional Requirements

| ID | Requirement | Status |
|----|-------------|--------|
| NFR-1 | **Concurrency**: Use goroutines and channels for parallel crawling | Done |
| NFR-2 | **Thread Safety**: Use `sync.RWMutex`, `sync.Map`, and atomics for shared data structures | Done |
| NFR-3 | **Back Pressure**: Three layers — bounded frontier, fixed worker pool, per-domain rate limiter | Done |
| NFR-4 | **No External Dependencies**: Only Go standard library packages | Done |
| NFR-5 | **Graceful Shutdown**: Handle SIGINT, drain in-flight requests, save state | Done |
| NFR-6 | **Logging**: Structured log output for crawl progress | Done |
| NFR-7 | **Test Coverage**: 80%+ coverage with race condition detection | Done |
| NFR-8 | **Immutability**: Metrics snapshots and document copies are immutable value types | Done |

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
              │     ┌──────▼──────┐     │
              └────►│ Rate Limiter │◄───┘
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Visited Set  │  (sync.Map)
                    │ + Index      │  (sync.RWMutex)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ HTTP Server  │  (Search + Dashboard + API)
                    └─────────────┘
```

### Core Components

1. **Frontier** — Bounded channel of `CrawlTask{URL, OriginURL, Depth}`. Natural back pressure via channel capacity.

2. **Worker Pool** — Fixed goroutines consuming from the frontier. Each worker fetches, parses, and indexes. Supports pause/resume via atomic state checks.

3. **Rate Limiter** — Per-domain throttle with configurable minimum interval between requests to the same host.

4. **Visited Set** — `sync.Map` for lock-free URL deduplication via `LoadOrStore`.

5. **Inverted Index** — Thread-safe keyword-to-document mapping with `sync.RWMutex`. Supports single and multi-word queries.

6. **Metrics Collector** — Atomic counters and mutex-protected history for real-time monitoring.

7. **HTTP Server** — Serves search UI, real-time dashboard, and REST API concurrently with the crawler.

## 7. Data Structures

```go
// CrawlTask — work unit in the frontier
type CrawlTask struct {
    URL, OriginURL string
    Depth          int
}

// Document — indexed page
type Document struct {
    URL, OriginURL string
    Depth          int            // actual hop count from origin
    MaxDepth       int            // k parameter passed to /index
    Title          string
    WordFreq       map[string]int
}

// Index — inverted index (keyword -> []Document)
type Index struct {
    mu   sync.RWMutex
    docs map[string][]*Document
    all  []*Document
}

// Metrics — atomic counters + worker state
type Metrics struct {
    urlsProcessed, urlsQueued, urlsDropped, urlsErrored int64  // atomic
    workerStatus map[int]WorkerInfo                             // mutex-protected
    history      []HistoryEntry                                 // bounded ring buffer
}

// CrawlState — lifecycle state machine
type CrawlState int32  // Idle → Running ⇄ Paused → Stopped/Completed
```

## 8. Relevancy Heuristic

Score for a document given a query:
```
single-word: score = (frequency × 10) + 1000 − (depth × 5)
multi-word:  score = (Σ frequency_i × 10) + 1000 − (depth × 5)
```

| Component | Description |
|-----------|-------------|
| `frequency × 10` | Higher word frequency = more relevant |
| `+ 1000` | Exact keyword match bonus |
| `− depth × 5` | Penalty for pages farther from the seed URL |

- Multi-word uses AND semantics — only documents containing ALL keywords are returned
- For multi-word queries, the combined frequency of all keywords is used
- Results are sorted by score descending

## 9. Back Pressure Strategy

Three independent layers:

1. **Bounded frontier channel**: capacity = `--queue` flag. When full, new URLs are dropped with logging.
2. **Fixed worker pool**: `--workers` flag (default: 5). Limits concurrent HTTP requests.
3. **Per-domain rate limiter**: `--rate-limit` flag (default: 500ms). Minimum delay between requests to the same host. Prevents overwhelming any single server.

## 10. Crawler State Machine

```
                ┌─────────┐
                │  Idle    │
                └────┬────┘
                     │ Start()
                ┌────▼────┐
           ┌───►│ Running  │◄───┐
           │    └────┬────┘    │
           │         │ Pause() │ Resume()
           │    ┌────▼────┐    │
           │    │ Paused   │───┘
           │    └────┬────┘
           │         │ Stop()
           │    ┌────▼────┐
           └────│ Stopped  │
     (natural)  └─────────┘
                ┌─────────┐
                │Completed │  (all tasks done)
                └─────────┘
```

## 11. Success Criteria

- [x] Crawler correctly discovers and indexes pages up to depth k
- [x] Search returns relevant results while crawling is in progress
- [x] No data races (verified with `go test -race`)
- [x] System handles back pressure without memory blowup or deadlock
- [x] Multi-word search returns correct AND intersections
- [x] Dashboard shows real-time metrics with auto-refresh
- [x] Pause/resume/stop work correctly via dashboard and API
- [x] State persistence saves and restores successfully
- [x] Per-domain rate limiting enforced
- [x] 80%+ test coverage on core packages
- [x] All code can be explained and design choices justified

## 12. Design Decisions

| Decision | Rationale |
|----------|-----------|
| Go standard library only | Assignment constraint + demonstrates language mastery |
| Bounded channel for frontier | Natural back pressure via Go channel semantics |
| sync.Map for visited set | Lock-free concurrent reads/writes, ideal for write-once patterns |
| sync.RWMutex for index | Multiple readers (search) concurrent with single writer (crawler) |
| Atomic state machine | Lock-free lifecycle control, no mutex needed for state checks |
| Per-domain rate limiter (not global) | Prevents hammering individual servers while allowing parallelism across domains |
| JSON persistence (not database) | Assignment scope — simple, portable, stdlib-compatible |
| Metrics snapshot pattern | Immutable copies prevent data races in dashboard rendering |
| Workers poll state for pause | Simpler than channel replacement; avoids channel recreation races |
| URL normalization (trailing slash) | Prevents `/about` and `/about/` from being crawled as separate pages |
| Body size limit (5 MB) | Bounds memory per page; prevents OOM on very large documents |
| Search depth = crawl depth | Triple's `depth` is the actual hop count from seed URL, reflecting how deep the page was discovered |
