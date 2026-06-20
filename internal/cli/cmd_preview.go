package cli

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/oleksbard/cosmobar/internal/anim"
	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/segments"
	"github.com/oleksbard/cosmobar/internal/session"
	"github.com/oleksbard/cosmobar/internal/spend"
	"github.com/oleksbard/cosmobar/internal/statusline"
)

//go:embed assets/mock-session.json
var mockSession []byte

// previewMockGit is the fixed, illustrative git status shared by every preview
// render (static and animated) so the mock data lives in one place.
var previewMockGit = git.Status{InRepo: true, Branch: "main", Ahead: 1, Staged: 1, Modified: 2, Stashes: 1, LinesAdded: 24, LinesRemoved: 7}

// previewClock is a fixed, illustrative instant shared by the gallery and the
// animation so showcase output (e.g. the clock segment) is deterministic.
var previewClock = time.Date(2026, 6, 15, 14, 32, 0, 0, time.UTC)

// previewMockSpend is the fixed, illustrative cross-session cost shown by every
// preview render so the cost rollups and 5h block cost are demonstrable.
var previewMockSpend = spend.Rollup{Today: 5.30, Week: 42.10, Month: 118.00, Block: 4.20}

// previewOpts override config fields so a candidate look can be rendered
// without writing the config file first. Empty/zero fields keep the config's
// value. The rendering pipeline is identical to the live status line — only
// the session/git data is mocked — so a preview matches the real output.
type previewOpts struct {
	cols                                          int
	theme, style, caps, glyphs, clock, rateWindow string
	order, cfgPath                                string
}

func previewConfig(o previewOpts) config.Config {
	cfg, _ := config.Load(o.cfgPath)
	if o.theme != "" {
		cfg.Theme = o.theme
	}
	if o.style != "" {
		cfg.Style = o.style
	}
	if o.caps != "" {
		cfg.BlockCaps = o.caps
	}
	if o.glyphs != "" {
		cfg.Glyphs = o.glyphs
	}
	if o.clock != "" {
		cfg.Clock.Format = o.clock
	}
	if o.rateWindow != "" {
		cfg.RateLimits.Window = o.rateWindow
	}
	if o.order != "" {
		var names []string
		for _, n := range strings.Split(o.order, ",") {
			if n = strings.TrimSpace(n); n != "" {
				names = append(names, n)
			}
		}
		cfg.Order = names
	}
	return cfg
}

// renderPreview renders the embedded mock session + previewMockGit for a given
// config, width, and instant. previewRender and the gallery both flow through
// it, so every preview matches the exact live render pipeline.
func renderPreview(cfg config.Config, cols int, now time.Time) string {
	if cols <= 0 {
		cols = 100
	}
	s, _ := session.Parse(bytes.NewReader(mockSession))
	// resets_at is absent from the static mock JSON; inject future resets
	// relative to now so previews showcase the rate-limit countdown.
	if s.RateLimits != nil {
		if s.RateLimits.FiveHour != nil {
			s.RateLimits.FiveHour.ResetsAt = now.Add(2*time.Hour + 30*time.Minute).Unix()
		}
		if s.RateLimits.SevenDay != nil {
			s.RateLimits.SevenDay.ResetsAt = now.Add(72 * time.Hour).Unix()
		}
	}
	mockTokens := session.TokenUsage{Input: 210_000, Output: 38_000}
	return statusline.Render(statusline.Input{
		Session:       s,
		Git:           previewMockGit,
		Config:        cfg,
		Cols:          cols,
		Profile:       render.DetectProfile(os.Getenv),
		Now:           now,
		SessionTokens: &mockTokens,
		Spend:         &previewMockSpend,
	})
}

// previewRender renders the mock session for the candidate config in o, so a
// theme/style/layout can be checked without Claude Code.
func previewRender(o previewOpts) string {
	return renderPreview(previewConfig(o), o.cols, time.Now())
}

// galleryPreset is one labeled example in `preview --gallery`: a theme, style,
// and block caps paired with an explicit segment set. Each renders on top of
// config.Default() so the showcase is stable regardless of the user's config.
type galleryPreset struct {
	name, theme, style, caps string
	order                    []string
}

// catalogOrder returns every catalog segment name in display order, so the
// "Everything" preset stays in sync as segments are added or removed.
func catalogOrder() []string {
	metas := segments.Catalog()
	out := make([]string, len(metas))
	for i, m := range metas {
		out[i] = m.Name
	}
	return out
}

// galleryPresets are the curated examples shown by `preview --gallery`. Across
// the set they exercise all four themes, all three styles, both block caps, and
// segment sets from minimal to the full catalog.
var galleryPresets = []galleryPreset{
	{name: "Minimal", theme: "coral", style: "lean", order: []string{"dir", "git", "model", "clock"}},
	{name: "Coder", theme: "nord", style: "tick", order: []string{"dir", "git", "lines", "context", "cost"}},
	{name: "Pro/Max", theme: "catppuccin", style: "blocks", caps: "soft", order: []string{"model", "context", "rate_limits", "cost", "tokens"}},
	{name: "Everything", theme: "gruvbox", style: "blocks", caps: "square", order: catalogOrder()},
}

// renderGallery renders every galleryPreset as a labeled block — a header naming
// the preset/theme/style, then the mock status line beneath it. The global
// --cols and --glyphs flags in o still apply to every example.
func renderGallery(o previewOpts) string {
	var b strings.Builder
	for i, p := range galleryPresets {
		if i > 0 {
			b.WriteString("\n\n")
		}
		cfg := config.Default()
		cfg.Theme = p.theme
		cfg.Style = p.style
		if p.caps != "" {
			cfg.BlockCaps = p.caps
		}
		cfg.Order = p.order
		cfg.MaxRows = 4            // headroom so "Everything" shows all its segments
		cfg.RateLimits.Show = true // surface rate_limits in presets that include it
		if o.glyphs != "" {
			cfg.Glyphs = o.glyphs
		}
		fmt.Fprintf(&b, "%s · %s\n", p.theme, p.style)
		b.WriteString(renderPreview(cfg, o.cols, previewClock))
	}
	b.WriteString("\n\n")
	b.WriteString(renderWidgetCatalog(o))
	return b.String()
}

// renderWidgetCatalog renders every catalog segment on its own labeled line,
// all in the default theme and style — a reference for what each widget looks
// like in isolation.
func renderWidgetCatalog(o previewOpts) string {
	def := config.Default()
	var b strings.Builder
	fmt.Fprintf(&b, "all widgets · %s · %s\n", def.Theme, def.Style)
	for i, m := range segments.Catalog() {
		cfg := config.Default()
		cfg.Order = []string{m.Name}
		cfg.RateLimits.Show = true // so rate_limits renders here too
		if o.glyphs != "" {
			cfg.Glyphs = o.glyphs
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "%-13s%s", m.Name, renderPreview(cfg, o.cols, previewClock))
	}
	return b.String()
}

// animationFrames renders the mock line `loops` times as a scramble that
// settles, returning each redraw frame plus a final static frame.
func animationFrames(o previewOpts, loops int) []string {
	if loops < 1 {
		loops = 1
	}
	cfg := previewConfig(o)
	cfg.Animation.Enabled = true
	cols := o.cols
	if cols <= 0 {
		cols = 100
	}
	prof := render.DetectProfile(os.Getenv)
	s, _ := session.Parse(bytes.NewReader(mockSession))
	dur := time.Duration(cfg.Animation.DurationMs) * time.Millisecond
	const stepN = 14                         // in-flight samples across one scramble sweep
	const loopPause = 400 * time.Millisecond // dwell between loops

	var frames []string
	base := previewClock
	for n := 0; n < loops; n++ {
		start := base.Add(time.Duration(n) * (dur + loopPause))
		sess := anim.Demo(cfg, prof, start)
		// k < stepN keeps every sampled frame in-flight (progress < 1); the
		// settled frame is appended once below, avoiding a duplicate.
		for k := 0; k < stepN; k++ {
			now := start.Add(time.Duration(k) * dur / time.Duration(stepN))
			frames = append(frames, statusline.Render(statusline.Input{
				Session: s, Git: previewMockGit, Config: cfg, Cols: cols, Profile: prof, Now: now, Anim: sess, Spend: &previewMockSpend,
			}))
		}
	}
	frames = append(frames, previewRender(o)) // final settled line
	return frames
}

// playAnimation prints frames in place, redrawing over the previous frame.
// It clears however many rows the previous frame occupied, so multi-row
// renders (narrow --cols) don't leave the top row stranded.
func playAnimation(o previewOpts, loops int) {
	frames := animationFrames(o, loops)
	const delay = 55 * time.Millisecond // ~18fps redraw cadence
	prevRows := 0
	for i, f := range frames {
		if prevRows > 1 {
			// Cursor is on the last row of the previous frame: go to column 0,
			// move up to its first row, then erase to end of screen.
			fmt.Printf("\r\x1b[%dA\x1b[J", prevRows-1)
		} else {
			fmt.Print("\r\x1b[2K")
		}
		fmt.Print(f)
		prevRows = strings.Count(f, "\n") + 1
		if i < len(frames)-1 {
			time.Sleep(delay)
		}
	}
	fmt.Println()
}

func cmdPreview(args []string) int {
	fs := flag.NewFlagSet("preview", flag.ContinueOnError)
	o := previewOpts{}
	fs.IntVar(&o.cols, "cols", 100, "terminal width to simulate")
	fs.StringVar(&o.theme, "theme", "", "override theme (coral|catppuccin|nord|gruvbox)")
	fs.StringVar(&o.style, "style", "", "override style (lean|tick|blocks)")
	fs.StringVar(&o.caps, "caps", "", "override block caps (soft|square)")
	fs.StringVar(&o.glyphs, "glyphs", "", "override glyphs (auto|ascii)")
	fs.StringVar(&o.clock, "clock", "", "override clock format (24h|12h|off)")
	fs.StringVar(&o.rateWindow, "rate-window", "", "override rate-limit window (both|5h|7d)")
	fs.StringVar(&o.order, "order", "", "override segment order (comma-separated)")
	fs.StringVar(&o.cfgPath, "config", config.DefaultPath(), "config file path")
	animate := fs.Bool("animate", false, "play the value-change animation in the terminal")
	loops := fs.Int("loops", 3, "number of animation loops when --animate is set")
	gallery := fs.Bool("gallery", false, "render a gallery of example presets (all themes + varied widget sets)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *gallery {
		fmt.Println(renderGallery(o))
		return 0
	}
	if *animate {
		playAnimation(o, *loops)
		return 0
	}
	fmt.Println(previewRender(o))
	return 0
}
