package render

import "unicode/utf8"

// Width returns the display width of s. cosmobar uses only single-width
// runes (ASCII, block elements, ·, arrows), so rune count is accurate.
func Width(s string) int {
	return utf8.RuneCountInString(s)
}

// Truncate shortens s to at most max display columns using a middle ellipsis.
func Truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	keep := max - 1
	left := keep / 2
	right := keep - left
	return string(r[:left]) + "…" + string(r[len(r)-right:])
}
