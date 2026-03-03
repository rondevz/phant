package collector

import "sync"

type RingBuffer struct {
	mu      sync.RWMutex
	events  []Event
	start   int
	size    int
	dropped uint64
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity < 1 {
		capacity = 1
	}

	return &RingBuffer{
		events: make([]Event, capacity),
	}
}

func (b *RingBuffer) Add(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.size < len(b.events) {
		idx := (b.start + b.size) % len(b.events)
		b.events[idx] = event
		b.size++
		return
	}

	b.events[b.start] = event
	b.start = (b.start + 1) % len(b.events)
	b.dropped++
}

func (b *RingBuffer) Snapshot() []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]Event, b.size)
	for i := 0; i < b.size; i++ {
		idx := (b.start + i) % len(b.events)
		result[i] = b.events[idx]
	}
	return result
}

func (b *RingBuffer) DroppedCount() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.dropped
}
