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

func stateFor(pct float64, warn, crit int) State {
	switch {
	case pct >= float64(crit):
		return StateCrit
	case pct >= float64(warn):
		return StateWarn
	default:
		return StateOK
	}
}

// Gauge renders a bar like "▓▓▓░░░░░ 42%" plus its threshold state.
// pct is clamped to [0,100]. When ascii is true it uses '#'/'-'.
func Gauge(pct float64, width, warn, crit int, ascii bool) (string, State) {
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
	bar := strings.Repeat(full, filled) + strings.Repeat(empty, width-filled)
	return fmt.Sprintf("%s %d%%", bar, int(pct+0.5)), stateFor(pct, warn, crit)
}
