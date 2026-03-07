package store

import (
	"sort"
	"sync"
	"time"
)

const (
	defaultRetention = 30 * time.Minute
	defaultCapacity  = 10000
)

type sample struct {
	startUnixNano int64
	durationNanos int64
}

type modelRing struct {
	samples []sample
	next    int
	full    bool
}

type LatencyStore struct {
	mu        sync.RWMutex
	retention time.Duration
	capacity  int
	models    map[string]*modelRing
}

type LatencySummary struct {
	SampleCount int
	P50         int64
	P95         int64
	P99         int64
	WindowStart time.Time
	WindowEnd   time.Time
}

func NewLatencyStore(retention time.Duration, capacity int) *LatencyStore {
	if retention <= 0 {
		retention = defaultRetention
	}
	if capacity <= 0 {
		capacity = defaultCapacity
	}

	return &LatencyStore{
		retention: retention,
		capacity:  capacity,
		models:    make(map[string]*modelRing),
	}
}

func (s *LatencyStore) Add(modelName string, startUnixNano, durationNanos int64) {
	if modelName == "" || startUnixNano <= 0 || durationNanos <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ring, ok := s.models[modelName]
	if !ok {
		ring = &modelRing{samples: make([]sample, s.capacity)}
		s.models[modelName] = ring
	}

	ring.samples[ring.next] = sample{startUnixNano: startUnixNano, durationNanos: durationNanos}
	ring.next = (ring.next + 1) % len(ring.samples)
	if ring.next == 0 {
		ring.full = true
	}
}

func (s *LatencyStore) Query(modelName string, window time.Duration, now time.Time) LatencySummary {
	if window <= 0 {
		window = 5 * time.Minute
	}
	if window > s.retention {
		window = s.retention
	}
	if now.IsZero() {
		now = time.Now()
	}

	windowStart := now.Add(-window)
	windowStartUnixNano := windowStart.UnixNano()
	windowEndUnixNano := now.UnixNano()

	s.mu.RLock()
	ring := s.models[modelName]
	if ring == nil {
		s.mu.RUnlock()
		return LatencySummary{WindowStart: windowStart, WindowEnd: now}
	}

	samples := s.snapshotRingLocked(ring)
	s.mu.RUnlock()

	durations := make([]int64, 0, len(samples))
	for _, smp := range samples {
		if smp.startUnixNano < windowStartUnixNano || smp.startUnixNano > windowEndUnixNano {
			continue
		}
		durations = append(durations, smp.durationNanos)
	}

	if len(durations) == 0 {
		return LatencySummary{WindowStart: windowStart, WindowEnd: now}
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	return LatencySummary{
		SampleCount: len(durations),
		P50:         percentileNearestRank(durations, 50),
		P95:         percentileNearestRank(durations, 95),
		P99:         percentileNearestRank(durations, 99),
		WindowStart: windowStart,
		WindowEnd:   now,
	}
}

func (s *LatencyStore) snapshotRingLocked(ring *modelRing) []sample {
	if !ring.full {
		out := make([]sample, 0, ring.next)
		for i := 0; i < ring.next; i++ {
			if ring.samples[i].startUnixNano > 0 {
				out = append(out, ring.samples[i])
			}
		}
		return out
	}

	out := make([]sample, 0, len(ring.samples))
	for i := range ring.samples {
		idx := (ring.next + i) % len(ring.samples)
		smp := ring.samples[idx]
		if smp.startUnixNano > 0 {
			out = append(out, smp)
		}
	}
	return out
}

func percentileNearestRank(sorted []int64, percentile int) int64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if percentile <= 0 {
		return sorted[0]
	}
	if percentile >= 100 {
		return sorted[n-1]
	}

	rank := (percentile*n + 99) / 100
	if rank < 1 {
		rank = 1
	}
	if rank > n {
		rank = n
	}
	return sorted[rank-1]
}
