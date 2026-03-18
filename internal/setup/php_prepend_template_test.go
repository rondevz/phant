package setup

import (
	"strings"
	"testing"
)

func TestPHPPrependTemplate_EmitsTraceCallsite(t *testing.T) {
	if !strings.Contains(phpPrependTemplate, "function phant_trace_callsite(): array") {
		t.Fatalf("phpPrependTemplate missing phant_trace_callsite helper")
	}

	if !strings.Contains(phpPrependTemplate, "'trace' => phant_trace_callsite()") {
		t.Fatalf("phpPrependTemplate should emit trace callsite data")
	}

	if !strings.Contains(phpPrependTemplate, "if ($file === __FILE__) {") {
		t.Fatalf("phpPrependTemplate should ignore prepend file frames")
	}

	if !strings.Contains(phpPrependTemplate, "if (str_contains($file, '/vendor/symfony/var-dumper/')) {") {
		t.Fatalf("phpPrependTemplate should ignore symfony var-dumper frames")
	}

	if !strings.Contains(phpPrependTemplate, "if (str_contains($function, 'phant_install_vardumper_handler')) {") {
		t.Fatalf("phpPrependTemplate should ignore internal handler closure frames")
	}
}

func TestPHPPrependTemplate_NormalizesObjectMetadata(t *testing.T) {
	if !strings.Contains(phpPrependTemplate, "function phant_normalize_value($value, int $depth = 0, array &$seen = [])") {
		t.Fatalf("phpPrependTemplate missing phant_normalize_value helper")
	}

	if !strings.Contains(phpPrependTemplate, "'__phantType' => 'object'") {
		t.Fatalf("phpPrependTemplate should emit object metadata payload")
	}

	if !strings.Contains(phpPrependTemplate, "'__className' => $className") {
		t.Fatalf("phpPrependTemplate should include object class name")
	}

	if !strings.Contains(phpPrependTemplate, "'__objectId' => $objectID") {
		t.Fatalf("phpPrependTemplate should include object id")
	}
}
