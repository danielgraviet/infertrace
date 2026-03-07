package store

import (
	"sync"
	"testing"
	"time"
)

func TestLatencyStore_ConcurrentAddAndQuery(t *testing.T) {
	s := NewLatencyStore(30*time.Minute, 5000)
	now := time.Now()

	var writers sync.WaitGroup
	for w := 0; w < 8; w++ {
		writers.Add(1)
		go func(worker int) {
			defer writers.Done()
			for i := 0; i < 1000; i++ {
				s.Add("gpt-4o-mini", now.Add(-time.Duration(i%300)*time.Second).UnixNano(), int64(100+(worker+i)%50))
			}
		}(w)
	}

	var readers sync.WaitGroup
	for r := 0; r < 8; r++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for i := 0; i < 400; i++ {
				summary := s.Query("gpt-4o-mini", 5*time.Minute, now)
				if summary.SampleCount < 0 {
					t.Fatalf("sample count should never be negative: %d", summary.SampleCount)
				}
			}
		}()
	}

	writers.Wait()
	readers.Wait()

	summary := s.Query("gpt-4o-mini", 5*time.Minute, now)
	if summary.SampleCount == 0 {
		t.Fatalf("expected non-empty summary after writes")
	}
	if summary.P50 <= 0 || summary.P95 <= 0 || summary.P99 <= 0 {
		t.Fatalf("expected positive quantiles, got p50=%d p95=%d p99=%d", summary.P50, summary.P95, summary.P99)
	}
}
