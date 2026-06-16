package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/oleksbard/cosmobar/internal/config"
	"github.com/oleksbard/cosmobar/internal/render"
	"github.com/oleksbard/cosmobar/internal/settings"
)

func check(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

// doctorReport runs offline diagnostics and returns a human-readable report.
func doctorReport() string {
	var b strings.Builder

	// git availability
	_, gitErr := exec.LookPath("git")
	fmt.Fprintf(&b, "%s git: %s\n", check(gitErr == nil), gitAvail(gitErr))

	// config validity
	cfgPath := config.DefaultPath()
	_, cfgErr := config.Load(cfgPath)
	fmt.Fprintf(&b, "%s config: %s (%s)\n", check(cfgErr == nil), cfgStatus(cfgErr), cfgPath)

	// color support
	prof := render.DetectProfile(os.Getenv)
	fmt.Fprintf(&b, "%s color: %s\n", check(prof != render.ProfileNone), profileName(prof))

	// statusLine wiring
	sp := settings.Path()
	wired, cmd := statusLineWired(sp)
	fmt.Fprintf(&b, "%s statusLine: %s\n", check(wired), wireStatus(wired, cmd, sp))

	return b.String()
}

func gitAvail(err error) string {
	if err == nil {
		return "found"
	}
	return "not found (git segment will be hidden)"
}

func cfgStatus(err error) string {
	if err == nil {
		return "valid"
	}
	return "ERROR: " + err.Error()
}

func profileName(p render.Profile) string {
	switch p {
	case render.ProfileTrueColor:
		return "truecolor"
	case render.Profile256:
		return "256-color"
	default:
		return "disabled (NO_COLOR or dumb terminal)"
	}
}

func statusLineWired(path string) (bool, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, ""
	}
	var m map[string]any
	if json.Unmarshal(data, &m) != nil {
		return false, ""
	}
	sl, ok := m["statusLine"].(map[string]any)
	if !ok {
		return false, ""
	}
	cmd, _ := sl["command"].(string)
	return cmd != "", cmd
}

func wireStatus(wired bool, cmd, path string) string {
	if !wired {
		return "not configured — run `cosmobar init` (" + path + ")"
	}
	return cmd
}

func cmdDoctor(_ []string) int {
	fmt.Print(doctorReport())
	return 0
}
