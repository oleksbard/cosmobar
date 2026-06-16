// Package statusline orchestrates segment rendering into the final output.
package statusline

import (
	"strings"
	"time"

	"github.com/oleksbard/cosmobar/internal/anim"
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
	// Anim, when non-nil and enabled, scrambles changed segment values.
	Anim *anim.Session
	// SessionTokens, when non-nil, supplies cumulative session token usage to
	// the tokens segment.
	SessionTokens *session.TokenUsage
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

	ctx := &segments.Context{Session: in.Session, Git: in.Git, Config: in.Config, Now: in.Now, Profile: in.Profile, SessionTokens: in.SessionTokens}

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

	if in.Anim != nil {
		for i := range segs {
			animateSegment(in.Anim, &segs[i], in.Now)
		}
		in.Anim.Commit()
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

// segmentSignature is the change-detection key for a segment: its part texts
// joined by NUL so distinct part splits can't collide.
func segmentSignature(parts []segments.Part) string {
	var b strings.Builder
	for i, pt := range parts {
		if i > 0 {
			b.WriteByte(0)
		}
		b.WriteString(pt.Text)
	}
	return b.String()
}

// animateSegment replaces a segment's part texts with their scramble frame for
// this instant, preserving each part's color/state and the segment's width.
func animateSegment(sess *anim.Session, seg *segments.Segment, now time.Time) {
	parts := seg.EffectiveParts()
	p := sess.Plan(seg.Name, segmentSignature(parts), now)
	if !p.Active {
		return
	}
	newParts := make([]segments.Part, len(parts))
	for i, pt := range parts {
		newParts[i] = segments.Part{
			Text:  anim.Frame(pt.Text, p.Progress, p.Variant, p.Seed, p.ASCII),
			State: pt.State,
		}
	}
	seg.Parts = newParts
	seg.Text = "" // Parts is now authoritative (EffectiveParts prefers Parts)
}
