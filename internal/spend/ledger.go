// Package spend maintains a small persistent ledger of per-session cost,
// keyed by session id, so the status line can show cross-session totals
// (cost spent today, and cost spent in the current 5-hour block). It stores
// Claude's own authoritative total_cost_usd per session, so no model-pricing
// table is needed. The ledger lives in the user cache dir (not the temp dir
// like anim/git) so daily totals survive a reboot.
//
// total_cost_usd is the session's *lifetime cumulative* cost, and Claude Code
// sessions routinely live for days. So the ledger does NOT report the raw
// cumulative — it tracks per-session baselines (the cumulative at the start of
// today, and at the start of the current 5h block) and reports only the delta
// since each baseline. That way a five-day-old session contributes only what
// it spent today / this block, never its whole lifetime.
package spend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const dayFmt = "2006-01-02"

// entry is one session's last-seen cumulative cost plus the baselines used to
// derive period spend. Today's spend = Cost - DayBaseCost; current-block spend
// = Cost - BlockBaseCost (when BlockResetsAt matches the live window).
type entry struct {
	Cost          float64 `json:"cost"`            // last-seen lifetime cumulative total_cost_usd
	Day           string  `json:"day"`             // local date of the last refresh
	DayBaseCost   float64 `json:"day_base"`        // cumulative as of the start of Day
	BlockBaseCost float64 `json:"block_base"`      // cumulative as of the start of the current block
	BlockResetsAt int64   `json:"block_resets_at"` // resets_at the block baseline belongs to
	Ts            int64   `json:"ts"`              // unix seconds of the last refresh (for prune)
}

// Rollup is the injected, render-ready set of cross-session totals.
type Rollup struct {
	Today, Block float64
}

// Ledger is the in-memory view of the on-disk ledger plus the instant ("now")
// all queries are evaluated against, so rendering is deterministic.
type Ledger struct {
	path    string
	now     time.Time
	entries map[string]entry
}

// Upsert records sessionID's current cumulative cost for the current 5h window
// (resetsAt, 0 if unknown), maintaining the day and block baselines so that
// only newly-spent cost is attributed to today / this block.
//
// On first sight a session is baselined to its current cumulative (we count
// only spend cosmobar actually observes accruing, never pre-existing lifetime
// cost). On a day rollover the day baseline advances to the prior cumulative;
// on a new block (changed resetsAt) the block baseline does the same. A blank
// session id is ignored.
func (l *Ledger) Upsert(sessionID string, cost float64, resetsAt int64) {
	if sessionID == "" {
		return
	}
	today := l.now.Format(dayFmt)
	e, seen := l.entries[sessionID]
	if !seen {
		l.entries[sessionID] = entry{
			Cost:          cost,
			Day:           today,
			DayBaseCost:   cost,
			BlockBaseCost: cost,
			BlockResetsAt: resetsAt,
			Ts:            l.now.Unix(),
		}
		return
	}
	if e.Day != today { // new day → everything so far is "previous days"
		e.DayBaseCost = e.Cost
		e.Day = today
	}
	if resetsAt != e.BlockResetsAt { // new 5h window → reset the block baseline
		e.BlockBaseCost = e.Cost
		e.BlockResetsAt = resetsAt
	}
	e.Cost = cost
	e.Ts = l.now.Unix()
	l.entries[sessionID] = e
}

// Today sums each session's spend on now's local calendar date (its cumulative
// minus its start-of-day baseline). Sessions last seen on a prior day did not
// spend today and are skipped.
func (l *Ledger) Today() float64 {
	today := l.now.Format(dayFmt)
	var sum float64
	for _, e := range l.entries {
		if e.Day == today {
			if d := e.Cost - e.DayBaseCost; d > 0 {
				sum += d
			}
		}
	}
	return sum
}

// Block sums each session's spend within the current 5-hour window, identified
// by resetsAt (unix seconds, from rate_limits.five_hour.resets_at): the session's
// cumulative minus its block baseline, for sessions whose baseline belongs to
// this same window. Returns 0 when resetsAt is absent.
func (l *Ledger) Block(resetsAt int64) float64 {
	if resetsAt <= 0 {
		return 0
	}
	var sum float64
	for _, e := range l.entries {
		if e.BlockResetsAt == resetsAt {
			if d := e.Cost - e.BlockBaseCost; d > 0 {
				sum += d
			}
		}
	}
	return sum
}

// pruneDays bounds the ledger: entries older than this are dropped on load.
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
