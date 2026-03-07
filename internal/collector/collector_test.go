package collector

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestCollector_QueueFullDrops(t *testing.T) {
	c := NewCollector(0, 1, nil)

	if err := c.Enqueue(Span{ServiceName: "svc"}); err != nil {
		t.Fatalf("first enqueue failed: %v", err)
	}
	if err := c.Enqueue(Span{ServiceName: "svc"}); err != ErrQueueFull {
		t.Fatalf("second enqueue error = %v, want %v", err, ErrQueueFull)
	}

	stats := c.Stats()
	if got, want := stats.Accepted, int64(1); got != want {
		t.Fatalf("accepted = %d, want %d", got, want)
	}
	if got, want := stats.Dropped, int64(1); got != want {
		t.Fatalf("dropped = %d, want %d", got, want)
	}
}

func TestCollector_StopDrainsQueueGracefully(t *testing.T) {
	var processed atomic.Int64
	c := NewCollector(2, 4, func(Span) {
		time.Sleep(5 * time.Millisecond)
		processed.Add(1)
	})

	for i := range 4 {
		if err := c.Enqueue(Span{ServiceName: "svc", SpanID: string(rune('a' + i))}); err != nil {
			t.Fatalf("enqueue %d failed: %v", i, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if got, want := processed.Load(), int64(4); got != want {
		t.Fatalf("processed = %d, want %d", got, want)
	}
}

func TestCollector_StopRejectsNewEnqueue(t *testing.T) {
	c := NewCollector(1, 2, nil)
	if err := c.Stop(context.Background()); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if err := c.Enqueue(Span{ServiceName: "svc"}); err != ErrCollectorClosed {
		t.Fatalf("enqueue after stop = %v, want %v", err, ErrCollectorClosed)
	}
}
