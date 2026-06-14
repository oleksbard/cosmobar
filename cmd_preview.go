package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
	"github.com/oleksbard/cosmobar/internal/statusline"
)

//go:embed assets/mock-session.json
var mockSession []byte

// previewRender renders the embedded mock session with a fixed, illustrative
// git status so themes/layout can be checked without Claude Code.
func previewRender(cols int, themeName, cfgPath string) string {
	cfg, _ := config.Load(cfgPath)
	if themeName != "" {
		cfg.Theme = themeName
	}
	s, _ := session.Parse(bytes.NewReader(mockSession))
	mockGit := git.Status{InRepo: true, Branch: "main", Ahead: 1, Staged: 1, Modified: 2, Stashes: 1}
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
	cols := fs.Int("cols", 100, "terminal width to simulate")
	themeName := fs.String("theme", "", "override theme (coral|catppuccin|nord|gruvbox)")
	cfgPath := fs.String("config", config.DefaultPath(), "config file path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	fmt.Println(previewRender(*cols, *themeName, *cfgPath))
	return 0
}
