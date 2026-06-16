package statusline

import (
	"strings"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/segments"
	"github.com/oleksbard/cosmobar/internal/theme"
)

// run is a colored text fragment. hasBG controls whether bg is applied.
type run struct {
	text  string
	fg    theme.RGB
	bg    theme.RGB
	hasBG bool
}

func runsWidth(rs []run) int {
	w := 0
	for _, r := range rs {
		w += render.Width(r.text)
	}
	return w
}

func emit(p render.Profile, rs []run) string {
	var b strings.Builder
	for _, r := range rs {
		if r.text == "" {
			continue
		}
		switch {
		case r.hasBG:
			b.WriteString(render.Fill(p, r.fg, r.bg, r.text))
		case strings.TrimSpace(r.text) == "":
			b.WriteString(r.text) // whitespace gaps need no color escape
		default:
			b.WriteString(render.Colorize(p, r.fg, r.text))
		}
	}
	return b.String()
}

// effectiveStyle resolves the style after degradation rules.
func effectiveStyle(cfg config.Config, prof render.Profile) string {
	if prof == render.ProfileNone {
		return "lean" // no color → no backgrounds/marks
	}
	switch cfg.Style {
	case "tick", "blocks":
		return cfg.Style
	default:
		return "lean"
	}
}

// partColor: state color if set, else the role color.
func partColor(pal theme.Palette, role string, st render.State) theme.RGB {
	switch st {
	case render.StateOK:
		return pal.OK
	case render.StateWarn:
		return pal.Warn
	case render.StateCrit:
		return pal.Crit
	}
	switch role {
	case "vcs":
		return pal.Secondary
	case "metric":
		return pal.Tertiary
	case "usage":
		return pal.Quaternary
	case "ambient":
		return pal.Muted
	default: // identity / gauge-without-state / unknown
		return pal.Accent
	}
}

func softCaps(cfg config.Config) (string, string) {
	if cfg.ASCII() {
		return "", "" // square in ascii
	}
	return "▐", "▌"
}

func tickGlyph(cfg config.Config) string {
	if cfg.ASCII() {
		return "|"
	}
	return "┃"
}

// segmentRuns builds the content runs for one segment under the active style
// (no inter-segment separator).
func segmentRuns(seg segments.Segment, role, style string, pal theme.Palette, cfg config.Config) []run {
	parts := seg.EffectiveParts()
	var rs []run
	switch style {
	case "blocks":
		// A segment renders as ONE contiguous pill: parts sit flush against
		// each other (e.g. lines = +N|-N) so they read as a single unit, with
		// soft caps only on the outer edges. No inner caps, no inter-part gap.
		l, r := softCaps(cfg)
		soft := cfg.BlockCaps != "square" && l != ""
		if soft {
			rs = append(rs, run{text: l, fg: partColor(pal, role, parts[0].State)})
		}
		for i, pt := range parts {
			c := partColor(pal, role, pt.State)
			fg := render.Contrast(c, pal.Dark, pal.Light)
			text := pt.Text
			if !soft { // square caps pad only the outer edges of the pill
				if i == 0 {
					text = " " + text
				}
				if i == len(parts)-1 {
					text += " "
				}
			}
			rs = append(rs, run{text: text, fg: fg, bg: c, hasBG: true})
		}
		if soft {
			rs = append(rs, run{text: r, fg: partColor(pal, role, parts[len(parts)-1].State)})
		}
	case "tick":
		rs = append(rs, run{text: tickGlyph(cfg), fg: partColor(pal, role, parts[0].State)})
		for i, pt := range parts {
			rs = append(rs, run{text: pt.Text, fg: partColor(pal, role, pt.State)})
			if i < len(parts)-1 {
				rs = append(rs, run{text: " "})
			}
		}
	default: // lean
		for i, pt := range parts {
			rs = append(rs, run{text: pt.Text, fg: partColor(pal, role, pt.State)})
			if i < len(parts)-1 {
				rs = append(rs, run{text: " "})
			}
		}
	}
	return rs
}

func separatorWidth(cfg config.Config, style string) int {
	switch style {
	case "blocks":
		return 1
	case "tick":
		return 2
	default:
		return render.Width(cfg.Separator)
	}
}

func separatorRuns(cfg config.Config, style string, pal theme.Palette) []run {
	switch style {
	case "blocks":
		return []run{{text: " "}}
	case "tick":
		return []run{{text: "  "}}
	default:
		return []run{{text: cfg.Separator, fg: pal.Muted}}
	}
}

// emitRow renders one row of segment indices.
func emitRow(p render.Profile, style string, cfg config.Config, pal theme.Palette, segRuns [][]run, row []int) string {
	var b strings.Builder
	sep := separatorRuns(cfg, style, pal)
	for j, idx := range row {
		if j > 0 {
			b.WriteString(emit(p, sep))
		}
		b.WriteString(emit(p, segRuns[idx]))
	}
	return b.String()
}
