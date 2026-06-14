package segments

import (
	"strings"

	"github.com/oleksbard/cosmobar/internal/render"
)

type modelSeg struct{}

func (modelSeg) Name() string { return "model" }

func (modelSeg) Render(ctx *Context) (Segment, bool) {
	name := ctx.Session.Model.DisplayName
	if name == "" {
		return Segment{}, false
	}
	return Segment{Name: "model", Text: shortenModel(name), State: render.StateNone, Prio: 80}, true
}

// shortenModel trims Claude Code's verbose model labels for the status line,
// e.g. "Opus 4.8 (1M context)" -> "Opus 4.8(1M)". Names without a parenthetical
// are returned unchanged.
func shortenModel(name string) string {
	name = strings.ReplaceAll(name, " context)", ")")
	name = strings.ReplaceAll(name, "context)", ")")
	name = strings.ReplaceAll(name, " (", "(")
	return name
}

func init() { register(modelSeg{}) }
