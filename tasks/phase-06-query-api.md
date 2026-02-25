# Phase 6: Query API — HTTP Server in Go

> The collector ingests spans, but users need to retrieve them. This phase builds a REST API in Go so a frontend (or curl) can query traces. You'll learn how Go's standard library handles HTTP — it's more minimal than FastAPI but equally powerful once you understand the patterns.

---

## Learning Objectives

By the end of this phase you will understand:

- How `net/http` works in Go — `Handler`, `HandlerFunc`, `ServeMux`
- How to use the `chi` router for cleaner route definitions and URL parameters
- How JSON encoding/decoding works in Go (struct tags, `json.Marshal`, `json.NewDecoder`)
- How middleware works in Go (functions that wrap handlers)
- How to write integration tests for HTTP handlers using `httptest`
- How to return appropriate HTTP status codes and error responses

---

## Tasks

1. **Install the `chi` router:**
   ```bash
   go get github.com/go-chi/chi/v5
   ```
   Chi is a lightweight, idiomatic router. It's not a framework — it wraps `net/http` cleanly.

2. **Define your response types** in `internal/api/types.go`:
   ```go
   type SpanResponse struct {
       TraceID       string  `json:"trace_id"`
       SpanID        string  `json:"span_id"`
       ServiceName   string  `json:"service_name"`
       OperationName string  `json:"operation_name"`
       DurationMs    float64 `json:"duration_ms"`
       ModelName     string  `json:"model_name,omitempty"`
       BatchSize     int32   `json:"batch_size,omitempty"`
       Status        string  `json:"status"`
   }

   type TraceResponse struct {
       TraceID string          `json:"trace_id"`
       Spans   []*SpanResponse `json:"spans"`
   }

   type ErrorResponse struct {
       Error string `json:"error"`
   }
   ```
   The backtick struct tags (`json:"..."`) control how fields are serialized. `omitempty` skips the field if it's a zero value.

3. **Create an `APIServer` struct** in `internal/api/server.go`:
   ```go
   type APIServer struct {
       storage storage.StorageWriter
       router  *chi.Mux
   }

   func NewAPIServer(store storage.StorageWriter) *APIServer {
       s := &APIServer{storage: store}
       s.router = chi.NewRouter()
       s.registerRoutes()
       return s
   }
   ```

4. **Register routes** — add this `registerRoutes` method:
   ```go
   func (s *APIServer) registerRoutes() {
       s.router.Use(loggingMiddleware)         // Phase 6, Task 7
       s.router.Get("/health", s.handleHealth)
       s.router.Get("/traces/{traceID}", s.handleGetTrace)
       s.router.Get("/spans", s.handleListSpans)
   }
   ```

5. **Implement `handleGetTrace`:**
   ```go
   func (s *APIServer) handleGetTrace(w http.ResponseWriter, r *http.Request) {
       traceID := chi.URLParam(r, "traceID")
       spans, err := s.storage.GetTrace(traceID)
       if err != nil {
           writeJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
           return
       }
       writeJSON(w, http.StatusOK, TraceResponse{
           TraceID: traceID,
           Spans:   toSpanResponses(spans),
       })
   }
   ```

6. **Implement `handleListSpans`** — read query parameters to build a `SpanFilter`:
   ```go
   func (s *APIServer) handleListSpans(w http.ResponseWriter, r *http.Request) {
       filter := storage.SpanFilter{
           ServiceName: r.URL.Query().Get("service"),
           ModelName:   r.URL.Query().Get("model"),
           Limit:       100, // sensible default
       }
       // Parse since/until from query params (time.Parse with RFC3339 format)
       spans, err := s.storage.ListSpans(filter)
       // ...
   }
   ```

7. **Write a `loggingMiddleware`** — this teaches the middleware pattern:
   ```go
   func loggingMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           start := time.Now()
           next.ServeHTTP(w, r)
           log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
       })
   }
   ```
   Notice the pattern: middleware is a function that takes a `Handler` and returns a `Handler`.

8. **Add a `writeJSON` helper:**
   ```go
   func writeJSON(w http.ResponseWriter, status int, v any) {
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(status)
       json.NewEncoder(w).Encode(v)
   }
   ```

9. **Start the server in `cmd/collector/main.go`** alongside the gRPC server:
   ```go
   apiServer := api.NewAPIServer(store)
   go func() {
       log.Println("API listening on :8080")
       http.ListenAndServe(":8080", apiServer.Router())
   }()
   ```

10. **Write integration tests** using `httptest` in `internal/api/server_test.go`:
    ```go
    func TestGetTrace(t *testing.T) {
        store := storage.NewMemoryStore()
        store.Write(&proto.Span{TraceId: []byte("trace-1"), ServiceName: "svc"})

        server := NewAPIServer(store)
        req := httptest.NewRequest("GET", "/traces/trace-1", nil)
        rec := httptest.NewRecorder()
        server.Router().ServeHTTP(rec, req)

        if rec.Code != http.StatusOK {
            t.Fatalf("expected 200, got %d", rec.Code)
        }
    }
    ```
    `httptest` lets you test HTTP handlers without starting a real server. This is the standard Go approach.

11. **Test manually with curl:**
    ```bash
    # Health check
    curl http://localhost:8080/health

    # Get a trace (use a real trace ID from your test client)
    curl http://localhost:8080/traces/YOUR-TRACE-ID

    # List spans for a model
    curl "http://localhost:8080/spans?model=gpt-4"
    ```
