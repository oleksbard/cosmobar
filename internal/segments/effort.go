package segments

import "github.com/oleksbard/cosmobar/internal/render"

type effortSeg struct{}

func (effortSeg) Name() string { return "effort" }

func (effortSeg) Render(ctx *Context) (Segment, bool) {
	if ctx.Session.Effort == nil || ctx.Session.Effort.Level == "" {
		return Segment{}, false
	}
	return Segment{Name: "effort", Text: "effort " + ctx.Session.Effort.Level, State: render.StateNone, Prio: 20}, true
}

func init() { register(effortSeg{}) }
