# Seed

Scaffold a useful set of documentation files (human and agent-friendly) and other minimally opinionated structures to support a proof-of-concept build.

## Quick start

```bash
curl -fsSL https://raw.githubusercontent.com/justinphilpott/seed/main/install.sh | sh
seed myproject
```

You get:

- **Agent-ready docs** — AGENTS.md, TODO.md, DECISIONS.md, LEARNINGS.md pre-wired for AI-assisted development
- **Optional dev container** — language-specific base image with `gh` CLI (via devcontainer feature), authenticated via host token
- **AI chat continuity** — setup script persists conversation context across container rebuilds
- **Agent skills** — reusable markdown procedures (`doc-health-check`, `entropy-guard`, ...) installed into `skills/`
- **No bloat** — just enough structure to grow into

## What you get

```
myproject/
├── README.md            Human entry point — what this project is and how to run it
├── AGENTS.md            Context and constraints for AI agents working in the repo
├── DECISIONS.md         Lightweight architectural decision log
├── TODO.md              Active work items and next steps
├── LEARNINGS.md         Validated discoveries worth preserving
├── .gitignore           Git ignore rules (language-aware)
├── .editorconfig        Editor formatting defaults
├── LICENSE              Open-source license (optional)
├── skills/              Reusable agent skill files
├── .vscode/             (optional, with devcontainer + extensions)
│   └── extensions.json  Prompts VS Code to install recommended extensions
└── .devcontainer/       (optional)
    ├── Dockerfile       Language-specific base image
    ├── devcontainer.json
    └── setup.sh         AI chat continuity (optional, with chat continuity)
```

Every file is a starting point, not a finished document. Fill them in as you build.

## Install

### Quick install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/justinphilpott/seed/main/install.sh | sh
```

Install to a custom directory:

```bash
curl -fsSL https://raw.githubusercontent.com/justinphilpott/seed/main/install.sh | INSTALL_DIR=/usr/local/bin sh
```

### From source

```bash
go install github.com/justinphilpott/seed@latest
```

### From release binaries

Download the binary for your platform from [GitHub Releases](https://github.com/justinphilpott/seed/releases), then:

```bash
chmod +x seed-linux-amd64
mv seed-linux-amd64 ~/.local/bin/seed
seed --version
```

Available binaries: `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`, `windows-amd64`.

## Usage

```bash
seed myproject              # Scaffold a new project
seed ~/dev/myapp            # Absolute paths work too
seed .                      # Use current directory (prompts if non-empty)
```

### Dev containers

Pick a language stack during the wizard and Seed generates a `.devcontainer/` config using [Microsoft Container Registry](https://mcr.microsoft.com) base images. `gh` CLI is included via a [devcontainer feature](https://github.com/devcontainers/features) and authenticated via your host token — before opening the container, run:

```bash
export GH_TOKEN=$(gh auth token)
```

If you enable AI chat continuity, a setup script auto-detects Claude Code and Codex and wires up conversation persistence so you keep your context across container rebuilds.

### Skills

Skills are markdown files that define reusable procedures your AI agent can follow. They are installed automatically into `skills/` when you scaffold a project.

Currently ships with:
- `doc-health-check` — an audit that reviews your project's documentation coverage and flags gaps
- `entropy-guard` — checks that docs remain coherent and self-consistent before committing
- `seed-ux-eval` — first-5-minutes evaluation of scaffolding quality from a fresh agent's perspective
- `seed-feedback` — an optional channel for agents to submit suggestions back to seed when they notice gaps in the scaffolding

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, architecture, and how to extend seed.

## License

MIT — see [LICENSE](LICENSE) for details.
