package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderTOMLRoundTrips(t *testing.T) {
	c := Default()
	c.Theme = "nord"
	c.Order = []string{"dir", "model", "context"}
	c.Clock.Format = "12h"
	c.Glyphs = "ascii"
	c.RateLimits.Show = true
	c.Style = "blocks"
	c.BlockCaps = "square"
	c.RateLimits.Window = "5h"

	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(p, []byte(RenderTOML(c)), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Load(p)
	if err != nil {
		t.Fatalf("Load after RenderTOML: %v", err)
	}
	if got.Theme != "nord" {
		t.Errorf("theme = %q", got.Theme)
	}
	if len(got.Order) != 3 || got.Order[2] != "context" {
		t.Errorf("order = %v", got.Order)
	}
	if got.Clock.Format != "12h" {
		t.Errorf("clock = %q", got.Clock.Format)
	}
	if got.Glyphs != "ascii" {
		t.Errorf("glyphs = %q", got.Glyphs)
	}
	if !got.RateLimits.Show {
		t.Error("rate_limits.show should round-trip as true")
	}
	if got.Style != "blocks" || got.BlockCaps != "square" {
		t.Errorf("style/caps round-trip: %q/%q", got.Style, got.BlockCaps)
	}
	if got.RateLimits.Window != "5h" {
		t.Errorf("rate window round-trip: %q", got.RateLimits.Window)
	}
}
