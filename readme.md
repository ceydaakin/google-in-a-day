# Google in a Day

A concurrent web crawler and real-time search engine built in Go using only the standard library. Built for BLG483E — AI-Aided Software Engineering.

## Features

- **Recursive web crawling** from a seed URL to configurable depth
- **Live search** — query the index while the crawler is still running
- **Thread-safe** concurrent design using goroutines, channels, and sync primitives
- **Back pressure** — bounded worker pool and frontier queue prevent resource exhaustion
- **Relevancy ranking** — keyword frequency + title match heuristic
- **Zero external dependencies** — Go standard library only
- **Web UI** — simple search interface at `http://localhost:8080`

## Architecture

```
Seed URL → Frontier (bounded channel) → Worker Pool (goroutines)
                                              ↓
                                    Visited Set (sync.Map)
                                    Index (sync.RWMutex)
                                              ↓
                                    Search Server (net/http)
```

- **Frontier**: Bounded channel acting as the URL work queue. Back pressure is applied when the channel is full — new URLs are dropped.
- **Worker Pool**: Fixed number of goroutines fetch pages, parse HTML, extract links, and update the index concurrently.
- **Visited Set**: `sync.Map` guarantees each URL is crawled exactly once using atomic `LoadOrStore`.
- **Inverted Index**: Maps keywords to documents, protected by `sync.RWMutex`. Writers (crawler) take write locks; readers (search) take read locks — enabling live search during crawling.
- **Search Server**: HTTP server runs in a separate goroutine from the start, serving queries against the growing index.

## Usage

```bash
# Build
go build -o search-engine ./cmd/crawler

# Run (crawl golang.org to depth 2 with 5 workers)
./search-engine --url https://golang.org --depth 2 --workers 5

# Search via browser
# Open http://localhost:8080

# Search via API
curl "http://localhost:8080/api/search?q=goroutine"

# Check index stats
curl "http://localhost:8080/api/stats"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--url` | (required) | Seed URL to start crawling |
| `--depth` | 2 | Maximum crawl depth |
| `--workers` | 5 | Number of concurrent crawler goroutines |
| `--queue` | 100 | Frontier queue capacity (back pressure) |
| `--timeout` | 10s | HTTP request timeout per page |
| `--same-domain` | true | Restrict crawling to the seed URL's domain |
| `--port` | 8080 | Search server port |

## Search API

### `GET /api/search?q=<keyword>`

Returns a JSON array of triples:

```json
[
  {
    "relevant_url": "https://golang.org/doc/effective_go",
    "origin_url": "https://golang.org",
    "depth": 1,
    "score": 15.0,
    "title": "Effective Go"
  }
]
```

### `GET /api/stats`

```json
{"documents": 42, "keywords": 1234}
```

## Concurrency Design

1. **Goroutine pool** — fixed N workers consuming from the frontier channel
2. **Bounded channel** — frontier has capacity `queue`; when full, new URLs are dropped (back pressure)
3. **sync.Map** — lock-free visited set for URL deduplication
4. **sync.RWMutex** — protects the inverted index; multiple concurrent readers, exclusive writer
5. **Atomic counter** — tracks in-flight tasks to detect crawl completion
6. **Graceful shutdown** — catches SIGINT/SIGTERM

## Project Structure

```
.
├── cmd/crawler/main.go          # Entry point, CLI flags, wiring
├── internal/
│   ├── crawler/
│   │   ├── crawler.go           # Concurrent crawler with worker pool
│   │   ├── parser.go            # HTML parsing (stdlib only)
│   │   ├── task.go              # CrawlTask type
│   │   └── wordfreq.go          # Word frequency counter
│   ├── index/
│   │   ├── document.go          # Document type
│   │   ├── index.go             # Thread-safe inverted index
│   │   └── result.go            # SearchResult + sorting
│   └── server/
│       └── server.go            # HTTP search server + web UI
├── product_prd.md               # Product Requirement Document
├── recommendation.md            # Production roadmap
├── readme.md                    # This file
└── go.mod
```

## Race Condition Testing

```bash
go run -race ./cmd/crawler --url https://golang.org --depth 1
```

The `-race` flag enables Go's race detector to verify thread safety.
