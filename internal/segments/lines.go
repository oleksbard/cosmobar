package segments

import (
	"fmt"

	"github.com/oleksbard/cosmobar/internal/render"
)

type linesSeg struct{}

func (linesSeg) Name() string { return "lines" }

func (linesSeg) Render(ctx *Context) (Segment, bool) {
	if !ctx.Git.InRepo {
		return Segment{}, false
	}
	a, r := ctx.Git.LinesAdded, ctx.Git.LinesRemoved
	if a == 0 && r == 0 {
		return Segment{}, false
	}
	return Segment{
		Name: "lines",
		Parts: []Part{
			{Text: fmt.Sprintf("+%d", a), State: render.StateOK},
			{Text: fmt.Sprintf("-%d", r), State: render.StateCrit},
		},
		Prio: 30,
	}, true
}

func init() { register(linesSeg{}) }
