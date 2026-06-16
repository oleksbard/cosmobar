package segments

import (
	"fmt"
	"strings"
	"time"

	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
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
		parts = append(parts, rateWindowPart("5h", rl.FiveHour, ctx.Now))
		if rl.FiveHour.UsedPercentage > maxPct {
			maxPct = rl.FiveHour.UsedPercentage
		}
	}
	if (window == "both" || window == "7d") && rl.SevenDay != nil {
		parts = append(parts, rateWindowPart("7d", rl.SevenDay, ctx.Now))
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

// rateWindowPart formats one window as "5h 31%", appending "(2h30m left)" when
// the window has a future reset time.
func rateWindowPart(label string, w *session.RateWindow, now time.Time) string {
	s := fmt.Sprintf("%s %.0f%%", label, w.UsedPercentage)
	if w.ResetsAt > 0 {
		if left := time.Unix(w.ResetsAt, 0).Sub(now); left > 0 {
			s += " (" + compactDuration(left) + " left)"
		}
	}
	return s
}

// compactDuration renders a coarse, human countdown: "3d", "2h30m", "45m",
// "<1m". Only the two largest relevant units are shown.
func compactDuration(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	case d >= time.Hour:
		return fmt.Sprintf("%dh%02dm", d/time.Hour, (d%time.Hour)/time.Minute)
	case d >= time.Minute:
		return fmt.Sprintf("%dm", d/time.Minute)
	default:
		return "<1m"
	}
}

func init() { register(rateLimitsSeg{}) }
