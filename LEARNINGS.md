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

**Validated by**: Seed's own docs had drifted — missing files in architecture lists, stale references, removed fields still documented. Adding maintenance checklists to CLAUDE.md, CONTRIBUTING.md, and AGENTS.md fixed this. The AGENTS.md template now includes the pattern so scaffolded projects inherit it from day one.

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

### AGENTS.md for Cross-Agent Compatibility

**Topic**: Project Setup

**Insight**: AGENTS.md is the most universal cross-agent context file — it's read by Claude Code, Codex, Copilot, and Cursor. Tool-specific files (.cursorrules, CODEX.md, etc.) add maintenance burden without proportional value for small projects. One well-maintained AGENTS.md plus a tool-specific file (e.g., CLAUDE.md) covers the landscape.
