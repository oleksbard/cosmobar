package main

import (
	"strings"
	"testing"
)

func TestThemesList(t *testing.T) {
	out := themesList()
	for _, name := range []string{"coral", "catppuccin", "nord", "gruvbox"} {
		if !strings.Contains(out, name) {
			t.Errorf("themes list missing %q: %q", name, out)
		}
	}
}
