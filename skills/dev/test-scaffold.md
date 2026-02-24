# Command: test-scaffold

Run before merging changes to wizard flow, templates, devcontainer generation, or skills. Verifies the scaffold produces correct output across representative option combinations.

## Steps

### 1. Run the test suite

```bash
make test
```

Report pass/fail. If any tests fail, stop here — do not proceed.

Then run with verbose output to see which combinations are exercised:

```bash
go test -count=1 -v ./...
```

### 2. Verify the binary builds clean

```bash
make build
```

Report any build errors or warnings.

### 3. Audit test coverage against wizard options

Read `wizard.go` to understand the full option space the wizard presents:
- License: none / MIT / Apache-2.0
- Git init: yes / no
- Dev container: yes / no
- If dev container: image (Go / Node / Python / Rust / Java / .NET / C++ / Universal)
- AI chat continuity: yes / no
- Agent extensions: Claude Code / Codex / both / neither

Read `scaffold_test.go` to identify which combinations are currently tested.

Report the matrix: which combinations are tested, which are not. Flag any axis that has zero test coverage (i.e., an option that never appears in any test case).

### 4. Inspect a representative generated output

Pick the combination with the most options enabled (devcontainer + chat continuity + MIT license + both extensions). If a test case for this already exists, read the test to find the temp dir it generates and inspect the output. If not, note it as a gap.

For any combination that IS tested, verify:
- All expected files are present (README.md, AGENTS.md, DECISIONS.md, TODO.md, LEARNINGS.md, .gitignore, .editorconfig)
- LICENSE file is present and correct when a license is selected
- `.devcontainer/devcontainer.json` is valid JSON when devcontainer is enabled
- `.devcontainer/setup.sh` is present when chat continuity is enabled
- `.vscode/extensions.json` is present and correct when extensions are selected
- `skills/` directory is present with all four project skills: doc-health-check.md, entropy-guard.md, seed-feedback.md, seed-ux-eval.md

### 5. Check template content for placeholder drift

Read each template in `templates/`:
- Are any hardcoded values that should be template variables (e.g. a literal project name)?
- Are any template variables referenced that don't exist in `TemplateData`?
- Are any `TemplateData` fields unused in all templates?

### 6. Report

Output a structured summary:

```
## test-scaffold Report

### Test Suite
[Pass/Fail — N tests, N passed, N failed]

### Build
[Pass/Fail]

### Coverage Gaps
| Option           | Values Tested        | Values Not Tested |
|------------------|----------------------|-------------------|
| License          | ...                  | ...               |
| Git init         | ...                  | ...               |
| Dev container    | ...                  | ...               |
| Image            | ...                  | ...               |
| Chat continuity  | ...                  | ...               |
| Extensions       | ...                  | ...               |

### Template Issues
[Any placeholder drift, unused variables, or missing template variables]

### Skills Check
[Whether all four project skills are present in skills/: doc-health-check.md, entropy-guard.md, seed-feedback.md, seed-ux-eval.md. Dev skills (test-scaffold.md, triage-feedback.md) must NOT appear there.]

### Recommended Actions
[Prioritised list: test gaps to fill, template issues to fix, anything that should block the change]
```

## Scope

- **Do**: Run the full test suite and report accurately
- **Do**: Flag coverage gaps — missing test coverage is a real risk
- **Do**: Check skills/ has all four project skills and none of the dev skills
- **Don't**: Add or modify test cases — flag gaps, don't fix them here
- **Don't**: Make code changes — this is diagnostic only
