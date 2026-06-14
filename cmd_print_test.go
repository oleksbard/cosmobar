package main

import (
	"strings"
	"testing"
)

func TestRenderFromInputProducesLine(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	json := `{"model":{"display_name":"Opus"},"workspace":{"current_dir":"/tmp/proj"},"context_window":{"used_percentage":10}}`
	out := renderFromJSON(strings.NewReader(json), 120)
	if !strings.Contains(out, "proj") || !strings.Contains(out, "Opus") {
		t.Errorf("output missing expected content: %q", out)
	}
}

func TestRenderFromInvalidJSONIsEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if out := renderFromJSON(strings.NewReader("garbage"), 120); out != "" {
		t.Errorf("invalid JSON should render empty, got %q", out)
	}
}
