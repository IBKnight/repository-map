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

	a := Analyze(root, "myproj")
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

	a := Analyze(root, "demo")
	if len(a.Ecosystems) != 1 || a.Ecosystems[0].Name != "Dart/Flutter" {
		t.Fatalf("expected Dart/Flutter ecosystem detected, got %+v", a.Ecosystems)
	}
	if a.Ecosystems[0].TestCmd != "flutter test" {
		t.Errorf("expected flutter test command, got %+v", a.Ecosystems[0])
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
	a := Analyze(root, "myproj")

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
	a := Analyze(root, "demo")

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
