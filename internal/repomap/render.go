// Package repomap turns a scanned repository tree into human-readable
// output: an ASCII directory tree, a language breakdown, and a Markdown
// report suitable for committing alongside a project or feeding to an
// AI coding agent as context.
package repomap

import (
	"fmt"
	"sort"
	"strings"

	"github.com/IBKnight/repository-map/internal/scanner"
)

// Format selects the output format for the repository map.
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatText     Format = "text"
)

// Options controls how the report is rendered.
type Options struct {
	Title     string
	Format    Format
	TreeDepth int // 0 = unlimited, otherwise caps the rendered tree depth
	ShowSizes bool
}

// Render builds the full report for root according to opts.
func Render(root *scanner.Node, opts Options) string {
	nodes := scanner.Flatten(root)
	langs := DetectLanguages(nodes)

	switch opts.Format {
	case FormatText:
		return renderText(root, nodes, langs, opts)
	default:
		return renderMarkdown(root, nodes, langs, opts)
	}
}

func renderMarkdown(root *scanner.Node, nodes []*scanner.Node, langs []LanguageStat, opts Options) string {
	var b strings.Builder

	title := opts.Title
	if title == "" {
		title = root.Name
	}
	fmt.Fprintf(&b, "# Repository Map: %s\n\n", title)

	dirCount, fileCount := countKinds(nodes)
	fmt.Fprintf(&b, "- **Files:** %d\n", fileCount)
	fmt.Fprintf(&b, "- **Directories:** %d\n", dirCount)
	if len(langs) > 0 {
		fmt.Fprintf(&b, "- **Primary language:** %s\n", langs[0].Language)
	}
	b.WriteString("\n")

	if len(langs) > 0 {
		b.WriteString("## Languages\n\n")
		b.WriteString("| Language | Files | Bytes |\n")
		b.WriteString("|---|---:|---:|\n")
		for _, l := range langs {
			fmt.Fprintf(&b, "| %s | %d | %s |\n", l.Language, l.Files, humanBytes(l.Bytes))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Directory Structure\n\n")
	b.WriteString("```\n")
	b.WriteString(renderTree(root, opts))
	b.WriteString("```\n")

	return b.String()
}

func renderText(root *scanner.Node, nodes []*scanner.Node, langs []LanguageStat, opts Options) string {
	var b strings.Builder

	dirCount, fileCount := countKinds(nodes)
	fmt.Fprintf(&b, "%s\n", root.Name)
	fmt.Fprintf(&b, "files: %d  directories: %d\n\n", fileCount, dirCount)

	if len(langs) > 0 {
		b.WriteString("Languages:\n")
		for _, l := range langs {
			fmt.Fprintf(&b, "  %-20s %5d files  %10s\n", l.Language, l.Files, humanBytes(l.Bytes))
		}
		b.WriteString("\n")
	}

	b.WriteString(renderTree(root, opts))
	return b.String()
}

func renderTree(root *scanner.Node, opts Options) string {
	var b strings.Builder
	b.WriteString(".\n")
	writeTreeChildren(&b, root, "", opts)
	return b.String()
}

func writeTreeChildren(b *strings.Builder, node *scanner.Node, prefix string, opts Options) {
	if opts.TreeDepth > 0 && node.Depth >= opts.TreeDepth {
		return
	}

	children := node.Children
	for i, c := range children {
		last := i == len(children)-1
		connector := "├── "
		nextPrefix := prefix + "│   "
		if last {
			connector = "└── "
			nextPrefix = prefix + "    "
		}

		name := c.Name
		if c.IsDir {
			name += "/"
		} else if opts.ShowSizes {
			name += fmt.Sprintf(" (%s)", humanBytes(c.Size))
		}

		fmt.Fprintf(b, "%s%s%s\n", prefix, connector, name)

		if c.IsDir {
			writeTreeChildren(b, c, nextPrefix, opts)
		}
	}
}

func countKinds(nodes []*scanner.Node) (dirs, files int) {
	for _, n := range nodes {
		if n.IsDir {
			dirs++
		} else {
			files++
		}
	}
	return
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(n)/float64(div), units[exp])
}

// TopLevelDirs returns the names of the immediate subdirectories of root,
// sorted alphabetically. Useful for building a quick project overview.
func TopLevelDirs(root *scanner.Node) []string {
	var out []string
	for _, c := range root.Children {
		if c.IsDir {
			out = append(out, c.Name)
		}
	}
	sort.Strings(out)
	return out
}
