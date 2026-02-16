# Seed

A CLI tool for rapid agentic POC scaffolding. Run `seed <directory>` to create a new project with minimal, agent-friendly documentation files.

## What It Does

Seed runs an interactive wizard and generates a project skeleton designed for AI-assisted development:

```
myproject/
├── README.md          Human entry point
├── AGENTS.md          Agent context and constraints
├── DECISIONS.md       Architectural decisions log
├── TODO.md            Active work tracking
├── LEARNINGS.md       Validated discoveries
└── .devcontainer/     (optional)
    ├── devcontainer.json
    └── setup.sh       AI chat continuity
```

Templates are ultra-minimal — scaffolding to grow into, not documentation homework.

## Install

### Quick install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/justinphilpott/seed/main/install.sh | sh
```

To install to a custom directory:

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
# Make it executable (Linux/macOS)
chmod +x seed-linux-amd64

# Move it onto your PATH
mv seed-linux-amd64 ~/.local/bin/seed

# Verify
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

The wizard collects a project name, description, and optional settings (git init, dev container, AI chat continuity).

### Dev Container Support

Seed can generate a `.devcontainer/` configuration with:
- Base images for Go, Node/TypeScript, Python, Rust, Java, .NET, C++, or Universal
- Optional AI chat continuity that auto-detects Claude Code and Codex, preserving conversations across host and container

### Skills

Agent skills are markdown files that define reusable procedures. Install them into an existing project:

```bash
seed skills ./myproject
```

Currently includes `doc-health-check` — an audit procedure for project documentation coverage.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, architecture, and how to extend seed.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
