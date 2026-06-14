package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
	"github.com/oleksbard/cosmobar/internal/statusline"
)

//go:embed assets/mock-session.json
var mockSession []byte

// previewOpts override config fields so a candidate look can be rendered
// without writing the config file first. Empty/zero fields keep the config's
// value. The rendering pipeline is identical to the live status line — only
// the session/git data is mocked — so a preview matches the real output.
type previewOpts struct {
	cols                                          int
	theme, style, caps, glyphs, clock, rateWindow string
	order, cfgPath                                string
}

// previewRender renders the embedded mock session with a fixed, illustrative
// git status so themes/styles/layout can be checked without Claude Code.
func previewRender(o previewOpts) string {
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
	cols := o.cols
	if cols <= 0 {
		cols = 100
	}
	s, _ := session.Parse(bytes.NewReader(mockSession))
	mockGit := git.Status{InRepo: true, Branch: "main", Ahead: 1, Staged: 1, Modified: 2, Stashes: 1, LinesAdded: 24, LinesRemoved: 7}
	return statusline.Render(statusline.Input{
		Session: s,
		Git:     mockGit,
		Config:  cfg,
		Cols:    cols,
		Profile: render.DetectProfile(os.Getenv),
		Now:     time.Now(),
	})
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
	if err := fs.Parse(args); err != nil {
		return 2
	}
	fmt.Println(previewRender(o))
	return 0
}
