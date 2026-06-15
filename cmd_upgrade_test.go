package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
)

type tarEntry struct{ name, body string }

// makeTarGz builds an in-memory .tar.gz from entries, in order.
func makeTarGz(t *testing.T, entries []tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		if err := tw.WriteHeader(&tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.body))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtractBinary(t *testing.T) {
	want := "#!/bin/sh\necho hi\n"
	got, err := extractBinary(makeTarGz(t, []tarEntry{{"cosmobar", want}}))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("extracted = %q, want %q", got, want)
	}
}

func TestExtractBinaryNested(t *testing.T) {
	// GoReleaser archives may carry extra files and a path prefix; the binary is
	// matched by basename, so a nested dist/cosmobar must still be found.
	want := "binary-bytes"
	archive := makeTarGz(t, []tarEntry{{"README.md", "docs"}, {"dist/cosmobar", want}})
	got, err := extractBinary(archive)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("nested extract = %q, want %q", got, want)
	}
}

func TestExtractBinaryNotFound(t *testing.T) {
	archive := makeTarGz(t, []tarEntry{{"README.md", "docs"}, {"LICENSE", "mit"}})
	if _, err := extractBinary(archive); err == nil {
		t.Error("expected error when archive has no cosmobar binary")
	}
}

func TestExtractBinaryNotGzip(t *testing.T) {
	if _, err := extractBinary([]byte("plain text, not gzip at all")); err == nil {
		t.Error("expected error on non-gzip input")
	}
}
