package main

import (
	"strings"
	"testing"
)

func TestPreviewRender(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := previewRender(80, "nord", "")
	if !strings.Contains(out, "cosmobar") {
		t.Errorf("preview missing dir: %q", out)
	}
	if !strings.Contains(out, "main") {
		t.Errorf("preview should show a git branch (mock git status): %q", out)
	}
}
