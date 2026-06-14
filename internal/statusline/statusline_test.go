package statusline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
)

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
