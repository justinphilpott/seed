# Seed

Scaffold a POC in seconds, ready for AI-assisted development from commit one.

```bash
curl -fsSL https://raw.githubusercontent.com/justinphilpott/seed/main/install.sh | sh
seed myproject
```

You get minimal, agent-friendly project docs (AGENTS.md, TODO.md, DECISIONS.md, LEARNINGS.md), optional dev container config, and AI chat continuity across container rebuilds. No bloat — just enough structure to grow into.

## What you get

```
myproject/
├── README.md          Human entry point — what this project is and how to run it
├── AGENTS.md          Context and constraints for AI agents working in the repo
├── DECISIONS.md       Lightweight architectural decision log
├── TODO.md            Active work items and next steps
├── LEARNINGS.md       Validated discoveries worth preserving
├── .vscode/           (optional, with devcontainer)
│   └── extensions.json  Prompts VS Code to install recommended extensions
└── .devcontainer/     (optional)
    ├── devcontainer.json
    └── setup.sh       AI chat continuity across rebuilds
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
seed .                      # Use current directory (if empty)
seed skills ./myproject     # Install agent skills into existing project
```

### Dev containers

Pick a language stack during the wizard and Seed generates a `.devcontainer/` config using [Microsoft Container Registry](https://mcr.microsoft.com) base images. `gh` CLI is pre-installed and authenticated via your host token — before opening the container, run:

```bash
export GH_TOKEN=$(gh auth token)
```

If you enable AI chat continuity, a setup script auto-detects Claude Code and Codex and wires up conversation persistence so you keep your context across container rebuilds.

### Skills

Skills are markdown files that define reusable procedures your AI agent can follow. Install them into any existing project:

```bash
seed skills ./myproject
```

Currently ships with:
- `doc-health-check` — an audit that reviews your project's documentation coverage and flags gaps
- `seed-feedback` — an optional channel for agents to submit suggestions back to seed when they notice gaps in the scaffolding

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, architecture, and how to extend seed.

## License

MIT — see [LICENSE](LICENSE) for details.
