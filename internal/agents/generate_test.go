package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/IBKnight/repository-map/internal/scanner"
)

func TestAnalyzeDetectsGoEcosystem(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "myproj")
	if len(a.Ecosystems) != 1 || a.Ecosystems[0].Name != "Go" {
		t.Fatalf("expected Go ecosystem detected, got %+v", a.Ecosystems)
	}
}

func TestAnalyzeDetectsFlutterEcosystem(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "demo")
	if len(a.Ecosystems) != 1 || a.Ecosystems[0].Name != "Dart/Flutter" {
		t.Fatalf("expected Dart/Flutter ecosystem detected, got %+v", a.Ecosystems)
	}
	if a.Ecosystems[0].TestCmd != "flutter test" {
		t.Errorf("expected flutter test command, got %+v", a.Ecosystems[0])
	}
}

func TestAnalyzeUsesPnpmCommandsWhenLockfilePresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "demo")
	if len(a.Ecosystems) != 1 || a.Ecosystems[0].Name != "Node.js" {
		t.Fatalf("expected Node.js ecosystem detected, got %+v", a.Ecosystems)
	}
	eco := a.Ecosystems[0]
	if eco.BuildCmd != "pnpm build" || eco.TestCmd != "pnpm test" || eco.LintCmd != "pnpm lint" {
		t.Errorf("expected pnpm commands, got %+v", eco)
	}
}

func TestAnalyzeDetectsEnvExample(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("API_KEY=\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "demo")
	if a.EnvExample != ".env.example" {
		t.Errorf("expected .env.example detected, got %q", a.EnvExample)
	}
}

func TestAnalyzeUsesFvmCommandsWhenFvmDirPresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".fvm"), 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "demo")
	if len(a.Ecosystems) != 1 || a.Ecosystems[0].Name != "Dart/Flutter" {
		t.Fatalf("expected Dart/Flutter ecosystem detected, got %+v", a.Ecosystems)
	}
	eco := a.Ecosystems[0]
	if eco.BuildCmd != "fvm flutter build" || eco.TestCmd != "fvm flutter test" || eco.LintCmd != "fvm flutter analyze" {
		t.Errorf("expected fvm-prefixed commands, got %+v", eco)
	}
}

func TestAnalyzeUsesFvmCommandsWhenFvmrcPresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".fvmrc"), []byte(`{"flutter":"3.19.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "demo")
	if len(a.Ecosystems) != 1 || a.Ecosystems[0].BuildCmd != "fvm flutter build" {
		t.Fatalf("expected fvm-prefixed commands, got %+v", a.Ecosystems)
	}
}

func TestAnalyzeWithoutFvmUsesPlainFlutterCommands(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "demo")
	if len(a.Ecosystems) != 1 || a.Ecosystems[0].BuildCmd != "flutter build" {
		t.Fatalf("expected plain flutter commands, got %+v", a.Ecosystems)
	}
}

func TestAnalyzeDetectsStyleConfigs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".editorconfig"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".golangci.yml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "demo")
	want := []string{".editorconfig", ".golangci.yml"}
	if len(a.StyleConfigs) != len(want) {
		t.Fatalf("expected %v, got %v", want, a.StyleConfigs)
	}
	for i, w := range want {
		if a.StyleConfigs[i] != w {
			t.Errorf("expected %v, got %v", want, a.StyleConfigs)
			break
		}
	}
}

func TestAnalyzeDetectsSubProjects(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "apps", "web"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "services", "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "apps", "web", "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "services", "api", "go.mod"), []byte("module api\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := scanner.Scan(dir, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}

	a := Analyze(dir, root, "demo")
	if len(a.Ecosystems) != 0 {
		t.Errorf("expected no root-level ecosystems, got %+v", a.Ecosystems)
	}
	want := []string{"apps/web", "services/api"}
	if len(a.SubProjects) != len(want) {
		t.Fatalf("expected %v, got %v", want, a.SubProjects)
	}
	for i, w := range want {
		if a.SubProjects[i] != w {
			t.Errorf("expected %v, got %v", want, a.SubProjects)
			break
		}
	}
}

func TestGenerateWritesAndSkipsExisting(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root, err := scanner.Scan(src, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}
	a := Analyze(src, root, "myproj")

	out := t.TempDir()
	targets, err := ParseTargets([]string{"generic"})
	if err != nil {
		t.Fatal(err)
	}

	res, err := Generate(a, targets, out, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Written) != 1 || res.Written[0] != "AGENTS.md" {
		t.Fatalf("expected AGENTS.md written, got %+v", res)
	}
	if _, err := os.Stat(filepath.Join(out, "AGENTS.md")); err != nil {
		t.Fatalf("expected AGENTS.md to exist: %v", err)
	}

	// Second run without force should skip.
	res2, err := Generate(a, targets, out, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res2.Written) != 0 || len(res2.Skipped) != 1 {
		t.Fatalf("expected file to be skipped on second run, got %+v", res2)
	}

	// With force, it should be rewritten.
	res3, err := Generate(a, targets, out, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res3.Written) != 1 {
		t.Fatalf("expected file to be rewritten with force, got %+v", res3)
	}
}

func TestParseTargetsRejectsUnknown(t *testing.T) {
	if _, err := ParseTargets([]string{"bogus"}); err == nil {
		t.Fatal("expected error for unknown target")
	}
}

func TestClaudeTargetGeneratesValidSettingsPermissions(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	root, err := scanner.Scan(src, scanner.Options{})
	if err != nil {
		t.Fatal(err)
	}
	a := Analyze(src, root, "demo")

	out := t.TempDir()
	if _, err := Generate(a, []Target{TargetClaude}, out, false); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(filepath.Join(out, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("expected .claude/settings.json to be written: %v", err)
	}

	var parsed struct {
		Permissions struct {
			Allow []string `json:"allow"`
			Deny  []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}

	wantAllow := map[string]bool{"Bash(go build*)": true, "Bash(go test*)": true, "Bash(npm run*)": true}
	got := map[string]bool{}
	for _, rule := range parsed.Permissions.Allow {
		got[rule] = true
	}
	for rule := range wantAllow {
		if !got[rule] {
			t.Errorf("expected allow rule %q, got %+v", rule, parsed.Permissions.Allow)
		}
	}

	foundDeny := false
	for _, rule := range parsed.Permissions.Deny {
		if rule == "Bash(git reset --hard*)" {
			foundDeny = true
		}
	}
	if !foundDeny {
		t.Errorf("expected destructive git commands to be denied, got %+v", parsed.Permissions.Deny)
	}
}
