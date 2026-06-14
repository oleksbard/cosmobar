package statusline

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/git"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/session"
)

func styledInput(style string, prof render.Profile) Input {
	c := config.Default()
	c.Theme = "catppuccin"
	c.Style = style
	c.Order = []string{"git", "model", "context"}
	pct := 63.0
	s := &session.Session{}
	s.Model.DisplayName = "Opus"
	s.ContextWindow.UsedPercentage = &pct
	return Input{
		Session: s,
		Git:     git.Status{InRepo: true, Branch: "main"},
		Config:  c,
		Cols:    200,
		Profile: prof,
		Now:     time.Date(2026, 6, 14, 14, 32, 0, 0, time.UTC),
	}
}

func TestLeanRoleColors(t *testing.T) {
	out := Render(styledInput("lean", render.ProfileTrueColor))
	// git uses Secondary (catppuccin #cba6f7 = 203,166,247)
	if !strings.Contains(out, "\x1b[38;2;203;166;247mmain\x1b[0m") {
		t.Errorf("git should be secondary-colored: %q", out)
	}
}

func TestBlocksSoftCapsAndFill(t *testing.T) {
	out := Render(styledInput("blocks", render.ProfileTrueColor))
	if !strings.Contains(out, "▐") || !strings.Contains(out, "▌") {
		t.Errorf("blocks soft should include caps: %q", out)
	}
	if !strings.Contains(out, "48;2;") {
		t.Errorf("blocks should set backgrounds: %q", out)
	}
	if !strings.Contains(out, "63%") || strings.Contains(out, "▓") {
		t.Errorf("context should be compact (no bar) in blocks: %q", out)
	}
}

func TestBlocksSquareNoCaps(t *testing.T) {
	in := styledInput("blocks", render.ProfileTrueColor)
	in.Config.BlockCaps = "square"
	out := Render(in)
	if strings.Contains(out, "▐") {
		t.Errorf("square caps should omit half-blocks: %q", out)
	}
}

func TestTickGlyph(t *testing.T) {
	out := Render(styledInput("tick", render.ProfileTrueColor))
	if !strings.Contains(out, "┃") {
		t.Errorf("tick should include ┃: %q", out)
	}
}

func TestNoColorDegradesToPlainLean(t *testing.T) {
	out := Render(styledInput("blocks", render.ProfileNone))
	if strings.Contains(out, "\x1b[") || strings.Contains(out, "▐") {
		t.Errorf("NO_COLOR should be plain lean text: %q", out)
	}
	if !strings.Contains(out, "▓") {
		t.Errorf("NO_COLOR keeps the gauge bar (lean): %q", out)
	}
}

func TestAsciiFallbacks(t *testing.T) {
	in := styledInput("tick", render.ProfileTrueColor)
	in.Config.Glyphs = "ascii"
	if out := Render(in); !strings.Contains(out, "|") || strings.Contains(out, "┃") {
		t.Errorf("ascii tick should use |: %q", out)
	}
	in = styledInput("blocks", render.ProfileTrueColor)
	in.Config.Glyphs = "ascii"
	if out := Render(in); strings.Contains(out, "▐") {
		t.Errorf("ascii blocks should omit half-block caps: %q", out)
	}
}

func TestLinesGreenRedInBlocks(t *testing.T) {
	in := styledInput("blocks", render.ProfileTrueColor)
	in.Config.Order = []string{"lines"}
	in.Git = git.Status{InRepo: true, LinesAdded: 5, LinesRemoved: 2}
	out := Render(in)
	// catppuccin OK green = 166,227,161 ; Crit red = 243,139,168
	if !strings.Contains(out, "48;2;166;227;161") {
		t.Errorf("+added should have green background: %q", out)
	}
	if !strings.Contains(out, "48;2;243;139;168") {
		t.Errorf("-removed should have red background: %q", out)
	}
	// The +N and -N blocks must be flush (one cohesive pill), not separated.
	if vis := stripANSI(out); !strings.Contains(vis, "+5-2") {
		t.Errorf("lines parts should be joined (no inter-part gap), visible = %q", vis)
	}
}

var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }
