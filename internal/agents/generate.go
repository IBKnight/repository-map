package agents

import (
	"fmt"
	"os"
	"path/filepath"
)

// Target identifies one AI tool/convention that repomap can scaffold
// instruction files for.
type Target string

const (
	TargetClaude  Target = "claude"
	TargetCopilot Target = "copilot"
	TargetCursor  Target = "cursor"
	TargetGeneric Target = "generic"
)

// AllTargets lists every supported target, in generation order.
var AllTargets = []Target{TargetClaude, TargetCopilot, TargetCursor, TargetGeneric}

// file pairs a path (relative to the output directory) with rendered content.
type file struct {
	relPath string
	content string
}

// filesForTarget returns the files that make up one target's instruction set.
func filesForTarget(t Target, a Analysis) ([]file, error) {
	switch t {
	case TargetClaude:
		slug := slugify(a.ProjectName) + "-overview"
		return []file{
			{relPath: "CLAUDE.md", content: renderClaudeMD(a)},
			{relPath: filepath.Join(".claude", "skills", slug, "SKILL.md"), content: renderClaudeSkill(a)},
		}, nil
	case TargetCopilot:
		return []file{
			{relPath: filepath.Join(".github", "copilot-instructions.md"), content: renderCopilotInstructions(a)},
		}, nil
	case TargetCursor:
		return []file{
			{relPath: filepath.Join(".cursor", "rules", "repository.mdc"), content: renderCursorRules(a)},
		}, nil
	case TargetGeneric:
		return []file{
			{relPath: "AGENTS.md", content: renderAgentsMD(a)},
		}, nil
	default:
		return nil, fmt.Errorf("unknown target: %s", t)
	}
}

// GenerateResult reports what happened to each candidate file.
type GenerateResult struct {
	Written []string
	Skipped []string // already existed and force was false
}

// Generate writes instruction files for the requested targets into outputDir.
// Existing files are left untouched unless force is true.
func Generate(a Analysis, targets []Target, outputDir string, force bool) (GenerateResult, error) {
	var result GenerateResult

	for _, t := range targets {
		files, err := filesForTarget(t, a)
		if err != nil {
			return result, err
		}
		for _, f := range files {
			fullPath := filepath.Join(outputDir, f.relPath)

			if !force {
				if _, err := os.Stat(fullPath); err == nil {
					result.Skipped = append(result.Skipped, f.relPath)
					continue
				}
			}

			if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
				return result, err
			}
			if err := os.WriteFile(fullPath, []byte(f.content), 0o644); err != nil {
				return result, err
			}
			result.Written = append(result.Written, f.relPath)
		}
	}

	return result, nil
}

// ParseTargets converts CLI target names to Target values. "all" expands
// to AllTargets.
func ParseTargets(names []string) ([]Target, error) {
	if len(names) == 0 {
		return AllTargets, nil
	}

	valid := map[Target]bool{}
	for _, t := range AllTargets {
		valid[t] = true
	}

	var out []Target
	for _, n := range names {
		if n == "all" {
			return AllTargets, nil
		}
		t := Target(n)
		if !valid[t] {
			return nil, fmt.Errorf("unknown target %q (valid: claude, copilot, cursor, generic, all)", n)
		}
		out = append(out, t)
	}
	return out, nil
}
