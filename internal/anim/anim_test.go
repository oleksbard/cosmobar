package anim

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/render"
)

func enabledCfg() config.Config {
	c := config.Default()
	c.Animation = config.AnimationConfig{Enabled: true, DurationMs: 700, Variants: []string{"decode", "glitch", "scatter"}}
	return c
}

func TestPlanFirstPaintNoScramble(t *testing.T) {
	s := newSession(enabledCfg(), render.ProfileTrueColor)
	now := time.Unix(1000, 0)
	p := s.Plan("cost", "$1.20", now)
	if p.Active {
		t.Error("first paint must not animate")
	}
}

func TestPlanChangeStartsAnimation(t *testing.T) {
	s := newSession(enabledCfg(), render.ProfileTrueColor)
	now := time.Unix(1000, 0)
	s.Plan("cost", "$1.20", now) // first paint
	s.Commit()
	p := s.Plan("cost", "$1.45", now.Add(10*time.Millisecond)) // value changed
	if !p.Active {
		t.Fatal("change should animate")
	}
	if p.Progress < 0 || p.Progress >= 1 {
		t.Errorf("progress = %v, want [0,1)", p.Progress)
	}
}

func TestPlanCompletesAfterDuration(t *testing.T) {
	s := newSession(enabledCfg(), render.ProfileTrueColor)
	now := time.Unix(1000, 0)
	s.Plan("cost", "$1.20", now)
	s.Commit()
	s.Plan("cost", "$1.45", now.Add(10*time.Millisecond))
	s.Commit()
	p := s.Plan("cost", "$1.45", now.Add(2*time.Second)) // well past duration
	if p.Active {
		t.Error("animation should be complete after duration")
	}
}

func TestPlanDisabledWhenNoColor(t *testing.T) {
	s := newSession(enabledCfg(), render.ProfileNone)
	if s.Plan("cost", "$1.20", time.Unix(1000, 0)).Active {
		t.Error("no-color profile disables animation")
	}
}

func TestNilSessionPlanInactive(t *testing.T) {
	var s *Session
	if s.Plan("x", "y", time.Unix(1, 0)).Active {
		t.Error("nil session must be inactive")
	}
}

func TestCommitPrunesVanishedSegments(t *testing.T) {
	s := newSession(enabledCfg(), render.ProfileTrueColor)
	now := time.Unix(1000, 0)
	s.Plan("cost", "$1.20", now)
	s.Plan("git", "main", now)
	s.Commit()
	// next render only sees "cost"
	s.Plan("cost", "$1.20", now.Add(time.Second))
	s.Commit()
	if _, ok := s.State.Segments["git"]; ok {
		t.Error("vanished segment should be pruned")
	}
	if _, ok := s.State.Segments["cost"]; !ok {
		t.Error("present segment should be kept")
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := newSession(enabledCfg(), render.ProfileTrueColor)
	s.path = filepath.Join(dir, "anim-state")
	now := time.Unix(1000, 0)
	s.Plan("cost", "$1.20", now)
	s.Commit()
	s.Save()

	s2 := newSession(enabledCfg(), render.ProfileTrueColor)
	s2.path = s.path
	s2.loadFrom(s.path)
	if s2.State.Segments["cost"].Target != "$1.20" {
		t.Errorf("round trip lost state: %+v", s2.State.Segments)
	}
}

func TestLoadMissingFileIsEmpty(t *testing.T) {
	s := Load("no-such-session-id-xyz", enabledCfg(), render.ProfileTrueColor)
	defer os.Remove(s.path) // always clean up, even if assertions fail
	if len(s.State.Segments) != 0 {
		t.Errorf("missing file should yield empty state, got %+v", s.State.Segments)
	}
}

func TestLoadFromCorruptFileIsEmpty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "corrupt")
	if err := os.WriteFile(p, []byte("not json{"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := newSession(enabledCfg(), render.ProfileTrueColor)
	s.loadFrom(p)
	if len(s.State.Segments) != 0 {
		t.Errorf("corrupt file should yield empty state, got %+v", s.State.Segments)
	}
}

func TestDemoPlanAlwaysActiveWithinDuration(t *testing.T) {
	start := time.Unix(2000, 0)
	s := Demo(enabledCfg(), render.ProfileTrueColor, start)
	p := s.Plan("anything", "value", start.Add(100*time.Millisecond))
	if !p.Active {
		t.Error("demo should animate within duration")
	}
	if s.Plan("anything", "value", start.Add(5*time.Second)).Active {
		t.Error("demo should complete after duration")
	}
}

func TestUnchangedSegmentKeepsStart(t *testing.T) {
	s := newSession(enabledCfg(), render.ProfileTrueColor)
	now := time.Unix(1000, 0)
	s.Plan("cost", "$1.20", now)
	s.Commit()
	start1 := s.State.Segments["cost"].StartMs
	// Same value on the next render: the entry must carry forward unchanged
	// (no restart), so the animation isn't perpetually reset.
	s.Plan("cost", "$1.20", now.Add(50*time.Millisecond))
	s.Commit()
	if s.State.Segments["cost"].StartMs != start1 {
		t.Errorf("unchanged segment changed StartMs: %d -> %d", start1, s.State.Segments["cost"].StartMs)
	}
}
