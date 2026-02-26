# Contributing to Seed

This guide covers development setup, architecture, and how to extend seed.

## Quick Start

### 1. Open in DevContainer

First, authenticate `gh` on your host so it's available inside the container:

```bash
export GH_TOKEN=$(gh auth token)
```

Then in VS Code:
- **Command Palette** (`Ctrl+Shift+P`)
- Type: **Dev Containers: Reopen in Container**
- Wait for container to build (Go base image, gh CLI added via devcontainer feature)

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Build and Test

```bash
make build    # Build binary with version injection
make test     # Run tests
./seed test-project
```

## Development Commands

| Command | Purpose |
|---------|---------|
| `go run .` | Run without building (faster iteration) |
| `make build` | Build binary with version injection |
| `make test` | Run tests |
| `go fmt ./...` | Format all Go files |
| `go vet ./...` | Static analysis |
| `go mod tidy` | Update dependencies |

## Architecture

Seed follows strict separation of concerns across four files:

- **main.go** — CLI entry point, argument parsing, orchestration. Thin glue layer.
- **wizard.go** — TUI wizard (Charm's Huh library). Collects user input. Knows nothing about templates or file I/O.
- **scaffold.go** — Template rendering (embed.FS + text/template), devcontainer generation (encoding/json). Knows nothing about TUI.
- **skills.go** — Skill file embedding and installation. Same embed pattern as scaffold.go.

Key CLI behavior coverage lives in **main_test.go** (argument parsing and output formatting expectations).

### Key Design Decisions

Full rationale for each decision lives in [DECISIONS.md](DECISIONS.md). Quick orientation for contributors:

- **Ultra-minimal templates** — scaffolding to grow into, not documentation homework
- **Working practices over structure** — AGENTS.md encodes habits, not checklists; habits outlast structure
- **TODO.md as live working context** — "Doing Now" doubles as crash recovery and commit message source
- **Embedded filesystem** — binary is self-contained via `//go:embed`; no external files needed
- **Programmatic JSON** — devcontainer config uses `encoding/json` to guarantee valid JSON across conditional fields
- **Extensions volume via staging path + symlink** — avoids root-owned `.vscode-server/` when mounting Docker volumes inside nested paths
- **Single dependency** — only `github.com/charmbracelet/huh` for the TUI; everything else is standard library

## Template Variables

Templates receive a `TemplateData` struct (defined in `scaffold.go`):

**Required** (collected by wizard):
- `ProjectName` — Name of the project
- `Description` — Short description (1-2 sentences)

**Optional**:
- `IncludeDevContainer` — Whether to scaffold .devcontainer/
- `DevContainerImage` — MCR image tag, e.g. `go:2-1.25-trixie`
- `AIChatContinuity` — Whether to enable AI chat continuity
- `VSCodeExtensions` — VS Code extension IDs; added to `devcontainer.json` customizations (auto-install in container) and to `.vscode/extensions.json` (workspace recommendation prompt)
- `License` — `"none"`, `"MIT"`, or `"Apache-2.0"`
- `Year` — Current year (auto-populated by Scaffolder)

## Extending Seed

### Add a Template Variable

1. Add field to `TemplateData` struct in `scaffold.go`
2. Add field to `WizardData` struct in `wizard.go`
3. Add input field inside `RunWizard()` in `wizard.go`
4. Map the field in `ToTemplateData()` in `wizard.go`
5. Use in templates: `{{.FieldName}}`

### Add a New Template

1. Create `templates/NEWFILE.md.tmpl`
2. Add to `coreTemplates` slice in the `Scaffold()` method in `scaffold.go`

The scaffold logic automatically strips `.tmpl` and renders with `TemplateData`.

### Add a New Skill

There are two categories of skill file — they live in different subdirectories and serve different audiences:

**Skills installed into seeded projects** (`skills/*.md`):
- Embedded in the binary at compile time via `//go:embed skills/*.md`
- Automatically copied to `targetDir/skills/` when seed scaffolds a new project
- Intended for agents working inside a seeded project (e.g., `seed-feedback`, `doc-health-check`, `entropy-guard`)
- To add: create `skills/your-skill.md` — it's automatically embedded and installed

**Seed development workflow skills** (`skills/dev/*.md`):
- Skills for use while developing seed itself; not embedded, not installed into seeded projects
- Exposed as Claude Code slash commands via symlinks in `.claude/commands/` (e.g., `/test-scaffold`, `/triage-feedback`)
- To add: create `skills/dev/your-skill.md`, then `ln -s ../../skills/dev/your-skill.md .claude/commands/your-skill.md`

## Feedback Loop

Seed has a structured feedback loop for gathering UX signal from freshly seeded projects and translating it into improvements. The loop runs in three stages:

**1. Evaluation** — `skills/seed-ux-eval.md` is installed into every seeded project. A fresh agent opening the project runs through a checklist to assess the scaffolding quality (first-5-minutes clarity, placeholder density, working practice fit, doc coherence). This produces an unbiased read on what works and what doesn't.

**2. Submission** — When the evaluating agent finds something concrete, they use `skills/seed-feedback.md` to file a GitHub issue on this repo with the `agent-feedback` label.

**3. Triage** — Before merging any change that affects the user surface (templates, wizard options, skill content), run `/triage-feedback` (Claude Code slash command). It fetches open `agent-feedback` issues, groups them by category, surfaces patterns, and presents a priority recommendation. You approve what to act on — nothing is acted on automatically.

**Your place in the loop**: between stage 3 and any code change. Triage is the gate. Changes that would affect seeded projects should only land after considering open feedback on that surface.

**Validation**: Before merging significant changes, run `/test-scaffold` (a Claude Code slash command backed by `skills/dev/test-scaffold.md`) to verify the test suite, build, and scaffold output across representative option combinations.

## Testing

- Run: `make test` or `go test -count=1 ./...`
- Table-driven tests with `t.Run()` subtests
- Each test uses `tempDir(t)` helper for isolated temp directories (auto-cleaned)
- `scaffold_test.go` — file existence, template content, devcontainer JSON validity, error handling, edge cases
- `wizard_test.go` — input validation boundaries, `WizardData` to `TemplateData` conversion
- `scripts/test-install.sh` — installer integration check (PATH guidance + binary install flow with mocked network)

### Manual Testing

```bash
# Basic flow
./seed mytest && ls mytest/

# Empty directory (should work)
mkdir existing && ./seed existing

# Non-empty directory (should prompt for confirmation)
mkdir nonempty && touch nonempty/file.txt && ./seed nonempty
```

## Roadmap

Feature requests and planned work are tracked as [GitHub Issues](https://github.com/justinphilpott/seed/issues).

## Releasing

Push a git tag to trigger automatic cross-platform builds via GitHub Actions:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Builds for Linux/macOS (amd64/arm64) and Windows (amd64) are published as GitHub Releases.

Version is injected at build time via `-ldflags "-X main.Version=$(VERSION)"`.

## Go Concepts Used

Brief reference for contributors less familiar with Go:

- **`embed.FS`** — Embeds files into the binary at compile time. `//go:embed templates/*.tmpl` must be directly above the variable declaration.
- **`text/template`** — Renders `{{.Field}}` placeholders. Uses `text/template` (not `html/template`) because output is markdown, not HTML.
- **Error wrapping** — `fmt.Errorf("context: %w", err)` preserves the error chain for debugging.
- **Struct methods** — `func (s *Scaffolder) Scaffold(...)` attaches methods to types. Pointer receivers (`*Scaffolder`) can modify the struct.
- **`defer`** — `defer file.Close()` schedules cleanup to run when the function exits, even on error.

## Troubleshooting

**"templates not found"** — Ensure `//go:embed templates/*.tmpl` is directly above `var templatesFS` with no blank lines between them.

**"module not found"** — Run `go mod tidy` to re-download dependencies.

**"cannot find package"** — Run `go clean -modcache && go mod tidy` to clear and rebuild the module cache.

## Maintaining These Docs

When adding/removing source files, templates, or changing architecture, update:
- The **Key Files** section in AGENTS.md
- The **Architecture** section in this file
