package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanSkipsGitignoredFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".gitignore"), "ignored.txt\nbuild/\n")
	mustWrite(t, filepath.Join(dir, "keep.txt"), "kept")
	mustWrite(t, filepath.Join(dir, "ignored.txt"), "ignored")
	mustWrite(t, filepath.Join(dir, "build", "out.bin"), "bin")

	root, err := Scan(dir, Options{RespectGitignore: true})
	if err != nil {
		t.Fatal(err)
	}

	names := map[string]bool{}
	for _, n := range Flatten(root) {
		names[n.Path] = true
	}

	if !names["keep.txt"] {
		t.Errorf("expected keep.txt to be present, got %+v", names)
	}
	if names["ignored.txt"] {
		t.Errorf("expected ignored.txt to be excluded")
	}
	if names["build"] || names["build/out.bin"] {
		t.Errorf("expected build/ to be excluded, got %+v", names)
	}
}

func TestScanRespectsMaxDepth(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a", "b", "c.txt"), "deep")

	root, err := Scan(dir, Options{MaxDepth: 1})
	if err != nil {
		t.Fatal(err)
	}

	for _, n := range Flatten(root) {
		if n.Depth > 1 {
			t.Errorf("expected max depth 1, found node at depth %d: %s", n.Depth, n.Path)
		}
	}
}

func TestScanDefaultIgnoresGitDir(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".git", "config"), "x")
	mustWrite(t, filepath.Join(dir, "main.go"), "package main")

	root, err := Scan(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}

	for _, n := range Flatten(root) {
		if n.Path == ".git" || n.Path == ".git/config" {
			t.Errorf(".git should always be skipped")
		}
	}
}
