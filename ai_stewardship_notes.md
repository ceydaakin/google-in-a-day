# AI Stewardship - Design Decisions & Code Walkthrough

This document explains the AI-generated code and justifies design choices made throughout the project.

---

## 1. Why Go?

Go's concurrency model (goroutines + channels) is ideal for this project. Goroutines are lightweight (a few KB stack), enabling thousands to run concurrently. Channels provide a natural communication and synchronization mechanism. Go's standard library (`net/http`, `sync`, `encoding/json`) covers all requirements without external dependencies.

---

## 2. Why `sync.Map` instead of `map` + `Mutex` for the visited set?

`sync.Map` is optimized for the "write-once, read-many" pattern. In our visited URL set, each URL is written once and then only checked — exactly this pattern. `LoadOrStore` provides atomic check-and-set, ensuring that even if two workers encounter the same URL simultaneously, only one succeeds. A regular `map` + `Mutex` would also work but creates unnecessary lock contention on every read.

---

## 3. How does back pressure work?

Three independent layers:

1. **Bounded frontier channel** (`--queue` flag, default 100): When the URL queue is full, new URLs are dropped. Go channel capacity provides this naturally — we use `select/default` for non-blocking sends.

2. **Fixed worker pool** (`--workers` flag, default 5): A fixed number of goroutines limits concurrent HTTP requests. No new workers are spawned.

3. **Per-domain rate limiter** (`--rate-limit` flag, default 500ms): Enforces a minimum delay between consecutive requests to the same host. Prevents overwhelming any single server while allowing parallelism across domains.

**Why 3 layers?** A single layer is insufficient. Queue limits alone would still allow 5 workers hitting the same site simultaneously. Rate limiting alone would allow unbounded queue growth. The layers complement each other.

---

## 4. How does concurrent search work without race conditions?

We use `sync.RWMutex`:

- **Write lock** (`Lock()`): Acquired by the crawler when adding a new document. Exclusive — only one writer at a time.
- **Read lock** (`RLock()`): Acquired by search queries. Shared — multiple readers can search simultaneously.

Search and crawl run fully concurrently. Search sees the index as of the moment the read lock is acquired — newly added pages appear in subsequent searches. Verified with `go test -race` (zero race conditions).

---

## 5. How does pause/resume work?

We use an atomic state variable (`int32`). Workers check `state == StatePaused` between each task:

```go
for c.state.Load() == StatePaused {
    select {
    case <-c.stopCh:
        return
    default:
        time.Sleep(100 * time.Millisecond)
    }
}
```

**Why not close and recreate channels?** Channel recreation is prone to race conditions — another goroutine might try to send on a closed channel (panic). Atomic state polling is simpler and safer. The 100ms sleep is negligible overhead.

---

## 6. Why does the state machine exist?

The crawler has 5 states: `Idle -> Running <-> Paused -> Stopped / Completed`

- `atomic.CompareAndSwap` ensures lock-free, thread-safe state transitions
- Invalid transitions return errors (e.g., cannot pause an idle crawler)
- Dashboard and API use state to enable/disable controls
- Graceful shutdown uses state to ensure correct ordering

---

## 7. Why JSON persistence instead of a database?

- **Stdlib-only constraint**: Even SQLite is an external dependency. JSON marshaling is in Go's stdlib.
- **Portable**: Single file, runs on any machine, no setup required.
- **Atomic write**: We write to `state.json.tmp` first, then `os.Rename` for atomic rename. If the process crashes, either the old file or the new file exists — never a corrupted file.
- **Sufficient**: At this scale (single machine, in-memory index), a database would be overkill.

---

## 8. Why AND semantics for multi-word search?

OR semantics returns too much noise — searching "go web" would include thousands of pages containing only "go". AND semantics returns only pages containing **all** keywords, producing more relevant results. Google also defaults to AND.

**How it works:** We create a posting list (keyword -> document list) for each keyword, then take the intersection. We start with the smallest set — this optimization avoids unnecessary comparisons.

---

## 9. Why do metrics use the snapshot pattern?

`MetricsSnapshot` is an immutable struct — it cannot be modified after creation. This matters because:

- The dashboard polls `GET /api/status` every 2 seconds
- If we returned the metrics struct directly, the crawler could modify counters while JSON is being encoded -> torn read
- By taking a snapshot, we get a consistent point-in-time view
- Atomic counters (`int64`) provide lock-free reading; only worker status and history use `RWMutex`

---

## 10. Why is graceful shutdown important and how does it work?

`os.Exit(0)` doesn't run defers, workers are cut mid-task, and state is lost.

Our shutdown sequence:
1. SIGINT/SIGTERM is caught
2. If the crawler is running, it is paused (workers finish current tasks)
3. If `--save-state` is set, state is written to disk
4. Crawler is stopped (`stopCh` is closed, workers exit)
5. HTTP server drains active connections with a 5-second timeout via `Shutdown(ctx)`
6. `main()` returns normally

---

## 11. Why does the dashboard use polling instead of WebSocket?

- **Simplicity**: WebSocket requires upgrade handshake, connection management, and reconnection logic.
- **2-second polling is sufficient**: The crawler processes a few pages per second; real-time updates (<100ms) are not needed.
- **Production upgrade**: WebSocket or Server-Sent Events should be used — noted in recommendation.md.

---

## 12. Search result triple: why `depth` = k parameter

The spec defines the search triple as `(relevant_url, origin_url, depth)` where "origin and depth define parameters passed to /index." So `depth` in the triple is the `k` parameter (max crawl depth) passed to the index call, not the actual number of hops. This correctly identifies which `/index` invocation discovered a given URL.

---

## 13. Test strategy

- **Unit tests**: Core functions in every package (parser, wordfreq, index, rate limiter, persistence)
- **Integration tests**: Real HTTP server via `httptest.NewServer` to test the crawler end-to-end
- **Race detection**: All tests run with `go test -race` (Go's built-in race detector)
- **Concurrent safety test**: 10 writer + 5 reader goroutines accessing the index simultaneously, zero races
- **Coverage**: crawler 81%+, index 96%+, server 95%+ — exceeds 80% target across all packages

---

## 14. URL normalization

URLs like `/about` and `/about/` point to the same page but would be treated as distinct without normalization. We strip trailing slashes during link resolution (in `resolveURL`) so the visited set treats them as one URL. Fragment removal (`#section`) is also applied. This prevents duplicate crawls and wasted work.

---

## 15. Body size limit (5 MB)

`parsePage` wraps the HTTP response body in `io.LimitReader(body, 5*1024*1024)` before reading. This caps memory usage per page at 5 MB. For a crawl indexing 100K pages, this bounds the worst-case memory spike from any single page. Without this limit, a malicious or misconfigured page could return gigabytes and OOM the process.

---

## 16. Why does the visited set reset between crawls?

Each call to `Start()` clears the `sync.Map` visited set. This ensures separate `/index` calls are independent — `index("a.com", 2)` followed by `index("b.com", 3)` should crawl b.com's pages even if some overlap with a.com. The inverted index accumulates documents across crawls (append-only), but the frontier and visited set are per-crawl.

---

## 17. Race condition fix in server initialization

Originally, `httpSrv` was assigned inside `Start()` and read by `Shutdown()` from a different goroutine — a data race. The fix: move route registration and `httpSrv` creation into `New()`, so the field is set before any goroutine can call `Shutdown()`. `Start()` now only calls `ListenAndServe()`. Detected by `go test -race`.
