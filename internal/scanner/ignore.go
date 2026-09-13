package scanner

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ignorePattern is one compiled line from a .gitignore file.
//
// baseDir is the directory (relative to the scan root, "/"-separated,
// "" for the root itself) that the pattern was declared in — patterns only
// apply to paths below that directory, matching git's own scoping rules.
type ignorePattern struct {
	negate   bool
	dirOnly  bool
	baseDir  string
	segments []string // pattern split on "/"; "**" is a wildcard segment
}

// ignoreMatcher aggregates every .gitignore found under the scan root so
// that nested rules are honored the same way git itself would apply them:
// patterns are evaluated in root-to-leaf, top-to-bottom order, and the last
// matching pattern for a given path wins (negated patterns re-include it).
type ignoreMatcher struct {
	enabled  bool
	patterns []ignorePattern
}

func newIgnoreMatcher(root string, enabled bool) *ignoreMatcher {
	m := &ignoreMatcher{enabled: enabled}
	if !enabled {
		return m
	}

	for _, giPath := range findGitignoreFiles(root) {
		baseDir := filepath.ToSlash(mustRel(root, filepath.Dir(giPath)))
		if baseDir == "." {
			baseDir = ""
		}

		lines, err := readLines(giPath)
		if err != nil {
			continue
		}
		for _, line := range lines {
			if p, ok := parseIgnoreLine(line, baseDir); ok {
				m.patterns = append(m.patterns, p)
			}
		}
	}

	return m
}

func mustRel(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}

// findGitignoreFiles walks dir (skipping the same directories the scanner
// itself never descends into) and returns every .gitignore file found,
// in root-first, alphabetical order.
func findGitignoreFiles(dir string) []string {
	var found []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return found
	}

	for _, e := range entries {
		if e.IsDir() {
			if defaultIgnoredDirs[e.Name()] {
				continue
			}
			found = append(found, findGitignoreFiles(filepath.Join(dir, e.Name()))...)
			continue
		}
		if e.Name() == ".gitignore" {
			found = append(found, filepath.Join(dir, e.Name()))
		}
	}

	return found
}

// parseIgnoreLine compiles a single non-empty, non-comment .gitignore line
// (as produced by readLines) into an ignorePattern scoped to baseDir.
func parseIgnoreLine(line, baseDir string) (ignorePattern, bool) {
	// Escaped leading '!' or '#' ("\!", "\#") denote a literal pattern
	// character rather than negation/comment; readLines only strips
	// unescaped '#' lines, so unescape here.
	if strings.HasPrefix(line, `\!`) || strings.HasPrefix(line, `\#`) {
		line = line[1:]
	}

	negate := strings.HasPrefix(line, "!")
	if negate {
		line = line[1:]
	}
	if line == "" {
		return ignorePattern{}, false
	}

	dirOnly := strings.HasSuffix(line, "/")
	line = strings.TrimSuffix(line, "/")
	if line == "" {
		return ignorePattern{}, false
	}

	// A slash anywhere but the trailing position anchors the pattern to
	// baseDir; a bare name (no slash) may match at any depth beneath it.
	anchored := strings.Contains(line, "/")
	line = strings.TrimPrefix(line, "/")

	segments := strings.Split(line, "/")
	if !anchored {
		segments = append([]string{"**"}, segments...)
	}

	return ignorePattern{
		negate:   negate,
		dirOnly:  dirOnly,
		baseDir:  baseDir,
		segments: segments,
	}, true
}

// match reports whether relPath (root-relative, "/"-separated) is ignored,
// applying every pattern in order so later, more specific rules win.
func (m *ignoreMatcher) match(relPath string, isDir bool) bool {
	if !m.enabled {
		return false
	}

	ignored := false
	for _, p := range m.patterns {
		pathInScope, ok := trimBase(relPath, p.baseDir)
		if !ok {
			continue
		}
		if p.dirOnly && !isDir {
			continue
		}
		if matchSegments(p.segments, strings.Split(pathInScope, "/")) {
			ignored = !p.negate
		}
	}
	return ignored
}

// trimBase reports whether relPath lies under baseDir and, if so, returns
// relPath relative to baseDir.
func trimBase(relPath, baseDir string) (string, bool) {
	if baseDir == "" {
		return relPath, true
	}
	if rest, ok := strings.CutPrefix(relPath, baseDir+"/"); ok {
		return rest, true
	}
	return "", false
}

// matchSegments matches a gitignore pattern (already split on "/", with
// "**" as a wildcard segment matching zero or more path components)
// against a path (also split on "/"). The full path must be consumed.
func matchSegments(pattern, pathSegs []string) bool {
	if len(pattern) == 0 {
		return len(pathSegs) == 0
	}

	if pattern[0] == "**" {
		if len(pattern) == 1 {
			return true
		}
		for i := 0; i <= len(pathSegs); i++ {
			if matchSegments(pattern[1:], pathSegs[i:]) {
				return true
			}
		}
		return false
	}

	if len(pathSegs) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], pathSegs[0])
	if err != nil || !ok {
		return false
	}
	return matchSegments(pattern[1:], pathSegs[1:])
}
