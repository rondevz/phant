package collector

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestServer_IngestsAndBroadcastsEvents(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "collector.sock")
	server := NewServer(socketPath, 4)
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start() error = %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	subID, ch := server.Subscribe(2)
	defer server.Unsubscribe(subID)

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("net.Dial(unix, %q) error = %v", socketPath, err)
	}
	defer conn.Close()

	if _, err := fmt.Fprintln(conn, validCLIEventLine("evt-1")); err != nil {
		t.Fatalf("write event line error = %v", err)
	}

	select {
	case got := <-ch:
		if got.ID != "evt-1" {
			t.Fatalf("subscriber event ID = %q, want %q", got.ID, "evt-1")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for broadcast event")
	}

	events := server.Events()
	if len(events) != 1 {
		t.Fatalf("server.Events() len = %d, want %d", len(events), 1)
	}
	if events[0].ID != "evt-1" {
		t.Fatalf("server.Events()[0].ID = %q, want %q", events[0].ID, "evt-1")
	}
}

func TestServer_RingBufferTracksDropped(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "collector.sock")
	server := NewServer(socketPath, 2)
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start() error = %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("net.Dial(unix, %q) error = %v", socketPath, err)
	}
	defer conn.Close()

	for _, id := range []string{"evt-1", "evt-2", "evt-3"} {
		if _, err := fmt.Fprintln(conn, validCLIEventLine(id)); err != nil {
			t.Fatalf("write event line for %q error = %v", id, err)
		}
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(server.Events()) == 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	events := server.Events()
	if len(events) != 2 {
		t.Fatalf("server.Events() len = %d, want %d", len(events), 2)
	}
	if events[0].ID != "evt-2" || events[1].ID != "evt-3" {
		t.Fatalf("server.Events() IDs = [%s %s], want [evt-2 evt-3]", events[0].ID, events[1].ID)
	}
	if got := server.DroppedCount(); got != 1 {
		t.Fatalf("server.DroppedCount() = %d, want %d", got, 1)
	}
}

func TestServer_IgnoresInvalidLines(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "collector.sock")
	server := NewServer(socketPath, 2)
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start() error = %v", err)
	}
	defer func() {
		_ = server.Stop()
	}()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("net.Dial(unix, %q) error = %v", socketPath, err)
	}
	defer conn.Close()

	if _, err := fmt.Fprintln(conn, "{"); err != nil {
		t.Fatalf("write invalid line error = %v", err)
	}
	if _, err := fmt.Fprintln(conn, ""); err != nil {
		t.Fatalf("write empty line error = %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if got := len(server.Events()); got != 0 {
		t.Fatalf("server.Events() len = %d, want %d", got, 0)
	}
}

func validCLIEventLine(id string) string {
	return fmt.Sprintf(`{"schemaVersion":1,"id":"%s","timestamp":"2026-03-02T12:00:00Z","sourceType":"cli","projectRoot":"/tmp/app","phpSapi":"cli","requestId":null,"command":{"name":"artisan"},"isDd":false,"payloadFormat":"json","payload":{"ok":true},"trace":[],"host":{"hostname":"test-host","pid":1234}}`, id)
}
