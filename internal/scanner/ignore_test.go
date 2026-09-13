package scanner

import (
	"path/filepath"
	"testing"
)

func scanNames(t *testing.T, dir string) map[string]bool {
	t.Helper()
	root, err := Scan(dir, Options{RespectGitignore: true})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, n := range Flatten(root) {
		names[n.Path] = true
	}
	return names
}

func TestIgnoreAnchoredPatternOnlyMatchesAtItsLevel(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".gitignore"), "/only-root.log\n")
	mustWrite(t, filepath.Join(dir, "only-root.log"), "x")
	mustWrite(t, filepath.Join(dir, "sub", "only-root.log"), "x")

	names := scanNames(t, dir)
	if names["only-root.log"] {
		t.Error("expected anchored pattern to exclude the root-level file")
	}
	if !names["sub/only-root.log"] {
		t.Error("expected anchored pattern to NOT exclude the nested file")
	}
}

func TestIgnoreUnanchoredPatternMatchesAnyDepth(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".gitignore"), "*.log\n")
	mustWrite(t, filepath.Join(dir, "a.log"), "x")
	mustWrite(t, filepath.Join(dir, "deep", "nested", "b.log"), "x")

	names := scanNames(t, dir)
	if names["a.log"] || names["deep/nested/b.log"] {
		t.Errorf("expected *.log to match at every depth, got %+v", names)
	}
}

func TestIgnoreDoubleStarMiddle(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".gitignore"), "a/**/z.txt\n")
	mustWrite(t, filepath.Join(dir, "a", "z.txt"), "x")
	mustWrite(t, filepath.Join(dir, "a", "b", "c", "z.txt"), "x")
	mustWrite(t, filepath.Join(dir, "a", "keep.txt"), "x")

	names := scanNames(t, dir)
	if names["a/z.txt"] {
		t.Error("expected a/**/z.txt to match a/z.txt (** matches zero dirs)")
	}
	if names["a/b/c/z.txt"] {
		t.Error("expected a/**/z.txt to match a/b/c/z.txt")
	}
	if !names["a/keep.txt"] {
		t.Error("expected unrelated file a/keep.txt to survive")
	}
}

func TestIgnoreNegationReincludes(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".gitignore"), "*.log\n!keep.log\n")
	mustWrite(t, filepath.Join(dir, "a.log"), "x")
	mustWrite(t, filepath.Join(dir, "keep.log"), "x")

	names := scanNames(t, dir)
	if names["a.log"] {
		t.Error("expected a.log to be excluded")
	}
	if !names["keep.log"] {
		t.Error("expected keep.log to be re-included by negation")
	}
}

func TestIgnoreDirOnlyDoesNotMatchFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".gitignore"), "build/\n")
	mustWrite(t, filepath.Join(dir, "build"), "") // a FILE named "build", not a directory
	mustWrite(t, filepath.Join(dir, "keep.txt"), "x")

	names := scanNames(t, dir)
	if !names["build"] {
		t.Error("expected a file named 'build' to survive a dir-only pattern")
	}
}

func TestIgnoreNestedGitignoreIsScopedToItsSubtree(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "sub", ".gitignore"), "local.txt\n")
	mustWrite(t, filepath.Join(dir, "local.txt"), "x")                 // outside sub/, must survive
	mustWrite(t, filepath.Join(dir, "sub", "local.txt"), "x")          // inside sub/, must be excluded
	mustWrite(t, filepath.Join(dir, "sub", "other", "local.txt"), "x") // still under sub/, any depth

	names := scanNames(t, dir)
	if !names["local.txt"] {
		t.Error("expected root-level local.txt to be unaffected by nested .gitignore")
	}
	if names["sub/local.txt"] {
		t.Error("expected sub/local.txt to be excluded by sub/.gitignore")
	}
	if names["sub/other/local.txt"] {
		t.Error("expected sub/other/local.txt to be excluded (unanchored pattern, any depth under sub/)")
	}
}
