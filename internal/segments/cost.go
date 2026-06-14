package segments

import (
	"fmt"

	"github.com/oleksbard/cosmobar/internal/render"
)

type costSeg struct{}

func (costSeg) Name() string { return "cost" }

func (costSeg) Render(ctx *Context) (Segment, bool) {
	return Segment{
		Name:  "cost",
		Text:  fmt.Sprintf("$%.2f", ctx.Session.Cost.TotalCostUSD),
		State: render.StateNone,
		Prio:  60,
	}, true
}

func init() { register(costSeg{}) }
