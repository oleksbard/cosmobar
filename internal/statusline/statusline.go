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

	ctx := &segments.Context{Session: in.Session, Git: in.Git, Config: in.Config, Now: in.Now, Profile: in.Profile}

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

	// Truncate an over-wide dir before measuring/styling.
	for i := range segs {
		if segs[i].Name == "dir" && len(segs[i].Parts) == 0 && render.Width(segs[i].Text) > cols {
			segs[i].Text = render.Truncate(segs[i].Text, cols)
		}
	}

	style := effectiveStyle(in.Config, in.Profile)

	segRuns := make([][]run, len(segs))
	widths := make([]int, len(segs))
	prios := make([]int, len(segs))
	for i, s := range segs {
		role := "identity"
		if m, ok := segments.MetaByName(s.Name); ok {
			role = m.Role
		}
		segRuns[i] = segmentRuns(s, role, style, pal, in.Config)
		widths[i] = runsWidth(segRuns[i])
		prios[i] = s.Prio
	}

	rows := render.Fit(widths, prios, separatorWidth(in.Config, style), cols, in.Config.MaxRows)

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, emitRow(in.Profile, style, in.Config, pal, segRuns, row))
	}
	return strings.Join(lines, "\n")
}
