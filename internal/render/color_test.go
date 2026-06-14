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

func TestLuminanceAndContrast(t *testing.T) {
	white := theme.RGB{R: 255, G: 255, B: 255}
	black := theme.RGB{R: 0, G: 0, B: 0}
	dark := theme.RGB{R: 10, G: 10, B: 20}
	light := theme.RGB{R: 230, G: 230, B: 230}
	if Contrast(white, dark, light) != dark {
		t.Error("light bg should pick dark ink")
	}
	if Contrast(black, dark, light) != light {
		t.Error("dark bg should pick light ink")
	}
}

func TestFill(t *testing.T) {
	fg := theme.RGB{R: 0, G: 0, B: 0}
	bg := theme.RGB{R: 255, G: 0, B: 0}
	if got := Fill(ProfileNone, fg, bg, "x"); got != "x" {
		t.Errorf("None should passthrough: %q", got)
	}
	if got := Fill(ProfileTrueColor, fg, bg, "x"); got != "\x1b[38;2;0;0;0;48;2;255;0;0mx\x1b[0m" {
		t.Errorf("truecolor fill wrong: %q", got)
	}
	if got := Fill(Profile256, fg, bg, "x"); got != "\x1b[38;5;16;48;5;196mx\x1b[0m" {
		t.Errorf("256 fill wrong: %q", got)
	}
}

func TestStateForExported(t *testing.T) {
	if StateFor(50, 70, 90) != StateOK || StateFor(75, 70, 90) != StateWarn || StateFor(95, 70, 90) != StateCrit {
		t.Error("StateFor thresholds wrong")
	}
}
