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

func TestTruncateEdgeCases(t *testing.T) {
	if got := Truncate("abc", 0); got != "" {
		t.Errorf("max=0 should be empty, got %q", got)
	}
	if got := Truncate("abc", -5); got != "" {
		t.Errorf("negative max should be empty, got %q", got)
	}
	// Odd keep → asymmetric split: max=4 → keep=3 → left=1, right=2.
	if got := Truncate("abcdefghij", 4); got != "a…ij" {
		t.Errorf("odd-keep split = %q, want %q", got, "a…ij")
	}
	// Invariant: a truncated string is exactly max columns wide.
	for _, max := range []int{1, 2, 3, 7, 9} {
		if w := Width(Truncate("abcdefghijklmnop", max)); w != max {
			t.Errorf("Truncate to %d cols produced width %d", max, w)
		}
	}
}
