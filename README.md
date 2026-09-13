# repository-map

A CLI tool with two jobs:

1. **`repomap map`** — generate a Markdown or plain-text map of an existing repository: directory tree, language breakdown, file/directory counts.
2. **`repomap agents`** — scaffold instruction/skill files for AI coding agents and assistants based on the repository's detected languages, layout, and build tooling — aimed at cutting the tokens a fresh agent session burns re-deriving that context, and the turns lost to permission prompts for routine commands:
   - `CLAUDE.md` + `.claude/skills/<project>-overview/SKILL.md` + `.claude/settings.json` for Claude Code — the settings file preloads a permissions allowlist (build/test/lint commands for every detected ecosystem, plus `git status`/`diff`/`log`/`add`, `find`, `ls`) so those don't prompt, while destructive git operations (`push --force`, `reset --hard`, `clean -f`, `rm -rf`, …) stay explicitly denied
   - `.github/copilot-instructions.md` for GitHub Copilot
   - `.cursor/rules/repository.mdc` for Cursor
   - `AGENTS.md`, a generic convention read by several agentic coding tools (e.g. OpenAI Codex CLI) and usable as a baseline for ChatGPT-based workflows

## Install

```sh
go install github.com/IBKnight/repository-map@latest
```

Or download a prebuilt binary from the [Releases](https://github.com/IBKnight/repository-map/releases) page.

## Usage

### Generate a repository map

```sh
repomap map                        # Markdown to stdout, current directory
repomap map ./some/project -o REPO_MAP.md
repomap map --format text --max-depth 2
repomap map --sizes                # include file sizes in the tree
```

### Scaffold AI agent instruction files

```sh
repomap agents                     # generate all targets in the current directory
repomap agents ./some/project --target claude,copilot
repomap agents --force             # overwrite existing files
repomap agents --name my-project   # override the detected project name
```

Existing files are left untouched unless `--force` is passed, so it's safe to re-run after the repository changes.

## Development

```sh
make ci      # fmt-check + vet + build + test, same as CI
make fmt     # gofmt -w .
make test    # go test ./...
```

CI runs formatting checks, `go vet`, build, and tests on every push/PR to `main` (see `.github/workflows/ci.yml`). Pushing a `vX.Y.Z` tag triggers `.github/workflows/release.yml`, which uses [GoReleaser](https://goreleaser.com) to build and publish cross-platform binaries to GitHub Releases.

## License

[MIT](LICENSE)
