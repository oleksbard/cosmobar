// Package segments defines the statusline segment registry. Each segment
// reads from the render Context and returns formatted text plus a priority
// used by the layout engine. Segments register themselves in init().
package segments

import (
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
)

// Context is everything a segment needs to render.
type Context struct {
	Session       *session.Session
	Git           git.Status
	Config        config.Config
	Now           time.Time
	Profile       render.Profile
	SessionTokens *session.TokenUsage
}

// Part is one colored fragment of a segment (e.g. the "+12" in a lines count).
// A zero State means "use the segment's role color".
type Part struct {
	Text  string
	State render.State
}

// Segment is one rendered piece of the status line.
type Segment struct {
	Name  string
	Text  string       // used when Parts is empty
	Parts []Part       // when non-empty, overrides Text
	State render.State // StateNone unless this is a gauge
	Prio  int          // higher = kept longer when width is tight
}

// EffectiveParts returns Parts, or a single part built from Text/State.
func (s Segment) EffectiveParts() []Part {
	if len(s.Parts) > 0 {
		return s.Parts
	}
	return []Part{{Text: s.Text, State: s.State}}
}

// Renderer produces a Segment. ok=false means "hide this segment now".
type Renderer interface {
	Name() string
	Render(ctx *Context) (Segment, bool)
}

var registry = map[string]Renderer{}

func register(r Renderer) { registry[r.Name()] = r }

// Get returns the registered renderer for name.
func Get(name string) (Renderer, bool) {
	r, ok := registry[name]
	return r, ok
}
