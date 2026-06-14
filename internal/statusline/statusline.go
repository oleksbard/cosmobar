// Package statusline orchestrates segment rendering into the final output.
package statusline

import (
	"strings"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/segments"
	"github.com/oleksbard/cosmobar/internal/session"
	"github.com/oleksbard/cosmobar/internal/theme"
)

// Input bundles everything Render needs. Git, Cols, Profile, and Now are
// injected so the whole render is deterministic and testable.
type Input struct {
	Session *session.Session
	Git     git.Status
	Config  config.Config
	Cols    int
	Profile render.Profile
	Now     time.Time
}

// Render produces the final status line (one or more newline-separated rows).
func Render(in Input) string {
	pal, ok := theme.Get(in.Config.Theme)
	if !ok {
		pal, _ = theme.Get("coral")
	}
	cols := in.Cols
	if cols <= 0 {
		cols = 80
	}

	ctx := &segments.Context{Session: in.Session, Git: in.Git, Config: in.Config, Now: in.Now}

	// 1. collect enabled, visible segments in configured order
	var segs []segments.Segment
	for _, name := range in.Config.Order {
		r, ok := segments.Get(name)
		if !ok {
			continue
		}
		if seg, show := r.Render(ctx); show {
			segs = append(segs, seg)
		}
	}
	if len(segs) == 0 {
		return ""
	}

	// 2. truncate the dir segment if it alone exceeds the width
	for i := range segs {
		if segs[i].Name == "dir" && render.Width(segs[i].Text) > cols {
			segs[i].Text = render.Truncate(segs[i].Text, cols)
		}
	}

	// 3. lay out into rows
	widths := make([]int, len(segs))
	prios := make([]int, len(segs))
	for i, s := range segs {
		widths[i] = render.Width(s.Text)
		prios[i] = s.Prio
	}
	rows := render.Fit(widths, prios, render.Width(in.Config.Separator), cols, in.Config.MaxRows)

	// 4. colorize and join
	sep := render.Colorize(in.Profile, pal.Muted, in.Config.Separator)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		parts := make([]string, 0, len(row))
		for _, i := range row {
			parts = append(parts, colorizeSeg(in.Profile, pal, segs[i]))
		}
		lines = append(lines, strings.Join(parts, sep))
	}
	return strings.Join(lines, "\n")
}

func colorizeSeg(p render.Profile, pal theme.Palette, s segments.Segment) string {
	var c theme.RGB
	switch s.State {
	case render.StateOK:
		c = pal.OK
	case render.StateWarn:
		c = pal.Warn
	case render.StateCrit:
		c = pal.Crit
	default:
		c = pal.Accent
	}
	return render.Colorize(p, c, s.Text)
}
