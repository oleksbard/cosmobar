package anim

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/render"
)

// Entry is the persisted animation state for one segment.
type Entry struct {
	Target  string `json:"target"`   // last-seen full text signature
	StartMs int64  `json:"start_ms"` // when the current animation started
}

// State is the per-session animation state, keyed by segment name.
type State struct {
	Segments map[string]Entry `json:"segments"`
}

// settings holds the resolved animation options.
type settings struct {
	enabled    bool
	durationMs int
	variants   []string
	ascii      bool
}

// Session bundles loaded state with resolved settings and a storage path, so
// Render can read prior state and accumulate the next state in place.
type Session struct {
	path      string
	State     State
	set       settings
	next      map[string]Entry
	demoStart time.Time // non-zero in preview demo mode
}

// Plan is the animation decision for one segment this frame.
type Plan struct {
	Active   bool
	Progress float64
	Variant  string
	Seed     uint64
	ASCII    bool
}

func validVariants(in []string) []string {
	var out []string
	for _, v := range in {
		if _, ok := palettes[v]; ok {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), variantOrder...)
	}
	return out
}

func newSession(cfg config.Config, prof render.Profile) *Session {
	return &Session{
		State: State{Segments: map[string]Entry{}},
		next:  map[string]Entry{},
		set: settings{
			enabled:    cfg.Animation.Enabled && prof != render.ProfileNone,
			durationMs: cfg.Animation.DurationMs,
			variants:   validVariants(cfg.Animation.Variants),
			ascii:      cfg.ASCII(),
		},
	}
}

// Load resolves settings and reads any persisted state for sessionID from the
// OS temp dir. A missing/corrupt file yields empty state.
func Load(sessionID string, cfg config.Config, prof render.Profile) *Session {
	s := newSession(cfg, prof)
	s.path = filepath.Join(os.TempDir(), "cosmobar-anim-"+sanitize(sessionID))
	s.loadFrom(s.path)
	return s
}

// Demo returns a session that animates every segment from start, ignoring
// persisted state. It is always enabled (it exists to be watched), regardless
// of the color profile. Used by `preview --animate`.
func Demo(cfg config.Config, prof render.Profile, start time.Time) *Session {
	s := newSession(cfg, prof)
	s.set.enabled = true
	s.demoStart = start
	return s
}

func (s *Session) loadFrom(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var st State
	if json.Unmarshal(data, &st) == nil && st.Segments != nil {
		s.State = st
	}
}

func (s *Session) dur() time.Duration {
	return time.Duration(s.set.durationMs) * time.Millisecond
}

// Plan advances the state machine for segment name whose current full-text
// signature is sig, records its next state, and returns how to render it.
func (s *Session) Plan(name, sig string, now time.Time) Plan {
	if s == nil || !s.set.enabled {
		return Plan{}
	}
	if !s.demoStart.IsZero() {
		return s.demoPlan(name, sig, now)
	}
	dur := s.dur()
	prior, seen := s.State.Segments[name]
	var entry Entry
	switch {
	case !seen:
		entry = Entry{Target: sig, StartMs: now.Add(-dur).UnixMilli()} // no scramble on first paint
	case prior.Target != sig:
		entry = Entry{Target: sig, StartMs: now.UnixMilli()} // value changed → (re)start
	default:
		entry = prior
	}
	s.next[name] = entry

	elapsed := now.Sub(time.UnixMilli(entry.StartMs))
	if dur <= 0 || elapsed >= dur {
		return Plan{}
	}
	seed := seedFor(name, entry.Target, entry.StartMs)
	return Plan{
		Active:   true,
		Progress: float64(elapsed) / float64(dur),
		Variant:  pickVariant(s.set.variants, seed),
		Seed:     seed,
		ASCII:    s.set.ascii,
	}
}

func (s *Session) demoPlan(name, sig string, now time.Time) Plan {
	dur := s.dur()
	elapsed := now.Sub(s.demoStart)
	if elapsed < 0 {
		elapsed = 0
	}
	if dur <= 0 || elapsed >= dur {
		return Plan{}
	}
	seed := seedFor(name, sig, s.demoStart.UnixMilli())
	return Plan{
		Active:   true,
		Progress: float64(elapsed) / float64(dur),
		Variant:  pickVariant(s.set.variants, seed),
		Seed:     seed,
		ASCII:    s.set.ascii,
	}
}

// Commit swaps in the entries accumulated this render, pruning vanished segments.
func (s *Session) Commit() {
	if s == nil {
		return
	}
	s.State.Segments = s.next
	s.next = map[string]Entry{}
}

// Save persists the committed state to disk (best-effort). No-op for demo/nil
// sessions and when animation is disabled — a disabled run has no state worth
// writing, and skipping the write preserves any state from an enabled run.
func (s *Session) Save() {
	if s == nil || !s.set.enabled || !s.demoStart.IsZero() || s.path == "" {
		return
	}
	if data, err := json.Marshal(s.State); err == nil {
		_ = os.WriteFile(s.path, data, 0o600)
	}
}

// sanitize keeps a session id safe for a filename.
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "default"
	}
	return b.String()
}
