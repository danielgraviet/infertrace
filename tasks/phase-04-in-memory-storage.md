# Phase 4: In-Memory Storage + Go Interfaces

> Before integrating a real database, you'll build an in-memory store. This phase teaches two critical Go concepts: interfaces (how Go achieves polymorphism without inheritance) and mutex-protected maps (how you share data safely across goroutines). The interface you define here will make swapping in TimescaleDB in Phase 5 seamless.

---

## Learning Objectives

By the end of this phase you will understand:

- How Go interfaces work — they are satisfied implicitly (no `implements` keyword)
- Why you define interfaces at the *consumer* side, not the *producer* side
- How `sync.RWMutex` prevents data races on shared maps
- The difference between a read lock and a write lock
- How Go's `map` is not safe for concurrent access (unlike Python's GIL-protected dict)
- How to write table-driven tests in Go

---

## Tasks

1. **Define the `StorageWriter` interface** in `internal/storage/storage.go`:
   ```go
   type StorageWriter interface {
       Write(span *proto.Span) error
       GetTrace(traceID string) ([]*proto.Span, error)
       ListSpans(filter SpanFilter) ([]*proto.Span, error)
   }

   type SpanFilter struct {
       ServiceName string
       ModelName   string
       Since       time.Time
       Until       time.Time
       Limit       int
   }
   ```
   This interface is the contract. Both `MemoryStore` (this phase) and `TimescaleStore` (Phase 5) will implement it.

2. **Implement `MemoryStore`** in `internal/storage/memory.go`:
   ```go
   type MemoryStore struct {
       mu     sync.RWMutex
       traces map[string][]*proto.Span  // traceID → spans
   }

   func NewMemoryStore() *MemoryStore {
       return &MemoryStore{
           traces: make(map[string][]*proto.Span),
       }
   }
   ```

3. **Implement `Write`:**
   ```go
   func (m *MemoryStore) Write(span *proto.Span) error {
       traceID := string(span.TraceId)
       m.mu.Lock()
       defer m.mu.Unlock()
       m.traces[traceID] = append(m.traces[traceID], span)
       return nil
   }
   ```
   Understand the `defer m.mu.Unlock()` pattern — `defer` runs when the function returns, so you never forget to unlock even if the function panics.

4. **Implement `GetTrace`** — use `RLock` (read lock) since you're only reading:
   ```go
   func (m *MemoryStore) GetTrace(traceID string) ([]*proto.Span, error) {
       m.mu.RLock()
       defer m.mu.RUnlock()
       spans, ok := m.traces[traceID]
       if !ok {
           return nil, fmt.Errorf("trace %s not found", traceID)
       }
       return spans, nil
   }
   ```
   Key insight: multiple goroutines can hold a read lock simultaneously, but a write lock is exclusive. This is why reads don't block each other.

5. **Implement `ListSpans`** with basic filtering:
   - Iterate over all traces and spans
   - Filter by `ServiceName` and `ModelName` if set in the filter
   - Filter by `Since`/`Until` using `span.StartTimeUnixNano`
   - Respect the `Limit` field

6. **Implement a TTL purge** — add a `StartPurge(interval time.Duration)` method:
   ```go
   func (m *MemoryStore) StartPurge(interval time.Duration) {
       go func() {
           ticker := time.NewTicker(interval)
           for range ticker.C {
               m.purgeOld(15 * time.Minute)
           }
       }()
   }
   ```
   Implement `purgeOld(maxAge time.Duration)` — it should remove spans older than `maxAge`. Remember to use `m.mu.Lock()` here.

7. **Wire `MemoryStore` into the Collector** — add a `StorageWriter` field to the `Collector` struct. In `processSpan`, call `c.storage.Write(span)`.

8. **Write table-driven tests** in `internal/storage/memory_test.go`:
   ```go
   func TestMemoryStore(t *testing.T) {
       tests := []struct {
           name    string
           spans   []*proto.Span
           queryID string
           wantLen int
           wantErr bool
       }{
           {
               name:    "returns spans for known trace",
               spans:   []*proto.Span{{TraceId: []byte("abc"), ServiceName: "svc"}},
               queryID: "abc",
               wantLen: 1,
           },
           {
               name:    "errors on unknown trace",
               queryID: "unknown",
               wantErr: true,
           },
       }
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // ... your test logic here
           })
       }
   }
   ```
   Table-driven tests are the Go standard. Get comfortable with this pattern.

9. **Run the race detector:** `go test -race ./...`
   The race detector finds concurrent map access bugs that don't always crash. If it reports a race condition, your mutex usage is wrong — fix it before moving on.
