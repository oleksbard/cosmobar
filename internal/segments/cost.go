package segments

import (
	"fmt"
	"strings"
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

// rollupSuffix appends " · $X today" / " · $X wk" / " · $X mo" for each enabled
// window with a non-trivial total. Windows always render in today→week→month
// order; unknown window names are ignored; values that round to $0.00 are
// dropped (so a fresh session doesn't show "$0.00 today").
func rollupSuffix(sp *spend.Rollup, windows []string) string {
	if sp == nil {
		return ""
	}
	enabled := make(map[string]bool, len(windows))
	for _, w := range windows {
		enabled[w] = true
	}
	var b strings.Builder
	add := func(name, label string, v float64) {
		if enabled[name] && v >= 0.005 {
			fmt.Fprintf(&b, " · $%.2f %s", v, label)
		}
	}
	add("today", "today", sp.Today)
	add("week", "wk", sp.Week)
	add("month", "mo", sp.Month)
	return b.String()
}

func init() { register(costSeg{}) }
