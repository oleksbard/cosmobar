package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/settings"
)

func cmdInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	theme := fs.String("theme", "", "theme name (see `cosmobar themes`)")
	order := fs.String("order", "", "comma-separated enabled segments, in display order")
	clock := fs.String("clock", "", "clock format: 24h | 12h | off")
	glyphs := fs.String("glyphs", "", "glyphs: auto | unicode | ascii")
	style := fs.String("style", "", "style: lean | tick | blocks")
	caps := fs.String("caps", "", "block caps: soft | square")
	rateWindow := fs.String("rate-window", "", "rate-limit window: both | 5h | 7d")
	animate := fs.String("animate", "", "value-change animation: on | off")
	force := fs.Bool("force", false, "overwrite an existing config file")
	noSkill := fs.Bool("no-skill", false, "do not install the Claude Code setup skill")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	bin, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: cannot resolve own path:", err)
		return 1
	}
	if resolved, err := filepath.EvalSymlinks(bin); err == nil {
		bin = resolved
	}

	sp := settings.Path()
	if err := settings.WireStatusLine(sp, bin, 10); err != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: failed to update settings:", err)
		return 1
	}
	fmt.Println("cosmobar: wired statusLine →", sp, "(command:", bin+")")

	// Build a config from the flags layered over the defaults.
	cfg := config.Default()
	customized := false
	if *theme != "" {
		cfg.Theme = *theme
		customized = true
	}
	if *order != "" {
		cfg.Order = splitCSV(*order)
		customized = true
	}
	if *clock != "" {
		cfg.Clock.Format = *clock
		customized = true
	}
	if *glyphs != "" {
		cfg.Glyphs = *glyphs
		customized = true
	}
	if *style != "" {
		cfg.Style = *style
		customized = true
	}
	if *caps != "" {
		cfg.BlockCaps = *caps
		customized = true
	}
	if *rateWindow != "" {
		cfg.RateLimits.Window = *rateWindow
		customized = true
	}
	if *animate != "" {
		cfg.Animation.Enabled = *animate == "on"
		customized = true
	}
	// Keep per-segment toggles consistent with the enabled order.
	cfg.Context.Show = contains(cfg.Order, "context")
	cfg.RateLimits.Show = contains(cfg.Order, "rate_limits")

	cp := config.DefaultPath()
	_, statErr := os.Stat(cp)
	exists := statErr == nil
	switch {
	case !exists || *force:
		if err := os.MkdirAll(filepath.Dir(cp), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "cosmobar: failed to create config dir:", err)
		} else if err := os.WriteFile(cp, []byte(config.RenderTOML(cfg)), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "cosmobar: failed to write config:", err)
		} else {
			fmt.Println("cosmobar: wrote config →", cp)
		}
	case customized:
		fmt.Println("cosmobar: config already exists; flags ignored (re-run with --force to overwrite):", cp)
	default:
		fmt.Println("cosmobar: existing config left as-is:", cp)
	}

	if !*noSkill {
		if p, err := installSkill(); err == nil {
			fmt.Println("cosmobar: installed setup skill →", p, "(run /cosmobar in Claude Code)")
		}
	}

	fmt.Println("cosmobar: done. Restart Claude Code or send a message to see the status line.")
	return 0
}
