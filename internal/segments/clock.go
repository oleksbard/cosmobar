package segments

import "github.com/oleksbard/cosmobar/internal/render"

type clockSeg struct{}

func (clockSeg) Name() string { return "clock" }

func (clockSeg) Render(ctx *Context) (Segment, bool) {
	switch ctx.Config.Clock.Format {
	case "off":
		return Segment{}, false
	case "12h":
		return Segment{Name: "clock", Text: ctx.Now.Format("3:04 PM"), State: render.StateNone, Prio: 50}, true
	default:
		return Segment{Name: "clock", Text: ctx.Now.Format("15:04"), State: render.StateNone, Prio: 50}, true
	}
}

func init() { register(clockSeg{}) }
