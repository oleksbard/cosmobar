package spend

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestTodayCountsOnlyTodaysDelta is the regression test for the bug where a
// long-running session's whole lifetime cumulative was attributed to "today".
func TestTodayCountsOnlyTodaysDelta(t *testing.T) {
	day1 := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "ledger.json")

	// Day 1: a session is first seen at $100 cumulative, then spends $10.
	l := load(path, day1)
	l.Upsert("sess", 100.0, 0) // first sight → baseline, contributes 0 today
	l.Upsert("sess", 110.0, 0) // +$10 today
	if got := l.Today(); got != 10.0 {
		t.Fatalf("day1 Today() = %.2f, want 10.00", got)
	}
	l.Save()

	// Day 2: the SAME session continues to $115 cumulative. Today must be the
	// $5 spent today — NOT the $115 lifetime total (that was the bug).
	l2 := load(path, day2)
	l2.Upsert("sess", 115.0, 0)
	if got := l2.Today(); got != 5.0 {
		t.Errorf("day2 Today() = %.2f, want 5.00 (only today's delta, not lifetime)", got)
	}
}

// TestTodaySumsMultipleSessions confirms today aggregates the per-session
// today-deltas across sessions, not their cumulative totals.
func TestTodaySumsMultipleSessions(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	l := load("", now)
	l.Upsert("a", 50.0, 0) // first sight → 0
	l.Upsert("a", 53.0, 0) // +$3
	l.Upsert("b", 8.0, 0)  // first sight → 0
	l.Upsert("b", 8.5, 0)  // +$0.50
	if got := l.Today(); got != 3.5 {
		t.Errorf("Today() = %.2f, want 3.50", got)
	}
}

// TestFirstSightAttributesNoSpend confirms a session already mid-life when
// first observed contributes nothing until it spends more (we never attribute
// pre-existing cumulative cost to any day).
func TestFirstSightAttributesNoSpend(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	l := load("", now)
	l.Upsert("pre-existing", 50.0, 0)
	if got := l.Today(); got != 0.0 {
		t.Errorf("first-sight Today() = %.2f, want 0", got)
	}
}

func TestTodayExcludesOtherDays(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	l := load("", now)
	l.Upsert("today-sess", 3.00, 0) // first sight → 0 today
	l.Upsert("today-sess", 4.00, 0) // +$1 today
	// A session last seen yesterday contributes nothing to today.
	l.entries["yesterday-sess"] = entry{Cost: 99, Day: "2026-06-24", DayBaseCost: 0, Ts: now.Add(-24 * time.Hour).Unix()}
	if got := l.Today(); got != 1.00 {
		t.Errorf("Today() = %.2f, want 1.00 (other days excluded)", got)
	}
}

func TestUpsertIgnoresBlankSession(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	l := load("", now)
	l.Upsert("", 5.00, 0)
	if len(l.entries) != 0 {
		t.Errorf("blank session id should be ignored, got %d entries", len(l.entries))
	}
}

// TestBlockCountsOnlyCurrentWindowDelta confirms block cost is the spend within
// the current 5h window, and that moving to a new window (changed resets_at)
// rebaselines instead of carrying the lifetime total.
func TestBlockCountsOnlyCurrentWindowDelta(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	windowA := now.Add(2 * time.Hour).Unix()
	l := load("", now)

	l.Upsert("s", 100.0, windowA) // first sight in window A → block baseline 100
	l.Upsert("s", 104.0, windowA) // +$4 in window A
	if got := l.Block(windowA); got != 4.0 {
		t.Fatalf("Block(windowA) = %.2f, want 4.00", got)
	}

	// A new 5h window opens (different resets_at): the block baseline resets.
	windowB := now.Add(7 * time.Hour).Unix()
	l.Upsert("s", 109.0, windowB) // +$5 since the new window started
	if got := l.Block(windowB); got != 5.0 {
		t.Errorf("Block(windowB) = %.2f, want 5.00 (not lifetime)", got)
	}
	if got := l.Block(windowA); got != 0.0 {
		t.Errorf("Block(old windowA) = %.2f, want 0 (session moved to windowB)", got)
	}
	if got := l.Block(0); got != 0.0 {
		t.Errorf("Block(0) = %.2f, want 0 (no reset time)", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "ledger.json")

	l := load(path, now)
	l.Upsert("sess-a", 4.20, 0) // first sight → baseline
	l.Upsert("sess-a", 6.20, 0) // +$2 today
	l.Save()

	l2 := load(path, now)
	if got := l2.Today(); got != 2.00 {
		t.Errorf("after reload Today() = %.2f, want 2.00", got)
	}
}

func TestLoadPrunesOldEntries(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "ledger.json")

	seed := load(path, now)
	seed.entries["fresh"] = entry{Cost: 1.00, Day: "2026-06-25", Ts: now.Unix()}
	seed.entries["ancient"] = entry{Cost: 99.00, Day: "2026-01-01", Ts: now.Add(-100 * 24 * time.Hour).Unix()}
	seed.Save()

	reloaded := load(path, now)
	if _, ok := reloaded.entries["ancient"]; ok {
		t.Error("ancient entry should have been pruned on load")
	}
	if _, ok := reloaded.entries["fresh"]; !ok {
		t.Error("fresh entry should survive pruning")
	}
}

func TestLoadCorruptFileYieldsEmpty(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	l := load(path, now)
	if got := l.Today(); got != 0 {
		t.Errorf("corrupt file Today() = %.2f, want 0", got)
	}
}

func TestLoadUsesDefaultPath(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	l := Load(now) // exported entry: resolves the real cache path
	if l == nil {
		t.Fatal("Load returned nil")
	}
	if l.entries == nil {
		t.Error("Load should initialize entries map")
	}
}
