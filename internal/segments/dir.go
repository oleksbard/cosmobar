package segments

import (
	"path/filepath"
	"strings"

	"github.com/oleksbard/cosmobar/internal/render"
)

type dirSeg struct{}

func (dirSeg) Name() string { return "dir" }

func (dirSeg) Render(ctx *Context) (Segment, bool) {
	full := ctx.Session.Dir()
	if full == "" {
		return Segment{}, false
	}
	var text string
	switch ctx.Config.Dir.Style {
	case "full":
		text = full
	case "short-path":
		text = shortPath(full)
	default:
		text = filepath.Base(full)
	}
	return Segment{Name: "dir", Text: text, State: render.StateNone, Prio: 70}, true
}

// shortPath keeps the last two path components, prefixed with "…/" if deeper.
func shortPath(p string) string {
	p = filepath.Clean(p)
	parts := strings.Split(p, string(filepath.Separator))
	var nonEmpty []string
	for _, s := range parts {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	if len(nonEmpty) <= 2 {
		return p
	}
	return "…/" + filepath.Join(nonEmpty[len(nonEmpty)-2:]...)
}

func init() { register(dirSeg{}) }
