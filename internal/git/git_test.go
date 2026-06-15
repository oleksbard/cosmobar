package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestParseNumstat(t *testing.T) {
	a, r := parseNumstat("12\t3\tfile.go\n0\t5\tother.go\n-\t-\tbin.png\n")
	if a != 12 || r != 8 {
		t.Errorf("numstat sum = +%d -%d, want +12 -8", a, r)
	}
}

func TestLinesResetAfterCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "user.email", "t@t")
	mustGit(t, dir, "config", "user.name", "t")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644)
	mustGit(t, dir, "add", "a.txt")
	mustGit(t, dir, "commit", "-m", "init")

	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\nthree\n"), 0o644)
	st := collect(dir)
	if st.LinesAdded != 2 || st.LinesRemoved != 0 {
		t.Errorf("after edit = +%d -%d, want +2 -0", st.LinesAdded, st.LinesRemoved)
	}
	mustGit(t, dir, "commit", "-am", "edit")
	st = collect(dir)
	if st.LinesAdded != 0 || st.LinesRemoved != 0 {
		t.Errorf("after commit = +%d -%d, want 0/0", st.LinesAdded, st.LinesRemoved)
	}
}

func TestParseStatus(t *testing.T) {
	out := "# branch.head main\n" +
		"# branch.ab +2 -1\n" +
		"1 M. N... 100644 100644 100644 aaa bbb staged.go\n" +
		"1 .M N... 100644 100644 100644 aaa bbb modified.go\n" +
		"? untracked.go\n"
	st := parseStatus(out)
	if st.Branch != "main" {
		t.Errorf("branch = %q", st.Branch)
	}
	if st.Ahead != 2 || st.Behind != 1 {
		t.Errorf("ahead/behind = %d/%d", st.Ahead, st.Behind)
	}
	if st.Staged != 1 || st.Modified != 1 || st.Untracked != 1 {
		t.Errorf("counts = %d/%d/%d", st.Staged, st.Modified, st.Untracked)
	}
}

func TestParseStatusRenameDetachedNoUpstream(t *testing.T) {
	// Detached HEAD (branch shows "(detached)"), no "# branch.ab" line (no
	// upstream → 0 ahead/behind), a renamed file ("2 " entry, which the earlier
	// test never exercised) that must count as staged, plus one modified file.
	out := "# branch.oid abc123\n" +
		"# branch.head (detached)\n" +
		"2 R. N... 100644 100644 100644 aaa bbb R100 new.go\told.go\n" +
		"1 .M N... 100644 100644 100644 aaa bbb mod.go\n"
	st := parseStatus(out)
	if st.Branch != "(detached)" {
		t.Errorf("detached branch = %q, want %q", st.Branch, "(detached)")
	}
	if st.Ahead != 0 || st.Behind != 0 {
		t.Errorf("no upstream should yield 0/0, got %d/%d", st.Ahead, st.Behind)
	}
	if st.Staged != 1 {
		t.Errorf("renamed file should count as staged, got %d", st.Staged)
	}
	if st.Modified != 1 {
		t.Errorf("modified count = %d, want 1", st.Modified)
	}
}

func TestCountLines(t *testing.T) {
	if countLines("") != 0 {
		t.Error("empty")
	}
	if countLines("a\nb\n") != 2 {
		t.Error("two")
	}
	if countLines("a\nb") != 2 {
		t.Error("no trailing newline")
	}
}

func TestCacheRoundTripAndTTL(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cache")
	st := Status{InRepo: true, Branch: "feat|x", Ahead: 1, Behind: 2, Staged: 3, Modified: 4, Untracked: 5, Stashes: 6, LinesAdded: 7, LinesRemoved: 8}
	writeCache(p, st)
	got, ok := readCache(p)
	if !ok {
		t.Fatal("fresh cache should read")
	}
	if got != st {
		t.Errorf("roundtrip = %+v, want %+v", got, st)
	}
	// make it stale
	old := time.Now().Add(-2 * cacheTTL)
	os.Chtimes(p, old, old)
	if _, ok := readCache(p); ok {
		t.Error("stale cache should not read")
	}
}

func TestCollectAgainstRealRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "user.email", "t@t")
	mustGit(t, dir, "config", "user.name", "t")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi"), 0o644)
	mustGit(t, dir, "add", "a.txt")
	mustGit(t, dir, "commit", "-m", "init")
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("new"), 0o644)
	st := collect(dir)
	if !st.InRepo || st.Branch != "main" {
		t.Errorf("collect = %+v", st)
	}
	if st.Untracked != 1 {
		t.Errorf("expected 1 untracked, got %d", st.Untracked)
	}
}

func TestCollectNotARepo(t *testing.T) {
	if st := collect(t.TempDir()); st.InRepo {
		t.Error("temp dir is not a repo")
	}
}

func TestLinesNoHeadCountsStaged(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	mustGit(t, dir, "init", "-b", "main")
	mustGit(t, dir, "config", "user.email", "t@t")
	mustGit(t, dir, "config", "user.name", "t")
	// Staged content with no commit yet (no HEAD).
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\nthree\n"), 0o644)
	mustGit(t, dir, "add", "a.txt")
	st := collect(dir)
	if st.LinesAdded != 3 {
		t.Errorf("no-HEAD staged: added = %d, want 3", st.LinesAdded)
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
