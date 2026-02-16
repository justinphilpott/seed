# Contributing to Seed

This guide covers development setup, architecture, and how to extend seed.

## Quick Start

### 1. Open in DevContainer

In VS Code:
- **Command Palette** (`Ctrl+Shift+P`)
- Type: **Dev Containers: Reopen in Container**
- Wait for container to build (has Go pre-installed)

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

### Key Design Decisions

**Ultra-minimal templates.** Templates are scaffolding to build on, not documentation homework. Removed: TechStack, Author, "Last Updated" fields, format sections, verbose guidelines. Kept: clean examples, minimal placeholders, navigation links.

**Embedded filesystem.** Templates and skills are embedded at compile time via `//go:embed`. The binary is fully self-contained — no external files needed.

**Programmatic JSON.** Devcontainer JSON is generated via `encoding/json`, not text/template. JSON with conditional fields is fragile in text/template (trailing commas, escaping). Programmatic generation guarantees valid JSON.

**Single dependency.** Only one external dependency: `github.com/charmbracelet/huh` for the TUI wizard. Everything else uses Go's standard library.

## Template Variables

Templates receive a `TemplateData` struct (defined in `scaffold.go`):

**Required** (collected by wizard):
- `ProjectName` — Name of the project
- `Description` — Short description (1-2 sentences)

**Optional**:
- `IncludeDevContainer` — Whether to scaffold .devcontainer/
- `DevContainerImage` — MCR image tag, e.g. `go:2-1.25-trixie`
- `AIChatContinuity` — Whether to enable AI chat continuity

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

1. Create `skills/your-skill.md`
2. It's automatically embedded and installed by `seed skills`

## Testing

- Run: `make test` or `go test -count=1 ./...`
- Table-driven tests with `t.Run()` subtests
- Each test uses `t.TempDir()` for isolated temp directories (auto-cleaned)
- `scaffold_test.go` — file existence, template content, devcontainer JSON validity, error handling, edge cases
- `wizard_test.go` — input validation boundaries, `WizardData` to `TemplateData` conversion

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
- The **Architecture** section in CLAUDE.md
- The **Key Files** section in AGENTS.md
- The **Architecture** section in this file
