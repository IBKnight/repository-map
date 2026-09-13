// Package agents inspects a scanned repository and generates instruction
// files for AI coding agents and assistants (Claude Code, GitHub Copilot,
// Cursor, and generic tools that read an AGENTS.md convention).
package agents

import (
	"github.com/IBKnight/repository-map/internal/repomap"
	"github.com/IBKnight/repository-map/internal/scanner"
)

// Ecosystem describes one recognized project toolchain and the commands
// typically used to work with it.
type Ecosystem struct {
	Name     string
	Marker   string
	BuildCmd string
	TestCmd  string
	LintCmd  string
}

// markers is checked in order against the root-level file/directory names.
var markers = []Ecosystem{
	{Name: "Go", Marker: "go.mod", BuildCmd: "go build ./...", TestCmd: "go test ./...", LintCmd: "go vet ./..."},
	{Name: "Node.js", Marker: "package.json", BuildCmd: "npm run build", TestCmd: "npm test", LintCmd: "npm run lint"},
	{Name: "Python (Poetry)", Marker: "pyproject.toml", BuildCmd: "poetry build", TestCmd: "poetry run pytest", LintCmd: "poetry run ruff check ."},
	{Name: "Python", Marker: "requirements.txt", BuildCmd: "pip install -r requirements.txt", TestCmd: "pytest", LintCmd: "ruff check ."},
	{Name: "Rust", Marker: "Cargo.toml", BuildCmd: "cargo build", TestCmd: "cargo test", LintCmd: "cargo clippy"},
	{Name: "Java (Maven)", Marker: "pom.xml", BuildCmd: "mvn package", TestCmd: "mvn test", LintCmd: "mvn checkstyle:check"},
	{Name: "Java (Gradle)", Marker: "build.gradle", BuildCmd: "./gradlew build", TestCmd: "./gradlew test", LintCmd: "./gradlew check"},
	{Name: "Ruby", Marker: "Gemfile", BuildCmd: "bundle install", TestCmd: "bundle exec rspec", LintCmd: "bundle exec rubocop"},
	{Name: "PHP", Marker: "composer.json", BuildCmd: "composer install", TestCmd: "composer test", LintCmd: "composer lint"},
	{Name: "C/C++ (CMake)", Marker: "CMakeLists.txt", BuildCmd: "cmake --build build", TestCmd: "ctest --test-dir build", LintCmd: ""},
	{Name: "Dart/Flutter", Marker: "pubspec.yaml", BuildCmd: "flutter build", TestCmd: "flutter test", LintCmd: "flutter analyze"},
	{Name: "Elixir", Marker: "mix.exs", BuildCmd: "mix compile", TestCmd: "mix test", LintCmd: "mix credo"},
	{Name: "Swift", Marker: "Package.swift", BuildCmd: "swift build", TestCmd: "swift test", LintCmd: ""},
	{Name: "Deno", Marker: "deno.json", BuildCmd: "", TestCmd: "deno test", LintCmd: "deno lint"},
}

// Analysis is the summary handed to the instruction-file templates.
type Analysis struct {
	ProjectName string
	Root        *scanner.Node
	Languages   []repomap.LanguageStat
	Ecosystems  []Ecosystem
	TopLevel    []string
}

// Analyze scans root's already-built tree and detects ecosystems present at
// the top level of the repository.
func Analyze(root *scanner.Node, projectName string) Analysis {
	nodes := scanner.Flatten(root)

	topNames := map[string]bool{}
	for _, c := range root.Children {
		topNames[c.Name] = true
	}

	var found []Ecosystem
	for _, m := range markers {
		if topNames[m.Marker] {
			found = append(found, m)
		}
	}

	return Analysis{
		ProjectName: projectName,
		Root:        root,
		Languages:   repomap.DetectLanguages(nodes),
		Ecosystems:  found,
		TopLevel:    repomap.TopLevelDirs(root),
	}
}
