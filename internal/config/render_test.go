package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestRenderTOMLIncludesCostAndBlockCost(t *testing.T) {
	c := Default()
	c.Cost.Rollups = []string{"today", "month"}
	c.RateLimits.ShowBlockCost = true
	out := RenderTOML(c)

	if !strings.Contains(out, `[cost]`) || !strings.Contains(out, `rollups = ["today", "month"]`) {
		t.Errorf("RenderTOML missing [cost] rollups; got:\n%s", out)
	}
	if !strings.Contains(out, "show_block_cost = true") {
		t.Errorf("RenderTOML missing show_block_cost; got:\n%s", out)
	}
}

func TestRenderTOMLCostRoundTrips(t *testing.T) {
	c := Default()
	c.Cost.Rollups = []string{"today"}
	tmp := t.TempDir() + "/config.toml"
	if err := os.WriteFile(tmp, []byte(RenderTOML(c)), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cost.Rollups) != 1 || got.Cost.Rollups[0] != "today" {
		t.Errorf("round-tripped rollups = %v, want [today]", got.Cost.Rollups)
	}
	if !got.RateLimits.ShowBlockCost {
		t.Error("round-tripped ShowBlockCost = false, want true")
	}
}
