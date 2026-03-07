package collector

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrQueueFull       = errors.New("collector queue is full")
	ErrCollectorClosed = errors.New("collector is stopped")
)

type Stats struct {
	Accepted int64
	Dropped  int64
	Invalid  int64
}

type Collector struct {
	queue    chan Span
	workerWg sync.WaitGroup
	stopOnce sync.Once
	closed   atomic.Bool

	accepted atomic.Int64
	dropped  atomic.Int64
	invalid  atomic.Int64

	sink func(Span)
}

func NewCollector(workerCount, queueSize int, sink func(Span)) *Collector {
	if queueSize <= 0 {
		queueSize = 1
	}
	if sink == nil {
		sink = func(Span) {}
	}

	c := &Collector{
		queue: make(chan Span, queueSize),
		sink:  sink,
	}

	for range workerCount {
		c.workerWg.Add(1)
		go c.worker()
	}

	return c
}

func (c *Collector) worker() {
	defer c.workerWg.Done()
	for span := range c.queue {
		c.sink(span)
	}
}

// Enqueue applies MVP backpressure policy: do not block callers when queue is full.
// Instead, drop new work quickly and return ErrQueueFull so ingestion latency stays bounded.
func (c *Collector) Enqueue(span Span) error {
	if c.closed.Load() {
		c.dropped.Add(1)
		return ErrCollectorClosed
	}

	select {
	case c.queue <- span:
		c.accepted.Add(1)
		return nil
	default:
		c.dropped.Add(1)
		return ErrQueueFull
	}
}

func (c *Collector) AddInvalid(n int) {
	if n <= 0 {
		return
	}
	c.invalid.Add(int64(n))
}

func (c *Collector) Stats() Stats {
	return Stats{
		Accepted: c.accepted.Load(),
		Dropped:  c.dropped.Load(),
		Invalid:  c.invalid.Load(),
	}
}

func (c *Collector) Stop(ctx context.Context) error {
	c.stopOnce.Do(func() {
		c.closed.Store(true)
		close(c.queue)
	})

	done := make(chan struct{})
	go func() {
		c.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
