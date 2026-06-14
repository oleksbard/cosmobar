package segments

import "github.com/oleksbard/cosmobar/internal/render"

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
	text, state := render.Gauge(*up, ctx.Config.GaugeWidth, warn, crit, ctx.Config.ASCII())
	return Segment{Name: "context", Text: text, State: state, Prio: 100}, true
}

func init() { register(contextSeg{}) }
