// Package settings reads and updates Claude Code's settings.json.
package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Path returns the user settings path (~/.claude/settings.json).
func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".claude/settings.json"
	}
	return filepath.Join(home, ".claude", "settings.json")
}

// WireStatusLine sets the statusLine block in the settings file at path,
// preserving all other keys. The existing file is copied to "<path>.bak"
// first. Missing files (and parent dirs) are created.
func WireStatusLine(path, binPath string, refreshInterval int) error {
	m := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &m) // start fresh if unparseable
		if err := os.WriteFile(path+".bak", data, 0o644); err != nil {
			return err
		}
	}
	m["statusLine"] = map[string]any{
		"type":            "command",
		"command":         binPath,
		"padding":         0,
		"refreshInterval": refreshInterval,
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// UnwireStatusLine removes the statusLine block from the settings file at path,
// preserving all other keys. The existing file is copied to "<path>.bak" first.
// Returns whether a statusLine key was present. A missing file is a no-op.
func UnwireStatusLine(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	m := map[string]any{}
	if err := json.Unmarshal(data, &m); err != nil {
		return false, err
	}
	if _, ok := m["statusLine"]; !ok {
		return false, nil
	}
	if err := os.WriteFile(path+".bak", data, 0o644); err != nil {
		return false, err
	}
	delete(m, "statusLine")
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, append(out, '\n'), 0o644)
}
