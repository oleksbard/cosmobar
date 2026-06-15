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

func TestAnimateFramesProgress(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// animationFrames returns a sequence of redraw strings ending at the final
	// settled line; the in-flight frames scramble, so the sequence holds several
	// distinct values (not just one repeated frame).
	frames := animationFrames(previewOpts{cols: 100}, 1)
	if len(frames) < 3 {
		t.Fatalf("want several frames, got %d", len(frames))
	}
	last := frames[len(frames)-1]
	if last == frames[0] {
		t.Error("first and last frame should differ (scramble then settle)")
	}
	distinct := map[string]bool{}
	for _, f := range frames {
		distinct[f] = true
	}
	if len(distinct) < 3 {
		t.Errorf("expected several distinct frames, got %d", len(distinct))
	}
}
