package agents

import (
	"path/filepath"
	"sort"

	"github.com/IBKnight/repository-map/internal/scanner"
)

// maxSubProjects caps how many nested manifests are reported, so a large
// monorepo doesn't blow up the generated file.
const maxSubProjects = 10

// detectSubProjects finds ecosystem manifest files below the repository
// root (e.g. apps/web/package.json, services/api/go.mod) and reports their
// containing directories, so an agent working in a monorepo knows where the
// independently-buildable units live without spending a turn on `find`.
func detectSubProjects(nodes []*scanner.Node) []string {
	markerNames := map[string]bool{}
	for _, m := range markers {
		markerNames[m.Marker] = true
	}

	seen := map[string]bool{}
	var dirs []string
	for _, n := range nodes {
		if n.IsDir || n.Path == "." {
			continue
		}
		if !markerNames[n.Name] {
			continue
		}
		dir := filepath.Dir(n.Path)
		if dir == "." {
			continue // root-level manifests are already reported as Ecosystems
		}
		dir = filepath.ToSlash(dir)
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}

	sort.Strings(dirs)
	if len(dirs) > maxSubProjects {
		dirs = dirs[:maxSubProjects]
	}
	return dirs
}
