package render

import "testing"

func TestWidth(t *testing.T) {
	if Width("abc") != 3 {
		t.Error("ascii width")
	}
	if Width("▓▓·") != 3 {
		t.Error("unicode width")
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("abcdef", 10); got != "abcdef" {
		t.Errorf("no truncation needed: %q", got)
	}
	if got := Truncate("abcdefghij", 5); got != "ab…ij" {
		t.Errorf("middle ellipsis: %q", got)
	}
	if got := Truncate("abc", 1); got != "…" {
		t.Errorf("min width: %q", got)
	}
}
