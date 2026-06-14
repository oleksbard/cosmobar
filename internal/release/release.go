// Package release implements self-update against GitHub Releases.
package release

import (
	"strconv"
	"strings"
)

const (
	Owner = "oleksbard"
	Repo  = "cosmobar"
)

// AssetName returns the GoReleaser archive name for an OS/arch pair.
func AssetName(goos, goarch string) string {
	return "cosmobar_" + goos + "_" + goarch + ".tar.gz"
}

// IsNewer reports whether latest is a newer semver tag than current.
// A "dev" (or otherwise unparseable) current version is always upgradeable.
func IsNewer(latest, current string) bool {
	lv := parseSemver(latest)
	cv, ok := tryParseSemver(current)
	if !ok {
		return true
	}
	for i := 0; i < 3; i++ {
		if lv[i] != cv[i] {
			return lv[i] > cv[i]
		}
	}
	return false
}

func parseSemver(s string) [3]int {
	v, _ := tryParseSemver(s)
	return v
}

func tryParseSemver(s string) ([3]int, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(strings.SplitN(parts[i], "-", 2)[0])
		if err != nil {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}

// ParseChecksums parses GoReleaser's "checksums.txt" ("<sha256>  <name>" lines).
func ParseChecksums(body string) map[string]string {
	m := map[string]string{}
	for _, line := range strings.Split(body, "\n") {
		f := strings.Fields(line)
		if len(f) == 2 {
			m[f[1]] = f[0]
		}
	}
	return m
}
