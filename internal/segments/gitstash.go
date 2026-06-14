package segments

import (
	"fmt"

	"github.com/oleksbard/cosmobar/internal/render"
)

type gitStashSeg struct{}

func (gitStashSeg) Name() string { return "git_stash" }

func (gitStashSeg) Render(ctx *Context) (Segment, bool) {
	if !ctx.Git.InRepo || ctx.Git.Stashes == 0 {
		return Segment{}, false
	}
	text := fmt.Sprintf("⚑%d", ctx.Git.Stashes)
	if ctx.Config.ASCII() {
		text = fmt.Sprintf("stash:%d", ctx.Git.Stashes)
	}
	return Segment{Name: "git_stash", Text: text, State: render.StateNone, Prio: 10}, true
}

func init() { register(gitStashSeg{}) }
