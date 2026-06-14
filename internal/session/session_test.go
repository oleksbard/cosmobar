package session

import (
	"strings"
	"testing"
)

const heavyJSON = `{
  "session_id": "abc-123",
  "cwd": "/Users/me/projects/cosmobar",
  "model": {"id": "claude-opus-4-8", "display_name": "Opus"},
  "workspace": {"current_dir": "/Users/me/projects/cosmobar", "project_dir": "/Users/me/projects/cosmobar", "repo": {"host": "github.com", "owner": "oleksbard", "name": "cosmobar"}},
  "output_style": {"name": "default"},
  "cost": {"total_cost_usd": 0.12, "total_duration_ms": 723000, "total_lines_added": 156, "total_lines_removed": 23},
  "context_window": {"used_percentage": 42, "context_window_size": 200000},
  "rate_limits": {"five_hour": {"used_percentage": 23.5, "resets_at": 1738425600}, "seven_day": {"used_percentage": 41.2, "resets_at": 1738857600}},
  "effort": {"level": "high"}
}`

func TestParseHeavy(t *testing.T) {
	s, err := Parse(strings.NewReader(heavyJSON))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if s.Model.DisplayName != "Opus" {
		t.Errorf("model = %q, want Opus", s.Model.DisplayName)
	}
	if s.Dir() != "/Users/me/projects/cosmobar" {
		t.Errorf("Dir() = %q", s.Dir())
	}
	if s.ContextWindow.UsedPercentage == nil || *s.ContextWindow.UsedPercentage != 42 {
		t.Errorf("used_percentage not parsed")
	}
	if s.RateLimits == nil || s.RateLimits.FiveHour == nil || s.RateLimits.FiveHour.UsedPercentage != 23.5 {
		t.Errorf("rate_limits not parsed")
	}
	if s.Effort == nil || s.Effort.Level != "high" {
		t.Errorf("effort not parsed")
	}
}

func TestParseMinimalAbsentFields(t *testing.T) {
	s, err := Parse(strings.NewReader(`{"model":{"display_name":"Sonnet"},"cwd":"/tmp"}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if s.RateLimits != nil {
		t.Errorf("rate_limits should be nil when absent")
	}
	if s.ContextWindow.UsedPercentage != nil {
		t.Errorf("used_percentage should be nil when absent")
	}
	if s.Effort != nil {
		t.Errorf("effort should be nil when absent")
	}
	if s.Dir() != "/tmp" {
		t.Errorf("Dir() fallback to cwd failed: %q", s.Dir())
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse(strings.NewReader("not json")); err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}
