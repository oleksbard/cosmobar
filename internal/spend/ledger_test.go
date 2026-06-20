package spend

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpsertOverwritesAndTodaySums(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	l := load("", now) // empty path → in-memory only

	l.Upsert("sess-a", 4.20)
	l.Upsert("sess-b", 1.10)
	l.Upsert("sess-a", 4.55) // same session refreshes: replace, not add

	if got := l.Today(); got != 5.65 {
		t.Errorf("Today() = %.2f, want 5.65", got)
	}
}

func TestTodayExcludesOtherDays(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	l := load("", now)
	l.Upsert("today-sess", 3.00)
	// Manually inject an entry stamped a different day.
	l.entries["old-sess"] = entry{Cost: 99.00, Day: "2026-06-19", Ts: now.Add(-24 * time.Hour).Unix()}

	if got := l.Today(); got != 3.00 {
		t.Errorf("Today() = %.2f, want 3.00 (other days excluded)", got)
	}
}

func TestUpsertIgnoresBlankSession(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	l := load("", now)
	l.Upsert("", 5.00)
	if got := l.Today(); got != 0 {
		t.Errorf("Today() = %.2f, want 0 (blank id ignored)", got)
	}
}

func TestWeekMonthBlock(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	l := load("", now)
	// today (in week, month, and block)
	l.entries["a"] = entry{Cost: 2.00, Day: "2026-06-20", Ts: now.Unix()}
	// 3 days ago (in week + month, NOT in 5h block)
	l.entries["b"] = entry{Cost: 3.00, Day: "2026-06-17", Ts: now.Add(-3 * 24 * time.Hour).Unix()}
	// 10 days ago (in month only)
	l.entries["c"] = entry{Cost: 5.00, Day: "2026-06-10", Ts: now.Add(-10 * 24 * time.Hour).Unix()}
	// last month (in none of week/month)
	l.entries["d"] = entry{Cost: 9.00, Day: "2026-05-30", Ts: now.Add(-21 * 24 * time.Hour).Unix()}

	if got := l.Week(); got != 5.00 { // a + b
		t.Errorf("Week() = %.2f, want 5.00", got)
	}
	if got := l.Month(); got != 10.00 { // a + b + c
		t.Errorf("Month() = %.2f, want 10.00", got)
	}

	// Block: 5h window ending at resetsAt. Put resetsAt 2h in the future so the
	// window is [now-3h, now+2h]; only "a" (at now) falls inside.
	resetsAt := now.Add(2 * time.Hour).Unix()
	if got := l.Block(resetsAt); got != 2.00 {
		t.Errorf("Block() = %.2f, want 2.00", got)
	}
	if got := l.Block(0); got != 0 {
		t.Errorf("Block(0) = %.2f, want 0 (no reset time)", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "ledger.json")

	l := load(path, now)
	l.Upsert("sess-a", 4.20)
	l.Save()

	l2 := load(path, now)
	if got := l2.Today(); got != 4.20 {
		t.Errorf("after reload Today() = %.2f, want 4.20", got)
	}
}

func TestLoadPrunesOldEntries(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "ledger.json")

	seed := load(path, now)
	seed.entries["fresh"] = entry{Cost: 1.00, Day: "2026-06-20", Ts: now.Unix()}
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
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
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
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	l := Load(now) // exported entry: resolves the real cache path
	if l == nil {
		t.Fatal("Load returned nil")
	}
	if l.entries == nil {
		t.Error("Load should initialize entries map")
	}
}
