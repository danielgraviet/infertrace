package queryapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgraviet/infertrace/internal/store"
)

func TestLatencyEndpoint_EmptyData(t *testing.T) {
	latencyStore := store.NewLatencyStore(30*time.Minute, 100)
	srv := NewServer(latencyStore)
	now := time.Unix(200, 0).UTC()
	srv.now = func() time.Time { return now }

	req := httptest.NewRequest(http.MethodGet, "/models/gpt-4o-mini/latency?window=5m", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp latencyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.SampleCount != 0 || resp.P50 != 0 || resp.P95 != 0 || resp.P99 != 0 {
		t.Fatalf("unexpected non-empty response: %+v", resp)
	}
}

func TestLatencyEndpoint_SinglePoint(t *testing.T) {
	latencyStore := store.NewLatencyStore(30*time.Minute, 100)
	srv := NewServer(latencyStore)
	now := time.Unix(1_000, 0).UTC()
	srv.now = func() time.Time { return now }

	latencyStore.Add("gpt-4o-mini", now.Add(-2*time.Minute).UnixNano(), 120)

	req := httptest.NewRequest(http.MethodGet, "/models/gpt-4o-mini/latency?window=5m", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp latencyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.SampleCount != 1 || resp.P50 != 120 || resp.P95 != 120 || resp.P99 != 120 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLatencyEndpoint_NormalWindowQuery(t *testing.T) {
	latencyStore := store.NewLatencyStore(30*time.Minute, 100)
	srv := NewServer(latencyStore)
	now := time.Unix(2_000, 0).UTC()
	srv.now = func() time.Time { return now }

	latencyStore.Add("gpt-4o-mini", now.Add(-4*time.Minute).UnixNano(), 10)
	latencyStore.Add("gpt-4o-mini", now.Add(-3*time.Minute).UnixNano(), 20)
	latencyStore.Add("gpt-4o-mini", now.Add(-2*time.Minute).UnixNano(), 30)
	latencyStore.Add("gpt-4o-mini", now.Add(-1*time.Minute).UnixNano(), 40)

	req := httptest.NewRequest(http.MethodGet, "/models/gpt-4o-mini/latency?window=5m", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp latencyResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if resp.SampleCount != 4 || resp.P50 != 20 || resp.P95 != 40 || resp.P99 != 40 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
