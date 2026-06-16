package segments

import (
	"fmt"
	"time"

	"github.com/oleksbard/cosmobar/internal/render"
)

type costSeg struct{}

func (costSeg) Name() string { return "cost" }

// burnMinDurationMS is the minimum session duration before a $/hr rate is
// shown — below it the extrapolated rate is too noisy to be useful.
const burnMinDurationMS = 60_000

func (costSeg) Render(ctx *Context) (Segment, bool) {
	cost := ctx.Session.Cost.TotalCostUSD
	text := fmt.Sprintf("$%.2f", cost)
	if durMS := ctx.Session.Cost.TotalDurationMS; cost > 0 && durMS >= burnMinDurationMS {
		hours := float64(durMS) / float64(time.Hour/time.Millisecond)
		text += fmt.Sprintf(" ($%.2f/hr)", cost/hours)
	}
	return Segment{
		Name:  "cost",
		Text:  text,
		State: render.StateNone,
		Prio:  60,
	}, true
}

func init() { register(costSeg{}) }
