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
	ProjectName  string
	Root         *scanner.Node
	Languages    []repomap.LanguageStat
	Ecosystems   []Ecosystem
	TopLevel     []string
	EntryPoints  []string
	CI           []CISystem
	Frameworks   []string
	EnvExample   string
	SubProjects  []string
	StyleConfigs []string
}

// Analyze scans root's already-built tree and detects ecosystems, entry
// points, CI configuration, and key frameworks/dependencies. rootPath is
// the filesystem path that was scanned; select manifest and workflow files
// under it are read directly for the content-based checks (CI workflow
// names, framework signatures) that a filename alone can't answer.
func Analyze(rootPath string, root *scanner.Node, projectName string) Analysis {
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
	applyNodePackageManager(found, topNames)

	return Analysis{
		ProjectName:  projectName,
		Root:         root,
		Languages:    repomap.DetectLanguages(nodes),
		Ecosystems:   found,
		TopLevel:     repomap.TopLevelDirs(root),
		EntryPoints:  detectEntryPoints(nodes),
		CI:           detectCI(rootPath, nodes),
		Frameworks:   detectFrameworks(rootPath, topNames),
		EnvExample:   detectEnvExample(topNames),
		SubProjects:  detectSubProjects(nodes),
		StyleConfigs: detectStyleConfigs(topNames),
	}
}

// nodeLockfiles maps a root-level lockfile to the package manager it implies,
// checked in this order so a repo with more than one (e.g. a leftover
// package-lock.json in a yarn project) still prefers the more specific one.
var nodeLockfiles = []struct {
	file    string
	manager string
}{
	{"bun.lockb", "bun"},
	{"pnpm-lock.yaml", "pnpm"},
	{"yarn.lock", "yarn"},
}

// nodePackageManagerCmds gives the invocation shape for each package
// manager's build/test/lint scripts; npm is the implicit default and isn't
// listed since markers already sets its commands.
var nodePackageManagerCmds = map[string]Ecosystem{
	"yarn": {BuildCmd: "yarn build", TestCmd: "yarn test", LintCmd: "yarn lint"},
	"pnpm": {BuildCmd: "pnpm build", TestCmd: "pnpm test", LintCmd: "pnpm lint"},
	"bun":  {BuildCmd: "bun run build", TestCmd: "bun test", LintCmd: "bun run lint"},
}

// applyNodePackageManager rewrites the Node.js entry in found (if present)
// to use the commands for the package manager implied by its lockfile,
// instead of always assuming npm.
func applyNodePackageManager(found []Ecosystem, topNames map[string]bool) {
	var manager string
	for _, lf := range nodeLockfiles {
		if topNames[lf.file] {
			manager = lf.manager
			break
		}
	}
	if manager == "" {
		return
	}
	cmds, ok := nodePackageManagerCmds[manager]
	if !ok {
		return
	}
	for i := range found {
		if found[i].Name != "Node.js" {
			continue
		}
		found[i].BuildCmd = cmds.BuildCmd
		found[i].TestCmd = cmds.TestCmd
		found[i].LintCmd = cmds.LintCmd
	}
}

// envExampleFiles are root-level filenames conventionally used to document
// required environment variables without committing real secrets.
var envExampleFiles = []string{".env.example", ".env.sample", ".env.dist"}

// detectEnvExample reports the first env-template file found at the
// repository root, if any, so generated instructions can point an agent at
// it instead of it guessing required configuration from code.
func detectEnvExample(topNames map[string]bool) string {
	for _, f := range envExampleFiles {
		if topNames[f] {
			return f
		}
	}
	return ""
}
