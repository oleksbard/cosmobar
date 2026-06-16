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
	state := render.StateFor(*up, warn, crit)
	label := contextLabel(*up, ctx.Session.ContextWindow.ContextWindowSize)
	// In background styles the colored block conveys state, so drop the bar —
	// but only when color is actually available.
	if ctx.Config.BackgroundStyle() && ctx.Profile != render.ProfileNone {
		return Segment{Name: "context", Text: label, State: state, Prio: 100}, true
	}
	bar := render.GaugeBar(*up, ctx.Config.GaugeWidth, ctx.Config.ASCII())
	return Segment{Name: "context", Text: bar + " " + label, State: state, Prio: 100}, true
}

// contextLabel formats the gauge readout. With a known window size it shows
// absolute tokens and the percent (e.g. "126k/200k (63%)"); otherwise it falls
// back to the bare percentage.
func contextLabel(pct float64, size int) string {
	rounded := int(pct + 0.5)
	if size <= 0 {
		return fmt.Sprintf("%d%%", rounded)
	}
	used := int(pct/100*float64(size) + 0.5)
	return fmt.Sprintf("%s/%s (%d%%)", humanTokens(used), humanTokens(size), rounded)
}

func init() { register(contextSeg{}) }
