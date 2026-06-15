package statusline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oleksbard/cosmobar/internal/anim"
	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/segments"
	"github.com/oleksbard/cosmobar/internal/session"
)

// animDemoSession returns an always-enabled demo session starting at fixedNow,
// so a render at fixedNow+Δ shows mid-animation frames.
func animDemoSession(cfg config.Config) *anim.Session {
	return anim.Demo(cfg, render.ProfileTrueColor, fixedNow)
}

func load(t *testing.T, name string) *session.Session {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	s, err := session.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

var fixedNow = time.Date(2026, 6, 14, 14, 32, 0, 0, time.UTC)

func TestRenderSingleLine(t *testing.T) {
	in := Input{
		Session: load(t, "heavy.json"),
		Git:     git.Status{InRepo: true, Branch: "main", Staged: 1, Modified: 2},
		Config:  config.Default(),
		Cols:    120,
		Profile: render.ProfileNone,
		Now:     fixedNow,
	}
	got := Render(in)
	want := "cosmobar · main +1 ~2 · Opus · ▓▓▓░░░░░ 42% · $0.12 · 14:32"
	if got != want {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

func TestRenderWrapsTwoRows(t *testing.T) {
	in := Input{
		Session: load(t, "heavy.json"),
		Git:     git.Status{InRepo: true, Branch: "main", Staged: 1, Modified: 2},
		Config:  config.Default(),
		Cols:    40,
		Profile: render.ProfileNone,
		Now:     fixedNow,
	}
	got := Render(in)
	want := "cosmobar · main +1 ~2 · Opus\n▓▓▓░░░░░ 42% · $0.12 · 14:32"
	if got != want {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

func TestRenderHidesGitWhenNoRepo(t *testing.T) {
	in := Input{
		Session: load(t, "minimal.json"),
		Git:     git.Status{InRepo: false},
		Config:  config.Default(),
		Cols:    120,
		Profile: render.ProfileNone,
		Now:     fixedNow,
	}
	got := Render(in)
	// no git, no context (used_percentage nil)
	want := "scratch · Sonnet · $0.00 · 14:32"
	if got != want {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

func TestRenderColorizesWithProfile(t *testing.T) {
	in := Input{
		Session: load(t, "minimal.json"),
		Config:  config.Default(),
		Cols:    120,
		Profile: render.ProfileTrueColor,
		Now:     fixedNow,
	}
	got := Render(in)
	if !contains(got, "\x1b[38;2;") {
		t.Errorf("expected ANSI color codes, got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

func TestRenderAnimatesChangedSegment(t *testing.T) {
	cfg := config.Default() // animation enabled by default
	// A demo session animates every segment from fixedNow. Partway through, the
	// model segment text should be scrambled (no longer the literal "Opus").
	sess := anim.Demo(cfg, render.ProfileTrueColor, fixedNow)
	in := Input{
		Session: load(t, "heavy.json"),
		Git:     git.Status{InRepo: true, Branch: "main", Staged: 1, Modified: 2},
		Config:  cfg,
		Cols:    120,
		Profile: render.ProfileTrueColor,
		Now:     fixedNow.Add(200 * time.Millisecond), // mid-animation
		Anim:    sess,
	}
	got := Render(in)
	// Safe (non-flaky): at ~29% only 1 of 4 cells is locked, and the glitch
	// palette contains no letters, so "Opus" cannot survive even partially.
	if contains(got, "Opus") {
		t.Errorf("expected model to be scrambled mid-animation, got %q", got)
	}
}

func TestRenderAnimationPreservesWidth(t *testing.T) {
	cfg := config.Default()
	base := Input{
		Session: load(t, "heavy.json"),
		Git:     git.Status{InRepo: true, Branch: "main", Staged: 1, Modified: 2},
		Config:  cfg,
		Cols:    120,
		Profile: render.ProfileNone, // plain text → render.Width is exact
		Now:     fixedNow,
	}
	baseline := Render(base)

	animIn := base
	animIn.Now = fixedNow.Add(200 * time.Millisecond)
	animIn.Anim = animDemoSession(cfg)
	got := Render(animIn)

	if render.Width(got) != render.Width(baseline) {
		t.Errorf("animation changed width:\n base %d: %q\n anim %d: %q",
			render.Width(baseline), baseline, render.Width(got), got)
	}
}

func TestAnimateSegmentPreservesParts(t *testing.T) {
	// A multi-part segment must keep its part count and per-part State; only the
	// text scrambles, at the same per-part width.
	seg := segments.Segment{
		Name:  "lines",
		Parts: []segments.Part{{Text: "+12", State: render.StateOK}, {Text: "-3", State: render.StateCrit}},
	}
	animateSegment(animDemoSession(config.Default()), &seg, fixedNow.Add(200*time.Millisecond))
	if len(seg.Parts) != 2 {
		t.Fatalf("part count changed: %d", len(seg.Parts))
	}
	if seg.Parts[0].State != render.StateOK || seg.Parts[1].State != render.StateCrit {
		t.Errorf("per-part State not preserved: %+v", seg.Parts)
	}
	if render.Width(seg.Parts[0].Text) != 3 || render.Width(seg.Parts[1].Text) != 2 {
		t.Errorf("per-part width not preserved: %+v", seg.Parts)
	}
}

func TestRenderNilAnimUnchanged(t *testing.T) {
	// With no Anim session the output must match the existing baseline exactly.
	in := Input{
		Session: load(t, "heavy.json"),
		Git:     git.Status{InRepo: true, Branch: "main", Staged: 1, Modified: 2},
		Config:  config.Default(),
		Cols:    120,
		Profile: render.ProfileNone,
		Now:     fixedNow,
	}
	want := "cosmobar · main +1 ~2 · Opus · ▓▓▓░░░░░ 42% · $0.12 · 14:32"
	if got := Render(in); got != want {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}
