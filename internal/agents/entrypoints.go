package agents

import (
	"sort"

	"github.com/IBKnight/repository-map/internal/scanner"
)

// entryPointNames are file basenames conventionally used as a program's
// starting point, across common languages/ecosystems.
var entryPointNames = map[string]bool{
	"main.go":     true,
	"main.rs":     true,
	"main.py":     true,
	"manage.py":   true,
	"__main__.py": true,
	"index.js":    true,
	"index.ts":    true,
	"index.jsx":   true,
	"index.tsx":   true,
	"main.dart":   true,
	"Program.cs":  true,
}

// maxEntryPoints caps how many entry points are reported, so a monorepo
// with many cmd/*/main.go binaries doesn't blow up the generated file.
const maxEntryPoints = 8

// detectEntryPoints finds files matching a known entry-point convention,
// so an agent knows where to start reading instead of guessing.
func detectEntryPoints(nodes []*scanner.Node) []string {
	var found []string
	for _, n := range nodes {
		if n.IsDir {
			continue
		}
		if entryPointNames[n.Name] {
			found = append(found, n.Path)
		}
	}

	sort.Strings(found)
	if len(found) > maxEntryPoints {
		found = found[:maxEntryPoints]
	}
	return found
}
