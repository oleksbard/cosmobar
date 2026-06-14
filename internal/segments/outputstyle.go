package segments

import "github.com/oleksbard/cosmobar/internal/render"

type outputStyleSeg struct{}

func (outputStyleSeg) Name() string { return "output_style" }

func (outputStyleSeg) Render(ctx *Context) (Segment, bool) {
	name := ctx.Session.OutputStyle.Name
	if name == "" {
		return Segment{}, false
	}
	return Segment{Name: "output_style", Text: name, State: render.StateNone, Prio: 20}, true
}

func init() { register(outputStyleSeg{}) }
