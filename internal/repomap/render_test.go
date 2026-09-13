package repomap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IBKnight/repository-map/internal/scanner"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRenderMarkdownIncludesLanguagesAndTree(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "main.go"), "package main\n")
	writeTestFile(t, filepath.Join(dir, "pkg", "util.go"), "package pkg\n")
	writeTestFile(t, filepath.Join(dir, "README.md"), "# Hello\n")

	root, err := scanner.Scan(dir, scanner.Options{RespectGitignore: false})
	if err != nil {
		t.Fatal(err)
	}

	out := Render(root, Options{Format: FormatMarkdown})

	if !strings.Contains(out, "Go") {
		t.Errorf("expected Go language in output, got:\n%s", out)
	}
	if !strings.Contains(out, "main.go") {
		t.Errorf("expected main.go listed in tree, got:\n%s", out)
	}
	if !strings.Contains(out, "pkg/") {
		t.Errorf("expected pkg/ directory listed in tree, got:\n%s", out)
	}
}

func TestDetectLanguagesCounts(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "a.go"), "package a\n")
	writeTestFile(t, filepath.Join(dir, "b.go"), "package b\n")
	writeTestFile(t, filepath.Join(dir, "c.py"), "print(1)\n")

	root, err := scanner.Scan(dir, scanner.Options{RespectGitignore: false})
	if err != nil {
		t.Fatal(err)
	}

	stats := DetectLanguages(scanner.Flatten(root))
	if len(stats) != 2 {
		t.Fatalf("expected 2 languages, got %d: %+v", len(stats), stats)
	}
	if stats[0].Language != "Go" || stats[0].Files != 2 {
		t.Errorf("expected Go with 2 files first, got %+v", stats[0])
	}
}
