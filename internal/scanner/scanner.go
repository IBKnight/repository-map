// Package scanner walks a repository's file tree, applying .gitignore-style
// exclusion rules, and produces the raw node list used by the repomap and
// agents packages.
package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Node is a single file or directory discovered while walking a repository.
type Node struct {
	Path     string // relative to the scan root, using "/" separators
	Name     string
	IsDir    bool
	Size     int64
	Depth    int
	Children []*Node
}

// Options controls how a repository is scanned.
type Options struct {
	MaxDepth         int // 0 means unlimited
	RespectGitignore bool
}

var defaultIgnoredDirs = map[string]bool{
	".git":          true,
	".hg":           true,
	".svn":          true,
	"node_modules":  true,
	"vendor":        true,
	"dist":          true,
	"build":         true,
	".venv":         true,
	"venv":          true,
	"__pycache__":   true,
	".idea":         true,
	".vscode":       true,
	".DS_Store":     true,
	".dart_tool":    true, // Dart/Flutter build cache
	"target":        true, // Rust (Cargo) / Java (Maven) build output
	".gradle":       true, // Gradle build cache
	"Pods":          true, // CocoaPods (iOS/macOS)
	"DerivedData":   true, // Xcode build output
	".terraform":    true, // Terraform provider/module cache
	"coverage":      true, // test coverage reports (many ecosystems)
	".next":         true, // Next.js build output
	".nuxt":         true, // Nuxt.js build output
	".pytest_cache": true, // pytest cache
	".mypy_cache":   true, // mypy cache
}

// Scan walks root and returns its tree as a single Node.
func Scan(root string, opts Options) (*Node, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	matcher := newIgnoreMatcher(absRoot, opts.RespectGitignore)

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}

	rootNode := &Node{Path: ".", Name: filepath.Base(absRoot), IsDir: info.IsDir(), Depth: 0}
	if !info.IsDir() {
		return rootNode, nil
	}

	if err := walk(absRoot, absRoot, rootNode, matcher, opts); err != nil {
		return nil, err
	}
	return rootNode, nil
}

func walk(root, dir string, node *Node, matcher *ignoreMatcher, opts Options) error {
	if opts.MaxDepth > 0 && node.Depth >= opts.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)
		relPath, _ := filepath.Rel(root, fullPath)
		relPath = filepath.ToSlash(relPath)

		if entry.IsDir() && defaultIgnoredDirs[name] {
			continue
		}
		if matcher.match(relPath, entry.IsDir()) {
			continue
		}

		child := &Node{
			Path:  relPath,
			Name:  name,
			IsDir: entry.IsDir(),
			Depth: node.Depth + 1,
		}

		if !entry.IsDir() {
			if fi, err := entry.Info(); err == nil {
				child.Size = fi.Size()
			}
		}

		node.Children = append(node.Children, child)

		if entry.IsDir() {
			if err := walk(root, fullPath, child, matcher, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Flatten returns every node in the tree (excluding the root) in depth-first order.
func Flatten(root *Node) []*Node {
	var out []*Node
	var visit func(n *Node)
	visit = func(n *Node) {
		for _, c := range n.Children {
			out = append(out, c)
			if c.IsDir {
				visit(c)
			}
		}
	}
	visit(root)
	return out
}

// readLines reads a file into a slice of non-empty, non-comment lines.
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, sc.Err()
}
