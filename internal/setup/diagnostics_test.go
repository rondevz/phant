package setup

import (
	"context"
	"testing"
)

func TestCollectDiagnostics_AlwaysSetsGeneratedAt(t *testing.T) {
	report := CollectDiagnostics(context.Background())
	if report.GeneratedAt == "" {
		t.Fatalf("CollectDiagnostics() GeneratedAt = %q, want non-empty", report.GeneratedAt)
	}
}

func TestFirstLine(t *testing.T) {
	got := firstLine("line-1\nline-2")
	if got != "line-1" {
		t.Fatalf("firstLine() = %q, want %q", got, "line-1")
	}
}
