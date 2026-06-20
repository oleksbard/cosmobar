// Package spend maintains a small persistent ledger of per-session cost,
// keyed by session id, so the status line can show cross-session totals
// (today / rolling week / calendar month / current 5-hour block). It stores
// Claude's own authoritative total_cost_usd per session, so no model-pricing
// table is needed. The ledger lives in the user cache dir (not the temp dir
// like anim/git) so daily totals survive a reboot.
package spend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const dayFmt = "2006-01-02"

// entry is one session's last-known cumulative cost and when it was last seen.
type entry struct {
	Cost float64 `json:"cost"`
	Day  string  `json:"day"` // local calendar date of the last refresh
	Ts   int64   `json:"ts"`  // unix seconds of the last refresh
}

// Rollup is the injected, render-ready set of cross-session totals.
type Rollup struct {
	Today, Week, Month, Block float64
}

// Ledger is the in-memory view of the on-disk ledger plus the instant ("now")
// all queries are evaluated against, so rendering is deterministic.
type Ledger struct {
	path    string
	now     time.Time
	entries map[string]entry
}

// Upsert records sessionID's current cumulative cost. total_cost_usd is
// cumulative per session, so the entry is overwritten, not accumulated. A
// blank session id is ignored.
func (l *Ledger) Upsert(sessionID string, cost float64) {
	if sessionID == "" {
		return
	}
	l.entries[sessionID] = entry{Cost: cost, Day: l.now.Format(dayFmt), Ts: l.now.Unix()}
}

// Today sums the cost of every session last seen on now's local calendar date.
func (l *Ledger) Today() float64 {
	today := l.now.Format(dayFmt)
	var sum float64
	for _, e := range l.entries {
		if e.Day == today {
			sum += e.Cost
		}
	}
	return sum
}

// pruneDays bounds the ledger: entries older than this are dropped on load.
// 35 days keeps a full calendar month of history.
const pruneDays = 35

// load builds a Ledger from the file at path (empty/missing/corrupt → empty),
// evaluated against now, pruning entries older than pruneDays. An empty path
// means in-memory only (used by tests).
func load(path string, now time.Time) *Ledger {
	l := &Ledger{path: path, now: now, entries: map[string]entry{}}
	if path == "" {
		return l
	}
	if data, err := os.ReadFile(path); err == nil {
		var m map[string]entry
		if json.Unmarshal(data, &m) == nil && m != nil {
			cutoff := now.Add(-pruneDays * 24 * time.Hour).Unix()
			for id, e := range m {
				if e.Ts >= cutoff {
					l.entries[id] = e
				}
			}
		}
	}
	return l
}

// Save writes the ledger to disk (best-effort). No-op when path is empty.
func (l *Ledger) Save() {
	if l.path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return
	}
	if data, err := json.Marshal(l.entries); err == nil {
		_ = os.WriteFile(l.path, data, 0o600)
	}
}

// Week sums sessions last seen within the rolling previous 7×24 hours.
func (l *Ledger) Week() float64 {
	cutoff := l.now.Add(-7 * 24 * time.Hour).Unix()
	var sum float64
	for _, e := range l.entries {
		if e.Ts >= cutoff {
			sum += e.Cost
		}
	}
	return sum
}

// Month sums sessions last seen in now's calendar month (YYYY-MM prefix match).
func (l *Ledger) Month() float64 {
	prefix := l.now.Format("2006-01")
	var sum float64
	for _, e := range l.entries {
		if len(e.Day) >= 7 && e.Day[:7] == prefix {
			sum += e.Cost
		}
	}
	return sum
}

// Block sums sessions last seen inside the current 5-hour window, which ends at
// resetsAt (unix seconds, from rate_limits.five_hour.resets_at). Returns 0 when
// resetsAt is absent.
func (l *Ledger) Block(resetsAt int64) float64 {
	if resetsAt <= 0 {
		return 0
	}
	start := resetsAt - 5*60*60
	var sum float64
	for _, e := range l.entries {
		if e.Ts >= start && e.Ts <= resetsAt {
			sum += e.Cost
		}
	}
	return sum
}

// ledgerPath returns the on-disk ledger location, or "" if no cache dir.
func ledgerPath() string {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		return ""
	}
	return filepath.Join(base, "cosmobar", "ledger.json")
}

// Load reads the ledger from the user cache dir, evaluated against now.
func Load(now time.Time) *Ledger {
	return load(ledgerPath(), now)
}
