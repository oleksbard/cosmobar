// Package render holds low-level rendering primitives: color profiles,
// ANSI coloring, gauges, width math, and layout. It must not import the
// segments package (orchestration lives in the statusline package).
package render

import (
	"fmt"
	"strings"

	"github.com/oleksbard/cosmobar/internal/theme"
)

// Profile is the detected terminal color capability.
type Profile int

const (
	ProfileNone Profile = iota
	Profile256
	ProfileTrueColor
)

// DetectProfile decides the color profile from environment variables.
// env is injected for testability (pass os.Getenv in production).
func DetectProfile(env func(string) string) Profile {
	if env("NO_COLOR") != "" {
		return ProfileNone
	}
	if env("TERM") == "dumb" {
		return ProfileNone
	}
	switch strings.ToLower(env("COLORTERM")) {
	case "truecolor", "24bit":
		return ProfileTrueColor
	}
	return Profile256
}

// Colorize wraps s in an ANSI foreground color for the given profile.
func Colorize(p Profile, c theme.RGB, s string) string {
	switch p {
	case ProfileTrueColor:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", c.R, c.G, c.B, s)
	case Profile256:
		return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", to256(c), s)
	default:
		return s
	}
}

// to256 maps an RGB color into the 6x6x6 color cube (indices 16-231).
func to256(c theme.RGB) int {
	r := int(c.R) * 5 / 255
	g := int(c.G) * 5 / 255
	b := int(c.B) * 5 / 255
	return 16 + 36*r + 6*g + b
}

// Luminance returns Rec.601 perceived luminance in [0,1].
func Luminance(c theme.RGB) float64 {
	return (0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)) / 255.0
}

// Contrast returns dark on light backgrounds, light on dark ones.
func Contrast(bg, dark, light theme.RGB) theme.RGB {
	if Luminance(bg) >= 0.55 {
		return dark
	}
	return light
}

// Fill wraps s with both a foreground and a background color.
func Fill(p Profile, fg, bg theme.RGB, s string) string {
	switch p {
	case ProfileTrueColor:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm%s\x1b[0m", fg.R, fg.G, fg.B, bg.R, bg.G, bg.B, s)
	case Profile256:
		return fmt.Sprintf("\x1b[38;5;%d;48;5;%dm%s\x1b[0m", to256(fg), to256(bg), s)
	default:
		return s
	}
}
