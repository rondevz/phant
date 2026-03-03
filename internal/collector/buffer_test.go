package collector

import "testing"

func TestRingBuffer_DropsOldestWhenFull(t *testing.T) {
	buffer := NewRingBuffer(2)

	buffer.Add(Event{ID: "1"})
	buffer.Add(Event{ID: "2"})
	buffer.Add(Event{ID: "3"})

	events := buffer.Snapshot()
	if len(events) != 2 {
		t.Fatalf("buffer.Snapshot() len = %d, want %d", len(events), 2)
	}

	if events[0].ID != "2" || events[1].ID != "3" {
		t.Fatalf("buffer.Snapshot() IDs = [%s %s], want [2 3]", events[0].ID, events[1].ID)
	}

	if got := buffer.DroppedCount(); got != 1 {
		t.Fatalf("buffer.DroppedCount() = %d, want %d", got, 1)
	}
}
