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
	Session *session.Session
	Git     git.Status
	Config  config.Config
	Now     time.Time
}

// Segment is one rendered piece of the status line.
type Segment struct {
	Name  string
	Text  string
	State render.State // StateNone unless this is a gauge
	Prio  int          // higher = kept longer when width is tight
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
