package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/oleksbard/cosmobar/internal/anim"
	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
	"github.com/oleksbard/cosmobar/internal/spend"
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
	prof := render.DetectProfile(os.Getenv)
	sess := anim.Load(s.SessionID, cfg, prof)
	in := statusline.Input{
		Session: s,
		Git:     git.Lookup(s.SessionID, s.Dir()),
		Config:  cfg,
		Cols:    cols,
		Profile: prof,
		Now:     time.Now(),
		Anim:    sess,
	}
	if s.TranscriptPath != "" && contains(cfg.Order, "tokens") {
		if tk, terr := session.SessionTokens(s.TranscriptPath); terr == nil {
			in.SessionTokens = &tk
		}
	}
	if needSpend(cfg) {
		var resetsAt int64
		if s.RateLimits != nil && s.RateLimits.FiveHour != nil {
			resetsAt = s.RateLimits.FiveHour.ResetsAt
		}
		l := spend.Load(in.Now)
		l.Upsert(s.SessionID, s.Cost.TotalCostUSD, resetsAt)
		in.Spend = &spend.Rollup{Today: l.Today(), Block: l.Block(resetsAt)}
		l.Save()
	}
	out := statusline.Render(in)
	sess.Save()
	return out
}

func cmdPrint(_ []string) int {
	out := renderFromJSON(os.Stdin, envInt("COLUMNS", 80))
	fmt.Println(out)
	return 0
}

// needSpend reports whether any segment that consumes cross-session cost
// (cost rollups or the rate_limits block cost) is enabled.
func needSpend(cfg config.Config) bool {
	return contains(cfg.Order, "cost") || contains(cfg.Order, "rate_limits")
}
