package segments

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oleksbard/cosmobar/internal/render"
)

// humanTokens formats a token count compactly: 847, 9.4k, 412k, 1.2M.
func humanTokens(n int) string {
	// Sub-100k uses one decimal (rounded); 100k–1M shows integer thousands
	// (truncated, so a value just under 1M never displays as "1000k").
	switch {
	case n < 1000:
		return strconv.Itoa(n)
	case n < 100_000:
		return trimZero(fmt.Sprintf("%.1f", float64(n)/1000)) + "k"
	case n < 1_000_000:
		return strconv.Itoa(n/1000) + "k"
	default:
		return trimZero(fmt.Sprintf("%.1f", float64(n)/1_000_000)) + "M"
	}
}

// trimZero drops a trailing ".0" (e.g. "9.0" -> "9").
func trimZero(s string) string {
	return strings.TrimSuffix(s, ".0")
}

type tokensSeg struct{}

func (tokensSeg) Name() string { return "tokens" }

func (tokensSeg) Render(ctx *Context) (Segment, bool) {
	if ctx.SessionTokens == nil {
		return Segment{}, false
	}
	return Segment{
		Name:  "tokens",
		Text:  humanTokens(ctx.SessionTokens.Total()) + " tok",
		State: render.StateNone,
		Prio:  55,
	}, true
}

func init() { register(tokensSeg{}) }
