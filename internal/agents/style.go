package agents

import "sort"

// styleConfigFiles are root-level filenames that encode formatting/lint
// conventions an agent should follow rather than re-derive by reading
// source and guessing.
var styleConfigFiles = []string{
	".editorconfig",
	".prettierrc",
	".prettierrc.json",
	".prettierrc.yaml",
	".prettierrc.yml",
	".prettierrc.js",
	".eslintrc",
	".eslintrc.json",
	".eslintrc.js",
	".eslintrc.cjs",
	".eslintrc.yaml",
	".eslintrc.yml",
	".golangci.yml",
	".golangci.yaml",
	"rustfmt.toml",
	".rustfmt.toml",
	".flake8",
	".rubocop.yml",
	".stylelintrc",
	".stylelintrc.json",
}

// detectStyleConfigs returns which known formatting/lint config files exist
// at the repository root, so generated instructions can tell an agent to
// match them instead of inferring style conventions from scratch.
func detectStyleConfigs(topNames map[string]bool) []string {
	var found []string
	for _, f := range styleConfigFiles {
		if topNames[f] {
			found = append(found, f)
		}
	}
	sort.Strings(found)
	return found
}
