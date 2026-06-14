package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
	"github.com/oleksbard/cosmobar/internal/statusline"
)

// renderFromJSON is the testable core of the print command.
func renderFromJSON(r io.Reader, cols int) string {
	s, err := session.Parse(r)
	if err != nil {
		return ""
	}
	cfg, cerr := config.Load(config.DefaultPath())
	if cerr != nil {
		fmt.Fprintln(os.Stderr, "cosmobar: config error:", cerr)
		cfg = config.Default()
	}
	return statusline.Render(statusline.Input{
		Session: s,
		Git:     git.Lookup(s.SessionID, s.Dir()),
		Config:  cfg,
		Cols:    cols,
		Profile: render.DetectProfile(os.Getenv),
		Now:     time.Now(),
	})
}

func cmdPrint(_ []string) int {
	out := renderFromJSON(os.Stdin, envInt("COLUMNS", 80))
	fmt.Println(out)
	return 0
}
