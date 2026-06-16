package render

import (
	"fmt"
	"strings"
)

// State is a threshold state used to color gauges.
type State int

const (
	StateNone State = iota
	StateOK
	StateWarn
	StateCrit
)

// StateFor maps a percentage to a threshold state.
func StateFor(pct float64, warn, crit int) State {
	switch {
	case pct >= float64(crit):
		return StateCrit
	case pct >= float64(warn):
		return StateWarn
	default:
		return StateOK
	}
}

// GaugeBar renders just the filled/empty bar like "▓▓▓░░░░░" (no label).
// pct is clamped to [0,100]. When ascii is true it uses '#'/'-'.
func GaugeBar(pct float64, width int, ascii bool) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct/100*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	full, empty := "▓", "░"
	if ascii {
		full, empty = "#", "-"
	}
	return strings.Repeat(full, filled) + strings.Repeat(empty, width-filled)
}

// Gauge renders a bar like "▓▓▓░░░░░ 42%" plus its threshold state.
// pct is clamped to [0,100]. When ascii is true it uses '#'/'-'.
func Gauge(pct float64, width, warn, crit int, ascii bool) (string, State) {
	return fmt.Sprintf("%s %d%%", GaugeBar(pct, width, ascii), int(clampPct(pct)+0.5)), StateFor(pct, warn, crit)
}

// clampPct clamps a percentage to [0,100] for display rounding.
func clampPct(pct float64) float64 {
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}
