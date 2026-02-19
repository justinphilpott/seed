# CLAUDE.md

## Project Overview

Seed is a Go CLI tool for rapid agentic POC scaffolding. Run `seed <directory>` to create a new project with minimal, agent-friendly documentation files.

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, architecture, template variables, and how to extend seed.

## Quick Reference

```bash
go mod tidy          # Install/update dependencies
go run .             # Run without building
make build           # Build binary with version injection
make test            # Run tests
go fmt ./...         # Format code
go vet ./...         # Static analysis
```

## Architecture

- **main.go** — CLI entry point, argument parsing, orchestration
- **wizard.go** — TUI wizard (Charm Huh library), user input collection
- **scaffold.go** — Template rendering (embed.FS + text/template), devcontainer generation (encoding/json), .vscode/extensions.json generation
- **scaffold_test.go** — Scaffold/template tests
- **wizard_test.go** — Wizard validation and data transformation tests
- **skills.go** — Skill file embedding and installation logic
- **templates/*.tmpl** — Embedded project templates (README, AGENTS, DECISIONS, TODO, LEARNINGS, Dockerfile)
- **skills/*.md** — Embedded agent skill definitions (doc-health-check, seed-feedback)

## Key Patterns

- Templates are embedded at compile time via `//go:embed templates/*.tmpl`
- Devcontainer JSON is generated programmatically (encoding/json), not via text/template
- Separation of concerns: wizard collects input, scaffold writes files, main orchestrates
- Version injected at build time via `-ldflags "-X main.Version=$(VERSION)"`

## Testing

- Run: `make test` or `go test -count=1 ./...`
- Table-driven tests with `t.Run()` subtests
- Each test uses `tempDir(t)` helper for isolated temp directories (auto-cleaned)
- scaffold_test.go: file existence, template content, devcontainer JSON validity, error handling, edge cases
- wizard_test.go: input validation boundaries, `WizardData` → `TemplateData` conversion

## Branch

Main branch: `main`. Feature branches are cut from `main` and merged via PR.

## Releasing

Push a git tag (e.g., `git tag v0.1.0 && git push origin v0.1.0`) to trigger GitHub Actions release builds.

## Maintaining These Docs

When adding/removing source files, templates, or changing architecture, update:
- The **Architecture** section in this file
- The **Key Files** section in AGENTS.md
- The **Architecture** section in CONTRIBUTING.md

When making architectural decisions, add an entry to DECISIONS.md.
