package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	c := Default()
	if c.Theme != "coral" {
		t.Errorf("theme = %q", c.Theme)
	}
	if len(c.Order) != 6 || c.Order[0] != "dir" {
		t.Errorf("order = %v", c.Order)
	}
	lo, hi := c.Thresholds()
	if lo != 70 || hi != 90 {
		t.Errorf("thresholds = %d,%d", lo, hi)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	c, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Theme != "coral" {
		t.Errorf("expected defaults, got theme %q", c.Theme)
	}
}

func TestLoadOverridesOnlyPresentKeys(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	os.WriteFile(p, []byte("theme = \"nord\"\nmax_rows = 1\n[clock]\nformat = \"12h\"\n"), 0o644)
	c, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Theme != "nord" {
		t.Errorf("theme = %q, want nord", c.Theme)
	}
	if c.MaxRows != 1 {
		t.Errorf("max_rows = %d, want 1", c.MaxRows)
	}
	if c.Clock.Format != "12h" {
		t.Errorf("clock.format = %q", c.Clock.Format)
	}
	// untouched keys keep defaults
	if c.Separator != " · " {
		t.Errorf("separator should keep default, got %q", c.Separator)
	}
	if len(c.Order) != 6 {
		t.Errorf("order should keep default, got %v", c.Order)
	}
}

func TestLoadInvalidTOMLReturnsError(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	os.WriteFile(p, []byte("theme = = ="), 0o644)
	if _, err := Load(p); err == nil {
		t.Errorf("expected parse error")
	}
}
