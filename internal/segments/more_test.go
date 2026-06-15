package segments

import (
	"strings"
	"testing"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
)

func TestGitSegment(t *testing.T) {
	r, _ := Get("git")
	c := config.Default()

	ctx := &Context{Session: &session.Session{}, Config: c, Git: git.Status{InRepo: true, Branch: "main", Staged: 1, Modified: 2}}
	seg, ok := ctx.render(t, r)
	if !ok || seg.Text != "main +1 ~2" {
		t.Errorf("git = %q ok=%v", seg.Text, ok)
	}

	ctx.Git = git.Status{InRepo: true, Branch: "main", Ahead: 1, Behind: 2}
	seg, _ = ctx.render(t, r)
	if seg.Text != "main ↑1 ↓2" {
		t.Errorf("git ahead/behind = %q", seg.Text)
	}

	// long branch names are capped at maxBranchWidth (28) with a middle ellipsis
	ctx.Git = git.Status{InRepo: true, Branch: "feature/a-very-long-branch-name"}
	seg, _ = ctx.render(t, r)
	if render.Width(seg.Text) > maxBranchWidth {
		t.Errorf("long branch should be capped at %d cols, got %q (%d)", maxBranchWidth, seg.Text, render.Width(seg.Text))
	}
	if !strings.Contains(seg.Text, "…") {
		t.Errorf("truncated branch should carry an ellipsis: %q", seg.Text)
	}

	// short branches are untouched
	ctx.Git = git.Status{InRepo: true, Branch: "main"}
	if seg, _ = ctx.render(t, r); seg.Text != "main" {
		t.Errorf("short branch should be unchanged, got %q", seg.Text)
	}

	ctx.Git = git.Status{InRepo: false}
	if _, ok := ctx.render(t, r); ok {
		t.Error("git should hide when not in a repo")
	}
}

func TestModelShorten(t *testing.T) {
	cases := map[string]string{
		"Opus 4.8 (1M context)": "Opus 4.8(1M)",
		"Sonnet 4.6":            "Sonnet 4.6",
		"Opus 4.8":              "Opus 4.8",
	}
	for in, want := range cases {
		if got := shortenModel(in); got != want {
			t.Errorf("shortenModel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestContextSegment(t *testing.T) {
	r, _ := Get("context")
	c := config.Default()
	pct := 42.0
	s := &session.Session{}
	s.ContextWindow.UsedPercentage = &pct
	seg, ok := (&Context{Session: s, Config: c}).render(t, r)
	if !ok || seg.Text != "▓▓▓░░░░░ 42%" || seg.State != render.StateOK {
		t.Errorf("context = %q state=%v", seg.Text, seg.State)
	}

	// hidden when nil
	if _, ok := (&Context{Session: &session.Session{}, Config: c}).render(t, r); ok {
		t.Error("context should hide when used_percentage is nil")
	}

	// hidden when toggled off
	c.Context.Show = false
	if _, ok := (&Context{Session: s, Config: c}).render(t, r); ok {
		t.Error("context should hide when show=false")
	}
}

func TestRateLimitsSegment(t *testing.T) {
	r, _ := Get("rate_limits")
	mk := func(five, seven float64) *session.Session {
		return &session.Session{RateLimits: &session.RateLimits{
			FiveHour: &session.RateWindow{UsedPercentage: five},
			SevenDay: &session.RateWindow{UsedPercentage: seven},
		}}
	}

	c := config.Default()
	c.RateLimits.Show = true // window "both"
	seg, ok := (&Context{Session: mk(23.5, 41.2), Config: c}).render(t, r)
	if !ok || seg.Text != "5h 24% 7d 41%" {
		t.Errorf("both = %q", seg.Text)
	}
	if seg.State != render.StateOK {
		t.Errorf("state should be OK below thresholds, got %v", seg.State)
	}

	c.RateLimits.Window = "7d"
	seg, _ = (&Context{Session: mk(23.5, 95), Config: c}).render(t, r)
	if seg.Text != "7d 95%" {
		t.Errorf("7d-only = %q", seg.Text)
	}
	if seg.State != render.StateCrit {
		t.Errorf("95%% should be crit, got %v", seg.State)
	}

	c.RateLimits.Window = "both"
	if _, ok := (&Context{Session: &session.Session{}, Config: c}).render(t, r); ok {
		t.Error("absent rate_limits should hide")
	}
}

func TestDurationSegment(t *testing.T) {
	r, _ := Get("duration")
	s := &session.Session{}
	s.Cost.TotalDurationMS = 723000 // 12m 03s
	seg, _ := (&Context{Session: s, Config: config.Default()}).render(t, r)
	if seg.Text != "12m 03s" {
		t.Errorf("duration = %q", seg.Text)
	}
	s.Cost.TotalDurationMS = 3723000 // 1h 02m
	seg, _ = (&Context{Session: s, Config: config.Default()}).render(t, r)
	if seg.Text != "1h 02m" {
		t.Errorf("duration hours = %q", seg.Text)
	}
}

func TestLinesSegment(t *testing.T) {
	r, _ := Get("lines")
	c := config.Default()

	ctx := &Context{Session: &session.Session{}, Config: c, Git: git.Status{InRepo: true, LinesAdded: 128, LinesRemoved: 17}}
	seg, ok := ctx.render(t, r)
	if !ok {
		t.Fatal("lines should show with changes")
	}
	if len(seg.Parts) != 2 || seg.Parts[0].Text != "+128" || seg.Parts[0].State != render.StateOK {
		t.Errorf("added part wrong: %+v", seg.Parts)
	}
	if seg.Parts[1].Text != "-17" || seg.Parts[1].State != render.StateCrit {
		t.Errorf("removed part wrong: %+v", seg.Parts)
	}
	// hidden when not in a repo
	if _, ok := (&Context{Session: &session.Session{}, Config: c, Git: git.Status{InRepo: false}}).render(t, r); ok {
		t.Error("lines should hide outside a repo")
	}
	// hidden when no changes
	if _, ok := (&Context{Session: &session.Session{}, Config: c, Git: git.Status{InRepo: true}}).render(t, r); ok {
		t.Error("lines should hide with no changes")
	}
}

func TestOutputStyleSegment(t *testing.T) {
	r, _ := Get("output_style")
	s := &session.Session{OutputStyle: session.OutputStyle{Name: "explanatory"}}
	seg, ok := (&Context{Session: s, Config: config.Default()}).render(t, r)
	if !ok || seg.Text != "explanatory" {
		t.Errorf("output_style = %q", seg.Text)
	}
}

func TestGitStashSegment(t *testing.T) {
	r, _ := Get("git_stash")
	c := config.Default()
	seg, ok := (&Context{Config: c, Session: &session.Session{}, Git: git.Status{InRepo: true, Stashes: 2}}).render(t, r)
	if !ok || seg.Text != "⚑2" {
		t.Errorf("git_stash = %q", seg.Text)
	}
	// ascii mode
	c.Glyphs = "ascii"
	seg, _ = (&Context{Config: c, Session: &session.Session{}, Git: git.Status{InRepo: true, Stashes: 2}}).render(t, r)
	if seg.Text != "stash:2" {
		t.Errorf("git_stash ascii = %q", seg.Text)
	}
	// hidden when none
	if _, ok := (&Context{Config: config.Default(), Session: &session.Session{}, Git: git.Status{InRepo: true}}).render(t, r); ok {
		t.Error("git_stash should hide when 0")
	}
}

func TestEffortSegment(t *testing.T) {
	r, _ := Get("effort")
	s := &session.Session{Effort: &session.Effort{Level: "high"}}
	seg, ok := (&Context{Session: s, Config: config.Default()}).render(t, r)
	if !ok || seg.Text != "effort high" {
		t.Errorf("effort = %q", seg.Text)
	}
	if _, ok := (&Context{Session: &session.Session{}, Config: config.Default()}).render(t, r); ok {
		t.Error("effort should hide when absent")
	}
}

func TestContextCompactInBackgroundStyle(t *testing.T) {
	r, _ := Get("context")
	pct := 63.0
	s := &session.Session{}
	s.ContextWindow.UsedPercentage = &pct

	c := config.Default()
	c.Style = "blocks"
	seg, ok := (&Context{Session: s, Config: c, Profile: render.ProfileTrueColor}).render(t, r)
	if !ok || seg.Text != "63%" {
		t.Errorf("blocks context should be compact, got %q", seg.Text)
	}

	// blocks + NO_COLOR degrades to lean → keep the bar
	seg, _ = (&Context{Session: s, Config: c, Profile: render.ProfileNone}).render(t, r)
	if seg.Text == "63%" {
		t.Errorf("NO_COLOR should keep the bar, got %q", seg.Text)
	}
}

// render is a tiny test helper that also sets Now to a fixed value.
func (c *Context) render(t *testing.T, r Renderer) (Segment, bool) {
	t.Helper()
	if c.Now.IsZero() {
		c.Now = time.Date(2026, 6, 14, 14, 32, 0, 0, time.UTC)
	}
	return r.Render(c)
}
