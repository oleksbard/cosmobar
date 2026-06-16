package cli

import (
	"strings"
	"testing"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/segments"
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

func TestGalleryHeadersShowThemeAndStyleOnly(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := renderGallery(previewOpts{cols: 120})
	// Each header names just the theme and style.
	for _, want := range []string{"coral · lean", "nord · tick", "catppuccin · blocks", "gruvbox · blocks"} {
		if !strings.Contains(out, want) {
			t.Errorf("gallery header missing %q:\n%s", want, out)
		}
	}
	// The internal preset label is not shown.
	for _, label := range []string{"Minimal", "Coder", "Pro/Max", "Everything"} {
		if strings.Contains(out, label) {
			t.Errorf("preset label %q should not appear in gallery output:\n%s", label, out)
		}
	}
	if !strings.Contains(out, "main") {
		t.Errorf("gallery presets include git; expected branch 'main':\n%s", out)
	}
}

func TestGalleryShowsNonDefaultSegments(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// The Pro/Max preset includes tokens and rate_limits, neither of which is
	// on by default — the gallery must surface them anyway.
	out := renderGallery(previewOpts{cols: 200})
	if !strings.Contains(out, "tok") {
		t.Errorf("gallery should render the tokens segment:\n%s", out)
	}
	if !strings.Contains(out, "5h") {
		t.Errorf("gallery should render the rate_limits segment:\n%s", out)
	}
}

func TestGalleryPresetSegmentsAreKnown(t *testing.T) {
	for _, p := range galleryPresets {
		for _, name := range p.order {
			if _, ok := segments.MetaByName(name); !ok {
				t.Errorf("preset %q references unknown segment %q", p.name, name)
			}
		}
	}
}

func TestGalleryEverythingCoversCatalog(t *testing.T) {
	var everything galleryPreset
	for _, p := range galleryPresets {
		if p.name == "Everything" {
			everything = p
		}
	}
	if len(everything.order) != len(segments.Catalog()) {
		t.Fatalf("Everything preset has %d segments, catalog has %d",
			len(everything.order), len(segments.Catalog()))
	}
	for i, m := range segments.Catalog() {
		if everything.order[i] != m.Name {
			t.Errorf("Everything[%d] = %q, want catalog order %q", i, everything.order[i], m.Name)
		}
	}
}

func TestGalleryListsEveryWidget(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := renderGallery(previewOpts{cols: 120})

	// A reference section renders each catalog widget on its own labeled line.
	if !strings.Contains(out, "all widgets") {
		t.Fatalf("gallery should include an 'all widgets' section:\n%s", out)
	}
	for _, m := range segments.Catalog() {
		if !strings.Contains(out, m.Name) {
			t.Errorf("widget reference should label %q:\n%s", m.Name, out)
		}
	}
	// Each widget actually renders with the mock data (spot-check a few).
	for _, want := range []string{"cosmobar", "Opus", "$0.42", "248k tok", "effort high"} {
		if !strings.Contains(out, want) {
			t.Errorf("widget reference missing rendered value %q:\n%s", want, out)
		}
	}
}

func TestWidgetCatalogUsesDefaultTheme(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out := renderWidgetCatalog(previewOpts{cols: 120})
	def := config.Default()
	// Header names the default theme/style and every catalog segment appears.
	if !strings.Contains(out, def.Theme) || !strings.Contains(out, def.Style) {
		t.Errorf("widget catalog header should name default theme/style %q/%q:\n%s", def.Theme, def.Style, out)
	}
	lines := strings.Count(out, "\n")
	if lines < len(segments.Catalog()) {
		t.Errorf("expected at least one line per widget (%d), got %d:\n%s", len(segments.Catalog()), lines, out)
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
