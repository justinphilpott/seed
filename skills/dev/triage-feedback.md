# Command: triage-feedback

Fetch and triage open agent-feedback issues on the seed repo. Use this before introducing notable functionality changes or user-surface changes — it's your checkpoint to ensure in-flight feedback is considered before the surface shifts.

## Steps

### 1. Fetch open feedback issues

```bash
gh issue list \
  --repo justinphilpott/seed \
  --label agent-feedback \
  --state open \
  --json number,title,body,createdAt,url \
  --limit 50
```

If `gh` is unavailable or unauthenticated, report that and stop.

If there are no open issues, report "No open agent-feedback issues" and stop.

### 2. Parse and categorise

For each issue, extract the category from the issue body (template / working-practice / skill / structure / other).

Group issues by category. Within each group, note:
- Issue number and title
- What was observed (the "What I Observed" field)
- What was suggested (the "Suggestion" field)
- Project context if relevant

### 3. Identify patterns

Look across issues for:
- **Repeated themes**: Multiple issues pointing at the same gap (different projects, same problem — high signal)
- **Contradictions**: Two issues suggesting opposite things (needs user judgement)
- **Project-specific noise**: Issues that seem tied to one project type, not general to seed

Flag each pattern explicitly.

### 4. Present triage summary

Output a structured summary for the user to review:

```
## Triage: agent-feedback issues
[N open issues across N categories]

### By Category

#### template (N issues)
- #N: [title] — [one-line summary of what was observed and suggested]
- ...

#### working-practice (N issues)
- ...

#### skill (N issues)
- ...

#### structure (N issues)
- ...

#### other (N issues)
- ...

### Patterns
- [Pattern name]: Issues #N, #N point at the same gap — [brief description]
- [Contradiction]: Issues #N and #N suggest conflicting things — [description]
- [Likely noise]: Issue #N seems specific to [context], probably not general

### Recommended Priority
High — address before this change:
- [issue or pattern] — because [reason]

Defer — valid but not blocking:
- [issue or pattern] — because [reason]

Review — may be noise or contradictory:
- [issue or pattern] — because [reason]
```

### 5. Ask what to do

Present the user with the options:
- **Act now**: For any high-priority issue, offer to help draft the concrete code or template change
- **Close as won't fix**: For noise issues, draft a closing comment for the user to review
- **Defer with a label**: For valid but non-blocking issues, suggest adding a `deferred` label
- **Nothing yet**: If the issues don't apply to the change being considered

Do not make any GitHub changes without explicit user confirmation.

## Scope

- **Do**: Fetch and analyse accurately — don't editorialize beyond what the issues say
- **Do**: Flag patterns and signal strength clearly
- **Do**: Ask before taking any action on issues
- **Don't**: Close, comment on, or label issues without the user saying so
- **Don't**: Let this become a blocker — if issues are low-signal, say so and move on
