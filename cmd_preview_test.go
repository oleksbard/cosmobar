package main

import (
	"strings"
	"testing"
)

func TestPreviewRender(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := previewRender(previewOpts{cols: 80, theme: "nord"})
	if !strings.Contains(out, "cosmobar") {
		t.Errorf("preview missing dir: %q", out)
	}
	if !strings.Contains(out, "main") {
		t.Errorf("preview should show a git branch (mock git status): %q", out)
	}
}

func TestPreviewOverrides(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := previewRender(previewOpts{cols: 120, theme: "coral", style: "lean", order: "git, model, lines"})
	if !strings.Contains(out, "main") {
		t.Errorf("preview --order should still render git: %q", out)
	}
	if !strings.Contains(out, "+24") || !strings.Contains(out, "-7") {
		t.Errorf("preview mock should show lines changes: %q", out)
	}
}
