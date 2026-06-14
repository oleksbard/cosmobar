package segments

import (
	"fmt"
	"time"

	"github.com/oleksbard/cosmobar/internal/render"
)

type durationSeg struct{}

func (durationSeg) Name() string { return "duration" }

func (durationSeg) Render(ctx *Context) (Segment, bool) {
	d := time.Duration(ctx.Session.Cost.TotalDurationMS) * time.Millisecond
	var text string
	if d >= time.Hour {
		text = fmt.Sprintf("%dh %02dm", int(d.Hours()), int(d.Minutes())%60)
	} else {
		text = fmt.Sprintf("%dm %02ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return Segment{Name: "duration", Text: text, State: render.StateNone, Prio: 30}, true
}

func init() { register(durationSeg{}) }
