package segments

import (
	"testing"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
)

func ctxWith(s *session.Session, c config.Config, now time.Time) *Context {
	return &Context{Session: s, Config: c, Now: now}
}

func TestDirBasename(t *testing.T) {
	s := &session.Session{Cwd: "/Users/me/projects/cosmobar"}
	c := config.Default()
	r, _ := Get("dir")
	seg, ok := r.Render(ctxWith(s, c, time.Time{}))
	if !ok || seg.Text != "cosmobar" {
		t.Errorf("dir = %q ok=%v", seg.Text, ok)
	}
}

func TestModelHiddenWhenEmpty(t *testing.T) {
	r, _ := Get("model")
	if _, ok := r.Render(ctxWith(&session.Session{}, config.Default(), time.Time{})); ok {
		t.Error("model should hide when display_name empty")
	}
}

func TestCost(t *testing.T) {
	s := &session.Session{}
	s.Cost.TotalCostUSD = 0.12
	r, _ := Get("cost")
	seg, _ := r.Render(ctxWith(s, config.Default(), time.Time{}))
	if seg.Text != "$0.12" {
		t.Errorf("cost = %q", seg.Text)
	}
}

func TestCostBurnRateSuffix(t *testing.T) {
	r, _ := Get("cost")

	// cost + a meaningful duration → append $/hr.
	s := &session.Session{}
	s.Cost.TotalCostUSD = 1.0
	s.Cost.TotalDurationMS = 3_600_000 // exactly one hour → $1.00/hr
	seg, _ := r.Render(ctxWith(s, config.Default(), time.Time{}))
	if seg.Text != "$1.00 ($1.00/hr)" {
		t.Errorf("cost with burn = %q, want %q", seg.Text, "$1.00 ($1.00/hr)")
	}

	// no duration → no burn suffix (avoids a divide-by-zero / absurd rate).
	s2 := &session.Session{}
	s2.Cost.TotalCostUSD = 0.12
	seg, _ = r.Render(ctxWith(s2, config.Default(), time.Time{}))
	if seg.Text != "$0.12" {
		t.Errorf("cost without duration = %q, want %q", seg.Text, "$0.12")
	}

	// zero cost → no burn suffix ($0.00/hr is meaningless).
	s3 := &session.Session{}
	s3.Cost.TotalDurationMS = 3_600_000
	seg, _ = r.Render(ctxWith(s3, config.Default(), time.Time{}))
	if seg.Text != "$0.00" {
		t.Errorf("zero cost = %q, want %q", seg.Text, "$0.00")
	}

	// sub-minute duration → too noisy, no suffix.
	s4 := &session.Session{}
	s4.Cost.TotalCostUSD = 0.50
	s4.Cost.TotalDurationMS = 30_000
	seg, _ = r.Render(ctxWith(s4, config.Default(), time.Time{}))
	if seg.Text != "$0.50" {
		t.Errorf("sub-minute duration = %q, want %q", seg.Text, "$0.50")
	}
}

func TestClockFormats(t *testing.T) {
	now := time.Date(2026, 6, 14, 14, 32, 0, 0, time.UTC)
	r, _ := Get("clock")

	c := config.Default()
	seg, _ := r.Render(ctxWith(&session.Session{}, c, now))
	if seg.Text != "14:32" {
		t.Errorf("24h clock = %q", seg.Text)
	}

	c.Clock.Format = "off"
	if _, ok := r.Render(ctxWith(&session.Session{}, c, now)); ok {
		t.Error("clock off should hide")
	}
}

func TestEffectivePartsFallsBackToText(t *testing.T) {
	s := Segment{Text: "hi", State: render.StateWarn}
	ps := s.EffectiveParts()
	if len(ps) != 1 || ps[0].Text != "hi" || ps[0].State != render.StateWarn {
		t.Errorf("single-part fallback wrong: %+v", ps)
	}
	multi := Segment{Parts: []Part{{Text: "+1", State: render.StateOK}, {Text: "-2", State: render.StateCrit}}}
	if len(multi.EffectiveParts()) != 2 {
		t.Errorf("multi-part should return its parts")
	}
}
