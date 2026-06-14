package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

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
	st := Status{InRepo: true, Branch: "feat|x", Ahead: 1, Behind: 2, Staged: 3, Modified: 4, Untracked: 5, Stashes: 6}
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

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
