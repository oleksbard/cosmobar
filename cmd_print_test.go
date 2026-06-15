package main

import (
	"os"
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

func TestRenderFromJSONRunsWithAnimation(t *testing.T) {
	hermeticConfig(t)
	// anim.Load persists state to $TMPDIR/cosmobar-anim-<sanitized id>; remove it
	// before and after so the test is hermetic across runs.
	statePath := filepath.Join(os.TempDir(), "cosmobar-anim-anim-smoke")
	os.Remove(statePath)
	defer os.Remove(statePath)

	// Smoke test: two consecutive renders of the same stdin exercise the
	// load-then-save round trip across invocations without panicking.
	in := `{"session_id":"anim-smoke","model":{"display_name":"Opus 4.8"},"cost":{"total_cost_usd":0.12}}`
	out1 := renderFromJSON(strings.NewReader(in), 120)
	out2 := renderFromJSON(strings.NewReader(in), 120)
	if out1 == "" || out2 == "" {
		t.Fatalf("empty output: %q / %q", out1, out2)
	}
}
