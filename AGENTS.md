# Agent Context for Seed

Go CLI tool for rapid agentic POC scaffolding. Run `seed <directory>` to create a new project with minimal, agent-friendly documentation files.

## Quick Links

- [CONTRIBUTING.md](CONTRIBUTING.md) - Development setup, architecture, template variables, extending seed
- [DECISIONS.md](DECISIONS.md) - Architectural decisions and rationale
- [TODO.md](TODO.md) - Active work and next steps
- [LEARNINGS.md](LEARNINGS.md) - Validated discoveries across both layers (see below)

## Quick Reference

```bash
go mod tidy          # Install/update dependencies
go run .             # Run without building
make build           # Build binary with version injection
make test            # Run tests
go fmt ./...         # Format code
go vet ./...         # Static analysis
```

## Working Practices

- **Small, atomic commits**: One logical change per commit. If you can't summarise it in a sentence, break it up
- **Commit early, commit often**: Working code with tests beats perfect code in progress. Small commits are easy to review, revert, and understand in git log
- **TODO.md as live context**: Before starting work, write what you're doing in TODO.md's "Doing Now" section — enough detail to resume if interrupted or context is lost. When the work is complete, derive your commit message from those items, then clear the section
- **Docs travel with code**: If a change affects how the project works, update the relevant docs in the same commit — not later
- **Check coherence before committing**: Skim project docs and verify they still agree with each other and with the code. Fix drift immediately — it compounds fast
- **Capture learnings**: When you discover something non-obvious — a gotcha, a pattern that works, a workaround — add it to LEARNINGS.md. If it's not worth writing down, it wasn't a real learning
- **Prune ruthlessly**: Replace placeholders with real content as soon as you can, or delete them. Stale scaffolding is worse than no scaffolding
- **Entropy guard**: Before committing non-trivial work, run `skills/entropy-guard.md` in full — don't shortcut it. It ensures the project's docs remain coherent and self-referential with what was just built

## Project Constraints

- Single external dependency (Charm's Huh library for TUI)
- Templates embedded at compile time via `//go:embed templates/*.tmpl`
- Devcontainer JSON generated programmatically (encoding/json), not via text/template
- Separation of concerns: wizard collects input, scaffold writes files, main orchestrates
- Version injected at build time via `-ldflags "-X main.Version=$(VERSION)"`

## Key Files

- **main.go** - CLI entry point, argument parsing, orchestration
- **wizard.go** - TUI wizard (Charm Huh), user input collection
- **scaffold.go** - Template rendering (embed.FS + text/template), devcontainer generation, .vscode/extensions.json generation
- **scaffold_test.go** - Scaffold/template tests
- **wizard_test.go** - Wizard validation and data transformation tests
- **skills.go** - Skill file embedding and installation logic
- **templates/*.tmpl** - Embedded project templates (README, AGENTS, DECISIONS, TODO, LEARNINGS, Dockerfile)
- **skills/*.md** - Skills installed into every seeded project (doc-health-check, entropy-guard, seed-feedback, seed-ux-eval)
- **skills/dev/*.md** - Seed development workflow skills; not embedded, not installed into seeded projects
- **.claude/commands/*.md** - Symlinks into skills/dev/ so Claude Code can expose them as slash commands

## Testing

- `make test` or `go test -count=1 ./...`
- Table-driven tests with `t.Run()` subtests
- Temp directory isolation via `tempDir(t)` helper

## Meta: Seed Documents What It Builds

Seed scaffolds agentic docs for other projects (`templates/*.tmpl`) and also maintains its own agentic docs for development. These are two layers of the same philosophy — insights from improving one should inform the other.

- **Seed's templates** — starter docs for new projects (AGENTS.md, README.md, DECISIONS.md, TODO.md, LEARNINGS.md)
- **Seed's own docs** — mature docs for this project (AGENTS.md, CONTRIBUTING.md, LEARNINGS.md)

When recording learnings, note which layer they apply to — or both. See [LEARNINGS.md](LEARNINGS.md).

## Branch

Main branch: `main`. Feature branches are cut from `main` and merged via PR.

## Releasing

Push a git tag to trigger automatic cross-platform builds via GitHub Actions:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Builds for Linux/macOS (amd64/arm64) and Windows (amd64) are published as GitHub Releases.

## Maintaining These Docs

When adding/removing source files, templates, or changing architecture, update:
- The **Key Files** section in this file
- The **Architecture** section in CONTRIBUTING.md

When making architectural decisions, add an entry to DECISIONS.md.
