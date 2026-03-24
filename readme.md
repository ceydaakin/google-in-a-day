# Google in a Day

A concurrent web crawler and real-time search engine built in Go using only the standard library. Built for BLG483E — AI-Aided Software Engineering at Istanbul Technical University.

## Features

- **Recursive web crawling** from a seed URL to configurable depth
- **Live search** — query the index while the crawler is still running
- **Real-time dashboard** — monitor indexing progress, queue depth, worker status, and back pressure
- **Crawler lifecycle control** — start, pause, resume, stop via dashboard or API
- **Per-domain rate limiting** — configurable delay between requests to the same host
- **Multi-word search** — AND semantics across multiple keywords with combined scoring
- **Back pressure** — bounded worker pool, frontier queue, and rate limiting prevent resource exhaustion
- **Persistence** — save and restore crawl state for resuming after interruption
- **Thread-safe** concurrent design using goroutines, channels, and sync primitives
- **Relevancy ranking** — keyword frequency + title match heuristic
- **Zero external dependencies** — Go standard library only
- **80%+ test coverage** with race condition detection

## Architecture

```
                    ┌─────────────┐
                    │  Seed URL    │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Frontier    │  (bounded channel — back pressure)
                    │  (URL Queue) │
                    └──────┬──────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
        ┌─────▼───┐  ┌────▼────┐  ┌────▼────┐
        │ Worker 1 │  │ Worker 2│  │ Worker N│  (goroutine pool)
        └─────┬───┘  └────┬────┘  └────┬────┘
              │            │            │
              │     ┌──────▼──────┐     │
              └────►│ Rate Limiter │◄───┘  (per-domain throttle)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Visited Set  │  (sync.Map)
                    │ + Index      │  (sync.RWMutex)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ HTTP Server  │  (Search UI + Dashboard + API)
                    └─────────────┘
```

### Core Components

1. **Frontier** — Bounded channel of `CrawlTask{URL, OriginURL, Depth}`. Acts as the work queue with natural back pressure (channel capacity). When full, new URLs are dropped.

2. **Worker Pool** — Fixed number of goroutines consuming from the frontier. Each worker fetches a page, parses links, and sends new tasks back to the frontier. Workers support pause/resume via atomic state checks.

3. **Rate Limiter** — Per-domain throttle ensuring a minimum interval between requests to the same host. Prevents overwhelming any single server.

4. **Visited Set** — `sync.Map` storing visited URLs. Checked before enqueuing to guarantee uniqueness using atomic `LoadOrStore`.

5. **Inverted Index** — Maps keywords to document entries. Protected by `sync.RWMutex` (writers acquire write lock, searchers acquire read lock). Supports multi-word AND queries.

6. **HTTP Server** — Runs concurrently with the crawler from startup. Serves:
   - Search UI at `/`
   - Real-time dashboard at `/dashboard`
   - REST API for search, status, and crawler lifecycle

## Prerequisites

- **Go 1.22+** — [Download and install](https://go.dev/dl/)
- **Git** (optional) — only needed if cloning from GitHub

Verify Go is installed:

```bash
go version
# go version go1.22.0 (or newer)
```

## Installation

**Option A — Clone from GitHub:**

```bash
git clone https://github.com/ceydaakin/google-in-a-day.git
cd google-in-a-day
```

**Option B — From ZIP:**

```bash
unzip google-in-a-day.zip
cd google-in-a-day
```

## Quick Start

```bash
# 1. Build the binary
go build -o search-engine ./cmd/crawler

# 2a. Run with a seed URL (starts crawling immediately)
./search-engine --url https://golang.org --depth 2 --workers 5
# Search UI:  http://localhost:3600
# Dashboard:  http://localhost:3600/dashboard

# 2b. Or run without a URL and start crawling from the dashboard
./search-engine --port 3600
# Open http://localhost:3600/dashboard and use the "Start Crawl" form
```

## Usage

### CLI Mode

```bash
# Crawl golang.org to depth 2 with 5 workers
./search-engine --url https://golang.org --depth 2 --workers 5

# With rate limiting (1 second between requests to same domain)
./search-engine --url https://golang.org --rate-limit 1s

# Save state on shutdown (Ctrl+C)
./search-engine --url https://golang.org --save-state crawl.json

# Resume from saved state
./search-engine --load-state crawl.json
```

### Dashboard Mode

Start without a URL and use the web dashboard:

```bash
./search-engine --port 8080
# Open http://localhost:3600/dashboard
```

The dashboard provides:
- **Start Crawl** form — enter URL, depth, workers, queue size
- **Real-time metrics** — URLs processed, queued, dropped, errors, active workers
- **Progress bar** — visual indicator of crawl completion
- **Pause/Resume/Stop** controls
- **Worker status** table showing what each goroutine is doing
- **Crawl history** — last 20 URLs with status codes and durations
- **Back pressure indicator** — count of dropped URLs and back pressure events

The frontend uses a separated design system (`web/static/css/style.css`) with CSS custom properties, embedded into the binary via Go's `embed` package for single-binary deployment.

### Search

```bash
# Via browser
open http://localhost:3600

# Single keyword search via API
curl "http://localhost:3600/api/search?q=goroutine"

# Multi-word search (AND semantics)
curl "http://localhost:3600/api/search?q=go+concurrency"
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--url` | (optional) | Seed URL to start crawling |
| `--depth` | 2 | Maximum crawl depth |
| `--workers` | 5 | Number of concurrent crawler goroutines |
| `--queue` | 100 | Frontier queue capacity (back pressure) |
| `--timeout` | 10s | HTTP request timeout per page |
| `--same-domain` | true | Restrict crawling to the seed URL's domain |
| `--port` | 3600 | Search server port |
| `--rate-limit` | 500ms | Minimum delay between requests to the same domain |
| `--save-state` | (disabled) | Path to save crawl state on shutdown |
| `--load-state` | (disabled) | Path to load previous crawl state from |

## REST API

### Search

**`GET /api/search?q=<keywords>`** — Returns JSON array of triples:

```json
[
  {
    "relevant_url": "https://golang.org/doc/effective_go",
    "origin_url": "https://golang.org",
    "depth": 2,
    "score": 15.0,
    "title": "Effective Go"
  }
]
```

> **Note:** `depth` in the triple is the actual crawl depth (number of hops from the seed URL) at which the page was discovered. `origin_url` identifies the seed URL that initiated the crawl.

### Index Stats

**`GET /api/stats`**

```json
{"documents": 42, "keywords": 1234}
```

### Crawler Lifecycle

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/index` | POST | Start crawling `{"origin": "...", "k": 2}` |
| `/api/status` | GET | Rich status with metrics, workers, history |
| `/api/pause` | POST | Pause the crawler |
| `/api/resume` | POST | Resume a paused crawler |
| `/api/stop` | POST | Stop the crawler |
| `/api/save` | POST | Save crawl state to disk |

### Status Response

**`GET /api/status`**

```json
{
  "state": "running",
  "seed_url": "https://golang.org",
  "metrics": {
    "urls_processed": 25,
    "urls_queued": 142,
    "urls_dropped": 3,
    "urls_errored": 2,
    "active_workers": 5,
    "back_pressures": 3,
    "workers": [
      {"id": 0, "state": "fetching", "url": "https://golang.org/doc/..."},
      {"id": 1, "state": "idle", "url": ""}
    ],
    "history": [
      {"url": "...", "status_code": 200, "duration_ms": 150000000, "timestamp": "..."}
    ]
  },
  "docs": 25,
  "keywords": 1842
}
```

### Search with Relevance Sorting

**`GET /search?query=<keywords>&sortBy=relevance`** — Returns JSON with `relevance_score`:

```json
[
  {
    "relevant_url": "https://go.dev/doc/effective_go",
    "origin_url": "https://go.dev",
    "depth": 1,
    "relevance_score": 1045,
    "title": "Effective Go"
  }
]
```

This endpoint supports both `q` and `query` parameters. When `sortBy` is provided, the response is JSON instead of HTML.

## Scoring Formula

Each search result is scored using the following formula:

```
score = (frequency × 10) + 1000 (exact match bonus) − (depth × 5)
```

| Component | Description |
|-----------|-------------|
| `frequency × 10` | Higher word frequency = more relevant |
| `+ 1000` | Exact keyword match bonus |
| `− depth × 5` | Penalty for pages farther from the seed URL |

**Example:** A page at depth 2 with word frequency 8:
```
score = (8 × 10) + 1000 − (2 × 5) = 80 + 1000 − 10 = 1070
```

For multi-word queries, AND semantics are applied (all keywords must appear in the page), and the combined frequency of all keywords is used in the formula.

## Raw Storage Format

On shutdown, the crawler saves the inverted index to `data/storage/p.data` as a flat text file. Each line contains one word-document entry:

```
word url origin depth frequency
```

| Field | Description |
|-------|-------------|
| `word` | Indexed keyword (lowercase) |
| `url` | URL of the page where the word was found |
| `origin` | Seed URL that initiated the crawl |
| `depth` | Crawl depth (hops from seed URL) |
| `frequency` | Number of times the word appears on the page |

**Example entries:**
```
build https://go.dev https://go.dev 0 5
engineer https://go.dev https://go.dev 0 3
tutorial https://go.dev/learn https://go.dev 1 2
```

## Concurrent Search During Indexing

Search runs concurrently with crawling from the moment the server starts. This is possible because the inverted index is protected by a `sync.RWMutex`:

- **Write path** (crawler): Acquires an exclusive write lock when adding a new document. Only one writer at a time.
- **Read path** (search): Acquires a shared read lock. Multiple searches execute simultaneously without blocking each other.
- **Consistency**: A search sees all documents indexed before the read lock was acquired. Documents being added concurrently appear in the next search.

This design means search latency is unaffected by crawl activity, and crawl throughput is unaffected by search queries (except for brief lock hand-offs during document insertion).

## Concurrency Design

1. **Goroutine pool** — fixed N workers consuming from the frontier channel
2. **Bounded channel** — frontier has capacity `queue`; when full, new URLs are dropped (back pressure)
3. **sync.Map** — lock-free visited set for URL deduplication
4. **sync.RWMutex** — protects the inverted index; multiple concurrent readers, exclusive writer
5. **Atomic counters** — tracks in-flight tasks, worker state, and metrics without lock contention
6. **Per-domain rate limiter** — mutex-protected per-host timestamp tracking with configurable intervals
7. **Pause/resume** — atomic state variable checked by workers between tasks; no channel recreation needed
8. **Graceful shutdown** — SIGINT/SIGTERM triggers: pause → save state → stop workers → drain HTTP → exit

## Back Pressure Strategy

Three layers of back pressure protect the system:

1. **Bounded frontier** — channel capacity limits queued URLs. Excess URLs are dropped with logging.
2. **Fixed worker pool** — limits concurrent HTTP requests to `--workers` count.
3. **Per-domain rate limiter** — ensures minimum `--rate-limit` delay between requests to the same host.

## Persistence & Resume

Save crawl state on shutdown:
```bash
./search-engine --url https://golang.org --save-state crawl.json
# Press Ctrl+C to stop — state is saved automatically
```

Resume from saved state:
```bash
./search-engine --load-state crawl.json
# Index is restored, search works immediately
```

The saved state includes: all indexed documents, visited URL set, and crawler configuration.

## Testing

```bash
# Run all tests with race detector
go test -race ./...

# Run with coverage
go test -race -cover ./...

# Verify no race conditions during crawling
go run -race ./cmd/crawler --url https://golang.org --depth 1
```

## Project Structure

```
.
├── cmd/crawler/main.go              # Entry point, CLI flags, graceful shutdown
├── internal/
│   ├── crawler/
│   │   ├── crawler.go               # Concurrent crawler with lifecycle control
│   │   ├── crawler_test.go          # Crawler tests with httptest
│   │   ├── metrics.go               # Thread-safe metrics collector
│   │   ├── parser.go                # HTML parsing (stdlib only)
│   │   ├── parser_test.go           # Parser tests
│   │   ├── persistence.go           # Save/load crawl state
│   │   ├── persistence_test.go      # Persistence tests
│   │   ├── ratelimiter.go           # Per-domain rate limiting
│   │   ├── ratelimiter_test.go      # Rate limiter tests
│   │   ├── state.go                 # Crawler state machine
│   │   ├── task.go                  # CrawlTask type
│   │   ├── pdata.go                 # Raw storage writer (p.data format)
│   │   ├── wordfreq.go             # Word frequency counter
│   │   └── wordfreq_test.go        # Word frequency tests
│   ├── index/
│   │   ├── document.go             # Document type
│   │   ├── index.go                # Thread-safe inverted index + multi-word search
│   │   ├── index_test.go           # Index tests (including concurrency)
│   │   └── result.go               # SearchResult + sorting
│   └── server/
│       ├── api.go                  # REST API handlers (lifecycle + save state)
│       ├── dashboard.go            # Dashboard HTML layout
│       ├── server.go               # HTTP server wiring + static file serving
│       ├── server_test.go          # HTTP handler tests
│       └── templates.go            # Search UI HTML layouts
├── web/
│   ├── embed.go                    # Embeds static assets into the binary
│   └── static/
│       ├── css/
│       │   └── style.css           # Design system (variables, components, responsive)
│       └── js/
│           ├── home.js             # Home page stats bar
│           └── dashboard.js        # Dashboard polling, controls, metrics
├── data/
│   └── storage/
│       └── p.data                  # Raw inverted index (word url origin depth frequency)
├── product_prd.md                  # Product Requirement Document
├── recommendation.md               # Production roadmap
├── ai_stewardship_notes.md         # AI design decisions & justifications (17 Q&A)
├── readme.md                       # This file
└── go.mod                          # Zero external dependencies
```
