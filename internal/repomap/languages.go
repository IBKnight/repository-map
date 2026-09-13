package repomap

import (
	"path/filepath"
	"strings"

	"github.com/IBKnight/repository-map/internal/scanner"
)

// extToLanguage maps common file extensions to a human-readable language name.
var extToLanguage = map[string]string{
	".go":         "Go",
	".py":         "Python",
	".js":         "JavaScript",
	".jsx":        "JavaScript (JSX)",
	".ts":         "TypeScript",
	".tsx":        "TypeScript (TSX)",
	".java":       "Java",
	".kt":         "Kotlin",
	".rb":         "Ruby",
	".rs":         "Rust",
	".c":          "C",
	".h":          "C Header",
	".cpp":        "C++",
	".hpp":        "C++ Header",
	".cs":         "C#",
	".php":        "PHP",
	".swift":      "Swift",
	".m":          "Objective-C",
	".scala":      "Scala",
	".sh":         "Shell",
	".bash":       "Shell",
	".zsh":        "Shell",
	".sql":        "SQL",
	".html":       "HTML",
	".css":        "CSS",
	".scss":       "SCSS",
	".vue":        "Vue",
	".lua":        "Lua",
	".ex":         "Elixir",
	".exs":        "Elixir",
	".erl":        "Erlang",
	".hs":         "Haskell",
	".clj":        "Clojure",
	".r":          "R",
	".jl":         "Julia",
	".dart":       "Dart",
	".yaml":       "YAML",
	".yml":        "YAML",
	".json":       "JSON",
	".toml":       "TOML",
	".md":         "Markdown",
	".proto":      "Protocol Buffers",
	".tf":         "Terraform",
	".dockerfile": "Docker",
}

// LanguageStat holds an aggregate of files and bytes for one detected language.
type LanguageStat struct {
	Language string
	Files    int
	Bytes    int64
}

// DetectLanguages summarizes file extensions found in nodes into per-language
// counts, sorted by file count descending.
func DetectLanguages(nodes []*scanner.Node) []LanguageStat {
	stats := map[string]*LanguageStat{}

	for _, n := range nodes {
		if n.IsDir {
			continue
		}
		lang, ok := languageForFile(n.Name)
		if !ok {
			continue
		}
		s, exists := stats[lang]
		if !exists {
			s = &LanguageStat{Language: lang}
			stats[lang] = s
		}
		s.Files++
		s.Bytes += n.Size
	}

	out := make([]LanguageStat, 0, len(stats))
	for _, s := range stats {
		out = append(out, *s)
	}
	sortLanguageStats(out)
	return out
}

func languageForFile(name string) (string, bool) {
	lower := strings.ToLower(name)
	if lower == "dockerfile" {
		return "Docker", true
	}
	ext := filepath.Ext(lower)
	if ext == "" {
		return "", false
	}
	lang, ok := extToLanguage[ext]
	return lang, ok
}

func sortLanguageStats(stats []LanguageStat) {
	for i := 1; i < len(stats); i++ {
		for j := i; j > 0 && stats[j].Files > stats[j-1].Files; j-- {
			stats[j], stats[j-1] = stats[j-1], stats[j]
		}
	}
}
