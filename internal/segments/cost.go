package segments

import (
	"fmt"
	"time"

	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/spend"
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
	text += rollupSuffix(ctx.Spend, ctx.Config.Cost.Rollups)
	return Segment{
		Name:  "cost",
		Text:  text,
		State: render.StateNone,
		Prio:  60,
	}, true
}

// rollupSuffix appends " · $X today" when the cost config enables the "today"
// rollup and there is non-trivial spend today; values that round to $0.00 are
// dropped (so a fresh session doesn't show "$0.00 today"). Only the today
// rollup is supported — see internal/spend for why week/month were removed.
func rollupSuffix(sp *spend.Rollup, windows []string) string {
	if sp == nil || sp.Today < 0.005 {
		return ""
	}
	for _, w := range windows {
		if w == "today" {
			return fmt.Sprintf(" · $%.2f today", sp.Today)
		}
	}
	return ""
}

func init() { register(costSeg{}) }
