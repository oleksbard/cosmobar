package segments

import "github.com/oleksbard/cosmobar/internal/render"

type modelSeg struct{}

func (modelSeg) Name() string { return "model" }

func (modelSeg) Render(ctx *Context) (Segment, bool) {
	name := ctx.Session.Model.DisplayName
	if name == "" {
		return Segment{}, false
	}
	return Segment{Name: "model", Text: name, State: render.StateNone, Prio: 80}, true
}

func init() { register(modelSeg{}) }
