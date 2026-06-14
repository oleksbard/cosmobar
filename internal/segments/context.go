package segments

import (
	"fmt"

	"github.com/oleksbard/cosmobar/internal/render"
)

type contextSeg struct{}

func (contextSeg) Name() string { return "context" }

func (contextSeg) Render(ctx *Context) (Segment, bool) {
	if !ctx.Config.Context.Show {
		return Segment{}, false
	}
	up := ctx.Session.ContextWindow.UsedPercentage
	if up == nil {
		return Segment{}, false
	}
	warn, crit := ctx.Config.Thresholds()
	// In background styles the colored block conveys state, so drop the bar —
	// but only when color is actually available.
	if ctx.Config.BackgroundStyle() && ctx.Profile != render.ProfileNone {
		return Segment{Name: "context", Text: fmt.Sprintf("%d%%", int(*up+0.5)), State: render.StateFor(*up, warn, crit), Prio: 100}, true
	}
	text, state := render.Gauge(*up, ctx.Config.GaugeWidth, warn, crit, ctx.Config.ASCII())
	return Segment{Name: "context", Text: text, State: state, Prio: 100}, true
}

func init() { register(contextSeg{}) }
