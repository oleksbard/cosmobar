package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionTokens(t *testing.T) {
	got, err := SessionTokens(filepath.Join("testdata", "transcript.jsonl"))
	if err != nil {
		t.Fatalf("SessionTokens error: %v", err)
	}
	if got.Input != 350 || got.Output != 35 {
		t.Errorf("got %+v, want {Input:350 Output:35}", got)
	}
	if got.Total() != 385 {
		t.Errorf("Total() = %d, want 385", got.Total())
	}
}

func TestSessionTokensMissingFile(t *testing.T) {
	if _, err := SessionTokens(filepath.Join("testdata", "nope.jsonl")); err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestSessionTokensEmptyMessageIDNotDeduped(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.jsonl")
	line := `{"type":"assistant","requestId":"r1","message":{"usage":{"input_tokens":10,"output_tokens":2}}}` + "\n"
	if err := os.WriteFile(p, []byte(line+line), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SessionTokens(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Input != 20 || got.Output != 4 {
		t.Errorf("got %+v, want {Input:20 Output:4} (empty-id entries must NOT be deduped)", got)
	}
}

func TestSessionTokensEmptyFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "empty.jsonl")
	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := SessionTokens(p)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got.Input != 0 || got.Output != 0 {
		t.Errorf("got %+v, want zero", got)
	}
}
