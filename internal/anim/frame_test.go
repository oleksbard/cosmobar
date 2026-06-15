package anim

import (
	"testing"

	"github.com/oleksbard/cosmobar/internal/render"
)

func TestFrameEndpoints(t *testing.T) {
	if got := Frame("Opus 4.8", 1.0, "decode", 123, false); got != "Opus 4.8" {
		t.Errorf("progress 1 = %q, want exact target", got)
	}
	if got := Frame("Opus 4.8", 1.5, "decode", 123, false); got != "Opus 4.8" {
		t.Errorf("progress >1 = %q, want exact target", got)
	}
	// progress 0: every eligible (non-space, single-width) cell differs from target.
	got := Frame("abc", 0.0, "decode", 123, false)
	if got == "abc" {
		t.Errorf("progress 0 should be fully scrambled, got %q", got)
	}
	if len([]rune(got)) != 3 {
		t.Errorf("width must be preserved: %q", got)
	}
}

func TestFramePreservesWidthAndSpaces(t *testing.T) {
	// Every variant has its own code path; the width contract holds for all.
	for _, variant := range []string{"decode", "glitch", "scatter"} {
		for _, p := range []float64{0, 0.3, 0.6, 0.9} {
			got := Frame("Opus 4.8", p, variant, 99, false)
			if render.Width(got) != render.Width("Opus 4.8") {
				t.Errorf("%s p=%v width %d != %d (%q)", variant, p, render.Width(got), render.Width("Opus 4.8"), got)
			}
			// the space at index 4 must remain a space
			if []rune(got)[4] != ' ' {
				t.Errorf("%s p=%v space not preserved: %q", variant, p, got)
			}
		}
	}
}

func TestFrameDecodeLocksLeftToRight(t *testing.T) {
	// At 60% of 10 eligible cells, the first 6 lock (final) and the rest glitch.
	got := []rune(Frame("ABCDEFGHIJ", 0.6, "decode", 7, false))
	want := []rune("ABCDEFGHIJ")
	for i := 0; i < 6; i++ {
		if got[i] != want[i] {
			t.Errorf("leading cell %d should be locked: %q", i, string(got))
		}
	}
	// A falsifying check: trailing cells must NOT all be final (i.e. still glitching).
	stillGlitching := false
	for i := 6; i < 10; i++ {
		if got[i] != want[i] {
			stillGlitching = true
		}
	}
	if !stillGlitching {
		t.Errorf("trailing cells should still be glitching at 60%%: %q", string(got))
	}
}

func TestFrameASCIIPaletteOnly(t *testing.T) {
	got := Frame("abcd", 0.0, "glitch", 5, true)
	for _, r := range got {
		if r > 127 {
			t.Errorf("ascii mode produced non-ascii %q in %q", r, got)
		}
	}
}

func TestFrameDeterministic(t *testing.T) {
	a := Frame("feat/anim", 0.4, "glitch", 42, false)
	b := Frame("feat/anim", 0.4, "glitch", 42, false)
	if a != b {
		t.Errorf("not deterministic: %q vs %q", a, b)
	}
}

func TestFrameScatterMidIsNotFinal(t *testing.T) {
	if got := Frame("12.34", 0.5, "scatter", 1, false); got == "12.34" {
		t.Errorf("scatter mid-progress should not be final: %q", got)
	}
	if got := Frame("12.34", 1.0, "scatter", 1, false); got != "12.34" {
		t.Errorf("scatter at progress 1 must be final, got %q", got)
	}
}
