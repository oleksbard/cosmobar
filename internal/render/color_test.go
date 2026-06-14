package render

import (
	"testing"

	"github.com/oleksbard/cosmobar/internal/theme"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDetectProfile(t *testing.T) {
	if DetectProfile(env(map[string]string{"NO_COLOR": "1"})) != ProfileNone {
		t.Error("NO_COLOR should disable color")
	}
	if DetectProfile(env(map[string]string{"TERM": "dumb"})) != ProfileNone {
		t.Error("TERM=dumb should disable color")
	}
	if DetectProfile(env(map[string]string{"COLORTERM": "truecolor"})) != ProfileTrueColor {
		t.Error("COLORTERM=truecolor should be truecolor")
	}
	if DetectProfile(env(map[string]string{})) != Profile256 {
		t.Error("default should be 256")
	}
}

func TestColorize(t *testing.T) {
	c := theme.RGB{R: 255, G: 0, B: 0}
	if got := Colorize(ProfileNone, c, "hi"); got != "hi" {
		t.Errorf("ProfileNone should not color: %q", got)
	}
	if got := Colorize(ProfileTrueColor, c, "hi"); got != "\x1b[38;2;255;0;0mhi\x1b[0m" {
		t.Errorf("truecolor escape wrong: %q", got)
	}
	if got := Colorize(Profile256, c, "hi"); got != "\x1b[38;5;196mhi\x1b[0m" {
		t.Errorf("256 escape wrong: %q", got)
	}
}
