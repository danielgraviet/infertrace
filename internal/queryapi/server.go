package queryapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgraviet/infertrace/internal/store"
)

type Server struct {
	store *store.LatencyStore
	now   func() time.Time
}

type latencyResponse struct {
	SampleCount int    `json:"sample_count"`
	P50         int64  `json:"p50"`
	P95         int64  `json:"p95"`
	P99         int64  `json:"p99"`
	WindowStart string `json:"window_start"`
	WindowEnd   string `json:"window_end"`
}

func NewServer(latencyStore *store.LatencyStore) *Server {
	return &Server{
		store: latencyStore,
		now:   time.Now,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/models/", s.handleModelLatency)
	return mux
}

func (s *Server) handleModelLatency(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	modelName, ok := parseModelPath(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	window := 5 * time.Minute
	if raw := r.URL.Query().Get("window"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid window", http.StatusBadRequest)
			return
		}
		window = parsed
	}

	summary := s.store.Query(modelName, window, s.now())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(latencyResponse{
		SampleCount: summary.SampleCount,
		P50:         summary.P50,
		P95:         summary.P95,
		P99:         summary.P99,
		WindowStart: summary.WindowStart.Format(time.RFC3339Nano),
		WindowEnd:   summary.WindowEnd.Format(time.RFC3339Nano),
	})
}

func parseModelPath(path string) (string, bool) {
	trimmed := strings.TrimPrefix(path, "/models/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "latency" {
		return "", false
	}
	decoded, err := url.PathUnescape(parts[0])
	if err != nil || decoded == "" {
		return "", false
	}
	return decoded, true
}
