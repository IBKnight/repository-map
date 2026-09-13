package scanner

import (
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

// ignoreMatcher aggregates every .gitignore found under the scan root so
// that nested rules are honored the same way git itself would apply them.
type ignoreMatcher struct {
	enabled  bool
	compiled *gitignore.GitIgnore
}

func newIgnoreMatcher(root string, enabled bool) *ignoreMatcher {
	m := &ignoreMatcher{enabled: enabled}
	if !enabled {
		return m
	}

	lines := collectGitignoreLines(root)
	if len(lines) == 0 {
		return m
	}
	m.compiled = gitignore.CompileIgnoreLines(lines...)
	return m
}

// collectGitignoreLines reads every .gitignore file under root and rewrites
// its patterns to be relative to root, so a single compiled matcher can be
// used for the whole tree regardless of which directory a rule came from.
func collectGitignoreLines(root string) []string {
	var all []string

	gitignorePaths := findGitignoreFiles(root)
	for _, path := range gitignorePaths {
		dir := filepath.Dir(path)
		relDir, _ := filepath.Rel(root, dir)
		relDir = filepath.ToSlash(relDir)

		lines, err := readLines(path)
		if err != nil {
			continue
		}
		for _, line := range lines {
			all = append(all, rebaseGitignorePattern(line, relDir))
		}
	}

	return all
}

func findGitignoreFiles(root string) []string {
	var found []string
	_ = walkForGitignore(root, &found)
	return found
}

func walkForGitignore(dir string, found *[]string) error {
	entries, err := readDirNames(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		full := filepath.Join(dir, e.name)
		if e.isDir {
			if defaultIgnoredDirs[e.name] {
				continue
			}
			_ = walkForGitignore(full, found)
			continue
		}
		if e.name == ".gitignore" {
			*found = append(*found, full)
		}
	}
	return nil
}

type dirEntryLite struct {
	name  string
	isDir bool
}

func readDirNames(dir string) ([]dirEntryLite, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]dirEntryLite, 0, len(entries))
	for _, e := range entries {
		out = append(out, dirEntryLite{name: e.Name(), isDir: e.IsDir()})
	}
	return out, nil
}

// rebaseGitignorePattern rewrites a pattern found in a nested .gitignore so
// it applies relative to the scan root instead of that file's directory.
func rebaseGitignorePattern(pattern, relDir string) string {
	if relDir == "." || relDir == "" {
		return pattern
	}
	negate := false
	p := pattern
	if len(p) > 0 && p[0] == '!' {
		negate = true
		p = p[1:]
	}
	p = "/" + filepath.ToSlash(filepath.Join(relDir, p))
	if negate {
		p = "!" + p
	}
	return p
}

func (m *ignoreMatcher) match(relPath string, isDir bool) bool {
	if !m.enabled || m.compiled == nil {
		return false
	}
	if isDir {
		return m.compiled.MatchesPath(relPath + "/")
	}
	return m.compiled.MatchesPath(relPath)
}
