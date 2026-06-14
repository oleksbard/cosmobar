// Package git gathers working-tree status via the git CLI, cached per session.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Status is the subset of git state cosmobar renders.
type Status struct {
	InRepo       bool
	Branch       string
	Ahead        int
	Behind       int
	Staged       int
	Modified     int
	Untracked    int
	Stashes      int
	LinesAdded   int
	LinesRemoved int
}

const cacheTTL = time.Second

// Lookup returns git status for dir, caching the result per session for ~1s
// in the OS temp dir. Errors (not a repo, git missing) yield a zero Status.
func Lookup(sessionID, dir string) Status {
	cache := filepath.Join(os.TempDir(), "cosmobar-git-"+sanitize(sessionID))
	if st, ok := readCache(cache); ok {
		return st
	}
	st := collect(dir)
	writeCache(cache, st)
	return st
}

func collect(dir string) Status {
	out, err := run(dir, "status", "--porcelain=v2", "--branch")
	if err != nil {
		return Status{}
	}
	st := parseStatus(out)
	st.InRepo = true
	if s, err := run(dir, "stash", "list"); err == nil {
		st.Stashes = countLines(s)
	}
	if s, err := run(dir, "diff", "HEAD", "--numstat"); err == nil {
		st.LinesAdded, st.LinesRemoved = parseNumstat(s)
	} else {
		// No HEAD yet (pre-first-commit): count unstaged + staged vs the empty tree.
		a1, r1 := 0, 0
		if s, err := run(dir, "diff", "--numstat"); err == nil {
			a1, r1 = parseNumstat(s)
		}
		a2, r2 := 0, 0
		if s, err := run(dir, "diff", "--cached", "--numstat"); err == nil {
			a2, r2 = parseNumstat(s)
		}
		st.LinesAdded, st.LinesRemoved = a1+a2, r1+r2
	}
	return st
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out strings.Builder
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func parseStatus(out string) Status {
	st := Status{}
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			st.Branch = strings.TrimPrefix(line, "# branch.head ")
		case strings.HasPrefix(line, "# branch.ab "):
			fmt.Sscanf(strings.TrimPrefix(line, "# branch.ab "), "+%d -%d", &st.Ahead, &st.Behind)
		case strings.HasPrefix(line, "1 "), strings.HasPrefix(line, "2 "):
			f := strings.Fields(line)
			if len(f) >= 2 && len(f[1]) == 2 {
				if f[1][0] != '.' {
					st.Staged++
				}
				if f[1][1] != '.' {
					st.Modified++
				}
			}
		case strings.HasPrefix(line, "? "):
			st.Untracked++
		}
	}
	return st
}

func countLines(s string) int {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// cache format: InRepo|Ahead|Behind|Staged|Modified|Untracked|Stashes|LinesAdded|LinesRemoved|Branch
// Branch is last and joined with SplitN so '|' in a branch name is preserved.
func writeCache(path string, st Status) {
	line := fmt.Sprintf("%t|%d|%d|%d|%d|%d|%d|%d|%d|%s",
		st.InRepo, st.Ahead, st.Behind, st.Staged, st.Modified, st.Untracked, st.Stashes,
		st.LinesAdded, st.LinesRemoved, st.Branch)
	_ = os.WriteFile(path, []byte(line), 0o600)
}

func readCache(path string) (Status, bool) {
	info, err := os.Stat(path)
	if err != nil || time.Since(info.ModTime()) > cacheTTL {
		return Status{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Status{}, false
	}
	p := strings.SplitN(strings.TrimSpace(string(data)), "|", 10)
	if len(p) != 10 {
		return Status{}, false
	}
	st := Status{InRepo: p[0] == "true", Branch: p[9]}
	st.Ahead, _ = strconv.Atoi(p[1])
	st.Behind, _ = strconv.Atoi(p[2])
	st.Staged, _ = strconv.Atoi(p[3])
	st.Modified, _ = strconv.Atoi(p[4])
	st.Untracked, _ = strconv.Atoi(p[5])
	st.Stashes, _ = strconv.Atoi(p[6])
	st.LinesAdded, _ = strconv.Atoi(p[7])
	st.LinesRemoved, _ = strconv.Atoi(p[8])
	return st, true
}

func parseNumstat(out string) (added, removed int) {
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		if n, err := strconv.Atoi(f[0]); err == nil {
			added += n
		}
		if n, err := strconv.Atoi(f[1]); err == nil {
			removed += n
		}
	}
	return added, removed
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "default"
	}
	return b.String()
}
