package segments

import (
	"fmt"

	"github.com/oleksbard/cosmobar/internal/render"
)

type linesSeg struct{}

func (linesSeg) Name() string { return "lines" }

func (linesSeg) Render(ctx *Context) (Segment, bool) {
	a := ctx.Session.Cost.TotalLinesAdded
	r := ctx.Session.Cost.TotalLinesRemoved
	if a == 0 && r == 0 {
		return Segment{}, false
	}
	return Segment{Name: "lines", Text: fmt.Sprintf("+%d -%d", a, r), State: render.StateNone, Prio: 30}, true
}

func init() { register(linesSeg{}) }
