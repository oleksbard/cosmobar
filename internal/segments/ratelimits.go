package segments

import (
	"fmt"
	"strings"

	"github.com/oleksbard/cosmobar/internal/render"
)

type rateLimitsSeg struct{}

func (rateLimitsSeg) Name() string { return "rate_limits" }

func (rateLimitsSeg) Render(ctx *Context) (Segment, bool) {
	if !ctx.Config.RateLimits.Show {
		return Segment{}, false
	}
	rl := ctx.Session.RateLimits
	if rl == nil {
		return Segment{}, false
	}
	var parts []string
	if rl.FiveHour != nil {
		parts = append(parts, fmt.Sprintf("5h %.0f%%", rl.FiveHour.UsedPercentage))
	}
	if rl.SevenDay != nil {
		parts = append(parts, fmt.Sprintf("7d %.0f%%", rl.SevenDay.UsedPercentage))
	}
	if len(parts) == 0 {
		return Segment{}, false
	}
	return Segment{Name: "rate_limits", Text: strings.Join(parts, " "), State: render.StateNone, Prio: 40}, true
}

func init() { register(rateLimitsSeg{}) }
