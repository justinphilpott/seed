# Learnings

Validated discoveries from building seed. Focus on what we proved, not opinions.

---

### Information Coverage > File Presence

**Topic**: Documentation Architecture

**Insight**: Agent-friendly docs need to cover the right *categories* of information — temporal context (what changed and why), constraints and decisions, code-spatial mapping (what's where and how it connects), active work, and learnings. The specific files are containers; the coverage is what matters. A project's docs will naturally diverge from any starter set as files merge, split, or migrate to serve real needs. That's healthy — the test is "can an agent find the decisions?" not "do you have a DECISIONS.md?"

**Validated by**: Seed's own doc set (CLAUDE.md, AGENTS.md, CONTRIBUTING.md, LEARNINGS.md) differs from what it scaffolds for users (AGENTS.md, README.md, DECISIONS.md, TODO.md, LEARNINGS.md). Every information category is covered in both, just in different containers. Seed's CONTRIBUTING.md absorbed what separate build and context docs would cover.

**Implication**: Templates should scaffold *structure to grow into*, not a prescriptive file set. Doc health tools should audit informational coverage, not file presence.

---

### Maintenance Instructions Prevent Doc Drift

**Topic**: Documentation Architecture

**Insight**: Without explicit "when X changes, update Y" instructions in the docs themselves, agents silently let docs go stale. Baking a maintenance section into project docs from creation is more effective than retrofitting it later.

**Validated by**: Seed's own docs had drifted — missing files in architecture lists, stale references, removed fields still documented. Adding maintenance checklists to CLAUDE.md, CONTRIBUTING.md, and AGENTS.md fixed this. The AGENTS.md template now includes the pattern so scaffolded projects inherit it from day one. Reinforced a second time when a doc sync found DECISIONS.md missing seven decisions, TODO.md not existing at all, and branch references stale — on the project that teaches these practices.

---

### Brittle References Accelerate Drift

**Topic**: Documentation Architecture

**Insight**: Line-number references in docs (e.g., `scaffold.go:28-35`) go stale as soon as code changes. Function and type name references (`the TemplateData struct in scaffold.go`) are stable anchors that survive refactoring. Prefer semantic references over positional ones.

---

### Docker Volume Mounts Inside Nested Paths Create Root Ownership

**Topic**: DevContainer Setup

**Insight**: Mounting a Docker volume at a nested path like `.vscode-server/extensions` causes Docker to create the parent `.vscode-server/` as root. This blocks VS Code from writing sibling files (`extensions.json`, `bin/`, `data/`). The fix is a two-step pattern: mount the volume to a staging path outside the sensitive parent (e.g., `.vscode-extensions-cache`), then symlink it into place at container startup. The Dockerfile must pre-create the staging dir with correct ownership so the volume mount inherits it.

**Validated by**: Seed's own devcontainer hit this exact issue — VS Code couldn't read `extensions.json`. Fixed in commit `35fce36` for seed's own container, then ported to the scaffold code that generates devcontainers for new projects.

**Implication**: When mounting volumes inside paths owned by non-root users, never mount into a subdirectory of a directory the user needs to write to. Use a staging path + symlink instead.

---

### New Docker Volumes Don't Seed extensions.json

**Topic**: DevContainer Setup

**Insight**: Even after fixing the root ownership issue with the staging path + symlink pattern, VS Code throws `Unable to resolve nonexistent file '.vscode-server/extensions/extensions.json'` on first container start. A brand-new named volume is empty, so after the symlink is created pointing `.vscode-server/extensions` → `.vscode-extensions-cache`, the `extensions.json` file VS Code expects simply doesn't exist in the volume yet. The fix: after creating the symlink in `postCreateCommand`/`setup.sh`, conditionally initialize the file: `[ -f ...extensions.json ] || echo '[]' > ...extensions.json`. On subsequent starts the file is already there and gets skipped.

**Validated by**: Persisted error after the staging/symlink fix was already in place. Root-caused by observing the error only on fresh volumes, not on rebuilds of existing ones.

**Implication**: When symlinking into a named volume, never assume the volume has the files VS Code or other tools expect. Initialize required files defensively in `setup.sh`, not the Dockerfile — Dockerfile content is not copied into a volume on first mount when the mount target path was created via `mkdir` rather than a file copy.

---

### gh CLI Auth in Devcontainers: Don't Mount ~/.config/gh

**Topic**: DevContainer Setup

**Insight**: Mounting `~/.config/gh` from the host into a devcontainer causes gh to fail fatally with `"failed to migrate config: couldn't find oauth token: exec: dbus-launch: not found"` — even when `GH_TOKEN` is set. The root cause: modern gh stores oauth tokens in the host system keyring (dbus/gnome-keyring), not inline in `hosts.yml`. When gh starts in the container it reads the mounted config, attempts the multi-account migration, tries to read the token from the keyring via `dbus-launch`, finds it absent, and refuses to continue — before ever checking `GH_TOKEN`. The fix is to not mount `~/.config/gh` at all. `GH_TOKEN` forwarding alone is sufficient; gh uses it directly without needing any config file.

**Validated by**: Confirmed by setting `GH_CONFIG_DIR=/tmp/gh-test` (clean dir, no host config) — `gh auth status` works immediately with just `GH_TOKEN`. Inspecting the mounted `hosts.yml` showed `oauth_token` absent from the file (stored in keyring instead). Fixed in seed's own devcontainer and scaffold output by removing the `~/.config/gh` bind mount.

**Implication**: Never mount `~/.config/gh` from a host that uses the system keyring for token storage. Forward `GH_TOKEN` (and `GITHUB_TOKEN` for Codespaces/CI) as env vars only. The container will create its own clean config on first use.

---

### Dogfooding a Doc-Generation Tool Has an Inherent Lag

**Topic**: Documentation Architecture

**Insight**: A tool that generates doc conventions for other projects cannot perfectly follow those conventions itself in real time. The tool's own docs represent a matured, evolved state; its templates represent the current best-guess starting point for new projects. They diverge naturally as the conventions are refined through use, and converge again through periodic sync — not continuous parity. The gap is a signal that the conventions are evolving healthily, not a maintenance failure.

**Validated by**: Seed's own docs repeatedly drifted from what seed scaffolds for users. The conventions in the templates were being improved faster than seed's own files were being updated. Trying to maintain perfect real-time parity is impractical when the output is itself being designed.

**Implication**: For tools in this class, schedule periodic coherence checks (e.g., run `doc-health-check` on seed itself) rather than expecting continuous sync. Accept that the project's own docs will lag the templates — what matters is that both layers cover the same informational categories, not that they use identical structure.

---

### AGENTS.md for Cross-Agent Compatibility

**Topic**: Project Setup

**Insight**: AGENTS.md is the most universal cross-agent context file — it's read by Claude Code, Codex, Copilot, and Cursor. Tool-specific files (.cursorrules, CODEX.md, etc.) add maintenance burden without proportional value for small projects. One well-maintained AGENTS.md plus a tool-specific file (e.g., CLAUDE.md) covers the landscape.
