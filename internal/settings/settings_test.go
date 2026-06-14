package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWireStatusLineCreatesFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	if err := WireStatusLine(p, "/usr/local/bin/cosmobar", 10); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	data, _ := os.ReadFile(p)
	json.Unmarshal(data, &m)
	sl, ok := m["statusLine"].(map[string]any)
	if !ok {
		t.Fatalf("statusLine missing: %v", m)
	}
	if sl["command"] != "/usr/local/bin/cosmobar" {
		t.Errorf("command = %v", sl["command"])
	}
	if sl["type"] != "command" {
		t.Errorf("type = %v", sl["type"])
	}
}

func TestWireStatusLinePreservesAndBacksUp(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	os.WriteFile(p, []byte(`{"model":"opus","theme":"dark"}`), 0o644)
	if err := WireStatusLine(p, "/bin/cosmobar", 5); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	data, _ := os.ReadFile(p)
	json.Unmarshal(data, &m)
	if m["theme"] != "dark" {
		t.Errorf("existing keys should be preserved, got %v", m)
	}
	if _, err := os.Stat(p + ".bak"); err != nil {
		t.Errorf("backup should exist: %v", err)
	}
}

func TestUnwireStatusLineRemovesAndPreserves(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	os.WriteFile(p, []byte(`{"theme":"dark","statusLine":{"type":"command","command":"x"}}`), 0o644)
	removed, err := UnwireStatusLine(p)
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Error("should report removed=true")
	}
	var m map[string]any
	data, _ := os.ReadFile(p)
	json.Unmarshal(data, &m)
	if _, ok := m["statusLine"]; ok {
		t.Error("statusLine should be removed")
	}
	if m["theme"] != "dark" {
		t.Errorf("other keys must be preserved, got %v", m)
	}
	if _, err := os.Stat(p + ".bak"); err != nil {
		t.Errorf("backup should exist: %v", err)
	}
}

func TestUnwireStatusLineNoEntry(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	os.WriteFile(p, []byte(`{"theme":"dark"}`), 0o644)
	removed, err := UnwireStatusLine(p)
	if err != nil {
		t.Fatal(err)
	}
	if removed {
		t.Error("should report removed=false when no statusLine present")
	}
}

func TestUnwireStatusLineMissingFile(t *testing.T) {
	removed, err := UnwireStatusLine(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("missing file should be a no-op, got %v", err)
	}
	if removed {
		t.Error("missing file should report removed=false")
	}
}
