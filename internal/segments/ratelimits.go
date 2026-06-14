package segments

import (
	"fmt"
	"strings"

	"github.com/oleksbard/cosmobar/internal/render"
)

type rateLimitsSeg struct{}

func (rateLimitsSeg) Name() string { return "rate_limits" }

func (rateLimitsSeg) Render(ctx *Context) (Segment, bool) {
	cfg := ctx.Config.RateLimits
	if !cfg.Show {
		return Segment{}, false
	}
	rl := ctx.Session.RateLimits
	if rl == nil {
		return Segment{}, false
	}
	window := cfg.Window
	if window == "" {
		window = "both"
	}
	var parts []string
	maxPct := -1.0
	if (window == "both" || window == "5h") && rl.FiveHour != nil {
		parts = append(parts, fmt.Sprintf("5h %.0f%%", rl.FiveHour.UsedPercentage))
		if rl.FiveHour.UsedPercentage > maxPct {
			maxPct = rl.FiveHour.UsedPercentage
		}
	}
	if (window == "both" || window == "7d") && rl.SevenDay != nil {
		parts = append(parts, fmt.Sprintf("7d %.0f%%", rl.SevenDay.UsedPercentage))
		if rl.SevenDay.UsedPercentage > maxPct {
			maxPct = rl.SevenDay.UsedPercentage
		}
	}
	if len(parts) == 0 {
		return Segment{}, false
	}
	warn, crit := ctx.Config.Thresholds()
	return Segment{
		Name:  "rate_limits",
		Text:  strings.Join(parts, " "),
		State: render.StateFor(maxPct, warn, crit),
		Prio:  40,
	}, true
}

func init() { register(rateLimitsSeg{}) }
