package main

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
	"github.com/oleksbard/cosmobar/internal/session"
	"github.com/oleksbard/cosmobar/internal/statusline"
)

//go:embed assets/mock-session.json
var mockSession []byte

// previewMockGit is the fixed, illustrative git status shared by every preview
// render (static and animated) so the mock data lives in one place.
var previewMockGit = git.Status{InRepo: true, Branch: "main", Ahead: 1, Staged: 1, Modified: 2, Stashes: 1, LinesAdded: 24, LinesRemoved: 7}

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

// previewRender renders the embedded mock session with a fixed, illustrative
// git status so themes/styles/layout can be checked without Claude Code.
func previewRender(o previewOpts) string {
	cfg := previewConfig(o)
	cols := o.cols
	if cols <= 0 {
		cols = 100
	}
	s, _ := session.Parse(bytes.NewReader(mockSession))
	mockTokens := session.TokenUsage{Input: 210_000, Output: 38_000}
	return statusline.Render(statusline.Input{
		Session:       s,
		Git:           previewMockGit,
		Config:        cfg,
		Cols:          cols,
		Profile:       render.DetectProfile(os.Getenv),
		Now:           time.Now(),
		SessionTokens: &mockTokens,
	})
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
	base := time.Date(2026, 6, 15, 14, 32, 0, 0, time.UTC)
	for n := 0; n < loops; n++ {
		start := base.Add(time.Duration(n) * (dur + loopPause))
		sess := anim.Demo(cfg, prof, start)
		// k < stepN keeps every sampled frame in-flight (progress < 1); the
		// settled frame is appended once below, avoiding a duplicate.
		for k := 0; k < stepN; k++ {
			now := start.Add(time.Duration(k) * dur / time.Duration(stepN))
			frames = append(frames, statusline.Render(statusline.Input{
				Session: s, Git: previewMockGit, Config: cfg, Cols: cols, Profile: prof, Now: now, Anim: sess,
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
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *animate {
		playAnimation(o, *loops)
		return 0
	}
	fmt.Println(previewRender(o))
	return 0
}
