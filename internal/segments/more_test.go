package segments

import (
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

	ctx.Git = git.Status{InRepo: false}
	if _, ok := ctx.render(t, r); ok {
		t.Error("git should hide when not in a repo")
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
	c := config.Default()
	c.RateLimits.Show = true
	s := &session.Session{RateLimits: &session.RateLimits{
		FiveHour: &session.RateWindow{UsedPercentage: 23.5},
		SevenDay: &session.RateWindow{UsedPercentage: 41.2},
	}}
	seg, ok := (&Context{Session: s, Config: c}).render(t, r)
	if !ok || seg.Text != "5h 24% 7d 41%" {
		t.Errorf("rate_limits = %q", seg.Text)
	}

	// hidden when absent
	if _, ok := (&Context{Session: &session.Session{}, Config: c}).render(t, r); ok {
		t.Error("rate_limits should hide when absent")
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
	s := &session.Session{}
	s.Cost.TotalLinesAdded = 156
	s.Cost.TotalLinesRemoved = 23
	seg, ok := (&Context{Session: s, Config: config.Default()}).render(t, r)
	if !ok || seg.Text != "+156 -23" {
		t.Errorf("lines = %q", seg.Text)
	}
	// hidden when no changes
	if _, ok := (&Context{Session: &session.Session{}, Config: config.Default()}).render(t, r); ok {
		t.Error("lines should hide when no changes")
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

// render is a tiny test helper that also sets Now to a fixed value.
func (c *Context) render(t *testing.T, r Renderer) (Segment, bool) {
	t.Helper()
	if c.Now.IsZero() {
		c.Now = time.Date(2026, 6, 14, 14, 32, 0, 0, time.UTC)
	}
	return r.Render(c)
}
