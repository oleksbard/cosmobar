package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// hermeticConfig points COSMOBAR_CONFIG at a nonexistent file so renderFromJSON
// uses built-in defaults instead of the developer's real ~/.config config.
func hermeticConfig(t *testing.T) {
	t.Helper()
	t.Setenv("NO_COLOR", "1")
	t.Setenv("COSMOBAR_CONFIG", filepath.Join(t.TempDir(), "none.toml"))
}

func TestRenderFromInputProducesLine(t *testing.T) {
	hermeticConfig(t)
	json := `{"model":{"display_name":"Opus"},"workspace":{"current_dir":"/tmp/proj"},"context_window":{"used_percentage":10}}`
	out := renderFromJSON(strings.NewReader(json), 120)
	if !strings.Contains(out, "proj") || !strings.Contains(out, "Opus") {
		t.Errorf("output missing expected content: %q", out)
	}
}

func TestRenderFromInvalidJSONIsEmpty(t *testing.T) {
	hermeticConfig(t)
	if out := renderFromJSON(strings.NewReader("garbage"), 120); out != "" {
		t.Errorf("invalid JSON should render empty, got %q", out)
	}
}
