package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/spend"
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

// tokensConfig writes a config enabling the tokens segment and points
// COSMOBAR_CONFIG at it, so renderFromJSON exercises the conditional gather.
func tokensConfig(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("NO_COLOR", "1")
	cfgPath := filepath.Join(dir, "cosmobar.toml")
	if err := os.WriteFile(cfgPath, []byte(`order = ["dir", "tokens"]`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COSMOBAR_CONFIG", cfgPath)
}

// oneTurnTranscript writes a transcript with a single assistant turn whose
// input+output totals 12000 tokens, which humanTokens renders as "12k".
func oneTurnTranscript(t *testing.T, dir string) string {
	t.Helper()
	p := filepath.Join(dir, "t.jsonl")
	line := `{"type":"assistant","requestId":"r1","message":{"id":"m1","usage":{"input_tokens":12000,"output_tokens":0}}}` + "\n"
	if err := os.WriteFile(p, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRenderFromJSONShowsTokensWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	tokensConfig(t, dir)
	tr := oneTurnTranscript(t, dir)
	json := `{"workspace":{"current_dir":"/tmp/proj"},"transcript_path":"` + tr + `"}`
	out := renderFromJSON(strings.NewReader(json), 120)
	if !strings.Contains(out, "12k tok") {
		t.Errorf("expected %q in output, got %q", "12k tok", out)
	}
}

func TestRenderFromJSONOmitsTokensWhenNotInOrder(t *testing.T) {
	hermeticConfig(t) // default order has no "tokens"
	dir := t.TempDir()
	tr := oneTurnTranscript(t, dir)
	json := `{"workspace":{"current_dir":"/tmp/proj"},"transcript_path":"` + tr + `"}`
	out := renderFromJSON(strings.NewReader(json), 120)
	if strings.Contains(out, "tok") {
		t.Errorf("tokens segment should be absent when not in order, got %q", out)
	}
}

func TestRenderFromJSONHidesTokensWhenTranscriptUnreadable(t *testing.T) {
	dir := t.TempDir()
	tokensConfig(t, dir)
	json := `{"workspace":{"current_dir":"/tmp/proj"},"transcript_path":"` + filepath.Join(dir, "nope.jsonl") + `"}`
	out := renderFromJSON(strings.NewReader(json), 120)
	if strings.Contains(out, "tok") {
		t.Errorf("tokens segment should hide when transcript unreadable, got %q", out)
	}
	if !strings.Contains(out, "proj") {
		t.Errorf("dir segment should still render, got %q", out)
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

func TestNeedSpend(t *testing.T) {
	cfg := config.Default()
	cfg.Order = []string{"dir", "model", "clock"}
	if needSpend(cfg) {
		t.Error("needSpend should be false without cost/rate_limits")
	}
	cfg.Order = []string{"dir", "cost"}
	if !needSpend(cfg) {
		t.Error("needSpend should be true when cost is present")
	}
	cfg.Order = []string{"rate_limits"}
	if !needSpend(cfg) {
		t.Error("needSpend should be true when rate_limits is present")
	}
}

func TestRenderFromJSONShowsTodayRollup(t *testing.T) {
	// Fully isolate the ledger from the developer's real cache: os.UserCacheDir
	// uses $HOME/Library/Caches on macOS and $XDG_CACHE_HOME (or $HOME/.cache)
	// on Linux, so redirect both. hermeticConfig pins the default config (whose
	// order includes "cost", so needSpend is true).
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, "cache"))
	hermeticConfig(t)
	// Remove any stale anim state so the single render is never a scramble frame.
	statePath := filepath.Join(os.TempDir(), "cosmobar-anim-test-rollup-sess")
	os.Remove(statePath)
	defer os.Remove(statePath)

	// Seed a $3.00 baseline for the session, then render at $5.00 cumulative.
	// today is the delta the binary observed — $2.00 — not the lifetime total.
	l := spend.Load(time.Now())
	l.Upsert("test-rollup-sess", 3.00, 0)
	l.Save()

	blob := `{"session_id":"test-rollup-sess","cost":{"total_cost_usd":5.00},"workspace":{"current_dir":"/tmp"}}`
	out := renderFromJSON(strings.NewReader(blob), 200)
	if !strings.Contains(out, "$2.00 today") {
		t.Errorf("expected '$2.00 today' rollup, got %q", out)
	}
}
