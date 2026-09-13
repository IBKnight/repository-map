package agents

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// frameworkSignature is a naive, dependency-free check: if manifest is
// present at the repo root and its content contains needle, name is
// reported as a detected framework/library. This is a substring match, not
// a real manifest parser — needles are chosen to avoid common false
// positives (e.g. quoted JSON keys, namespaced module paths) but it's a
// heuristic, not a guarantee.
type frameworkSignature struct {
	manifest string
	needle   string
	name     string
}

var frameworkSignatures = []frameworkSignature{
	// Dart/Flutter
	{"pubspec.yaml", "riverpod", "Riverpod"},
	{"pubspec.yaml", "flutter_bloc", "BLoC"},
	{"pubspec.yaml", "dio:", "Dio"},
	{"pubspec.yaml", "get_it:", "GetIt"},
	{"pubspec.yaml", "go_router:", "GoRouter"},

	// Node.js
	{"package.json", `"react"`, "React"},
	{"package.json", `"next"`, "Next.js"},
	{"package.json", `"vue"`, "Vue"},
	{"package.json", `"@angular/core"`, "Angular"},
	{"package.json", `"svelte"`, "Svelte"},
	{"package.json", `"express"`, "Express"},
	{"package.json", `"@nestjs/core"`, "NestJS"},

	// Python
	{"requirements.txt", "django", "Django"},
	{"requirements.txt", "flask", "Flask"},
	{"requirements.txt", "fastapi", "FastAPI"},
	{"pyproject.toml", "django", "Django"},
	{"pyproject.toml", "flask", "Flask"},
	{"pyproject.toml", "fastapi", "FastAPI"},

	// Go
	{"go.mod", "gin-gonic/gin", "Gin"},
	{"go.mod", "labstack/echo", "Echo"},
	{"go.mod", "gofiber/fiber", "Fiber"},
	{"go.mod", "spf13/cobra", "Cobra"},
	{"go.mod", "gorm.io/gorm", "GORM"},

	// Rust
	{"Cargo.toml", "actix-web", "Actix Web"},
	{"Cargo.toml", "axum", "Axum"},
	{"Cargo.toml", "rocket", "Rocket"},
	{"Cargo.toml", "tokio", "Tokio"},

	// Ruby
	{"Gemfile", "rails", "Rails"},
	{"Gemfile", "sinatra", "Sinatra"},
}

// detectFrameworks looks for known framework/library signatures inside
// root-level manifest files already known to be present (topNames), so
// generated instructions say "Flutter + Riverpod" instead of just "Dart".
func detectFrameworks(rootPath string, topNames map[string]bool) []string {
	content := map[string]string{}
	seen := map[string]bool{}
	var found []string

	for _, sig := range frameworkSignatures {
		if !topNames[sig.manifest] {
			continue
		}

		lower, cached := content[sig.manifest]
		if !cached {
			raw, err := os.ReadFile(filepath.Join(rootPath, sig.manifest))
			if err != nil {
				content[sig.manifest] = ""
				continue
			}
			lower = strings.ToLower(string(raw))
			content[sig.manifest] = lower
		}
		if lower == "" {
			continue
		}

		if strings.Contains(lower, strings.ToLower(sig.needle)) && !seen[sig.name] {
			seen[sig.name] = true
			found = append(found, sig.name)
		}
	}

	sort.Strings(found)
	return found
}
