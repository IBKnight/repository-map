package agents

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/IBKnight/repository-map/internal/scanner"
)

// CIWorkflow is a single CI config file, with a best-effort display name
// pulled from its own top-level "name:" field when one is present.
type CIWorkflow struct {
	Path string
	Name string
}

// CISystem is one CI provider detected in the repository, with the
// workflow/config files that belong to it.
type CISystem struct {
	Provider  string
	Workflows []CIWorkflow
}

// singleFileCI maps well-known single-file CI configs to their provider
// name, checked in this order for deterministic output.
var singleFileCI = []struct {
	path     string
	provider string
}{
	{".gitlab-ci.yml", "GitLab CI"},
	{filepath.Join(".circleci", "config.yml"), "CircleCI"},
	{"azure-pipelines.yml", "Azure Pipelines"},
	{"Jenkinsfile", "Jenkins"},
	{".travis.yml", "Travis CI"},
	{"bitbucket-pipelines.yml", "Bitbucket Pipelines"},
}

// detectCI finds CI configuration under rootPath so generated instructions
// can point an agent at what's already set up instead of it re-proposing
// CI from scratch. GitHub Actions workflows are listed with a best-effort
// name pulled from each file; other providers are reported by presence.
func detectCI(rootPath string, nodes []*scanner.Node) []CISystem {
	var systems []CISystem

	byPath := map[string]bool{}
	for _, n := range nodes {
		if !n.IsDir {
			byPath[n.Path] = true
		}
	}

	var workflows []CIWorkflow
	for _, n := range nodes {
		if n.IsDir {
			continue
		}
		if strings.HasPrefix(n.Path, ".github/workflows/") && (strings.HasSuffix(n.Name, ".yml") || strings.HasSuffix(n.Name, ".yaml")) {
			workflows = append(workflows, CIWorkflow{
				Path: n.Path,
				Name: readTopLevelYAMLName(filepath.Join(rootPath, n.Path)),
			})
		}
	}
	if len(workflows) > 0 {
		sort.Slice(workflows, func(i, j int) bool { return workflows[i].Path < workflows[j].Path })
		systems = append(systems, CISystem{Provider: "GitHub Actions", Workflows: workflows})
	}

	for _, sf := range singleFileCI {
		if byPath[filepath.ToSlash(sf.path)] {
			systems = append(systems, CISystem{Provider: sf.provider, Workflows: []CIWorkflow{{Path: sf.path}}})
		}
	}

	return systems
}

// readTopLevelYAMLName does a best-effort, dependency-free extraction of a
// top-level (unindented) "name:" field from a YAML file. It's line-based,
// not a real YAML parser, so it only handles the common case.
func readTopLevelYAMLName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "name:") {
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		return strings.Trim(name, `"'`)
	}
	return ""
}
