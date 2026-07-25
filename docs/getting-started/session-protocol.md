# Session Protocol

Claude Code sessions are first-class users of PRISM Control. Every session that implements work follows this four-step protocol. The same service layer backs both the CLI (`prismctl`) and MCP tools (`prismctl mcp`), so the two are interchangeable.

## 1. Plan

Before writing code, verify the initiative, phases, and RMIs exist:

```bash
prismctl initiative list
prismctl initiative get INIT-X-001
prismctl work ready --initiative INIT-X-001
```

The `--initiative` form automatically checks for ROADMAP.md drift against the database and warns if discrepancies are found.

To find work by repository:

```bash
prismctl work ready --repo github.com/org/myrepo
```

If new work needs to be decomposed, create RMIs first:

```bash
prismctl rmi create --id RMI-MYREPO-042 --repo github.com/org/myrepo \
  --initiative INIT-X-001 --phase INIT-X-001/phase-2 \
  --title "Add widget support" --type capability --required
```

## 2. Claim

Claim a single RMI or an entire phase before starting work. The `--worker` flag is auto-detected from `CLAUDE_CODE_SESSION_ID` when running inside Claude Code — no need to pass it manually:

```bash
# Single RMI
prismctl work claim RMI-MYREPO-042 --lease-hours 4

# Entire phase (claims all ready, unblocked, unclaimed RMIs)
prismctl work claim-phase INIT-X-001/phase-2 --lease-hours 4

# Auto-transition proposed/planned RMIs to ready first
prismctl work claim-phase INIT-X-001/phase-2 --ready --lease-hours 4
```

Renew the lease if work takes longer:

```bash
prismctl work renew RMI-MYREPO-042 --lease-hours 4
```

## 3. Execute

Work in the product repo. Every commit carries the `Refs:` trailer:

```
feat(widget): add alpha-channel overlay support

Refs: RMI-MYREPO-042
```

### Trailer Rules

- Git trailer with key `Refs:` — value is one RMI ID (comma-separated if a commit genuinely serves several)
- RMI only — never the initiative ID; the RMI→initiative edge is resolved at report time
- Subject line stays clean; the trailer goes in the footer
- Squash-merge fallbacks: RMI in the PR body; branch naming `rmi/<repo>-<nnn>-<slug>`
- Attribution precedence: trailer → PR reference → branch name

## 4. Update

After completing work, update the assignment with evidence and handoff:

```bash
# Add handoff context
prismctl work update RMI-MYREPO-042 \
  --handoff '{"completed":["widget overlay"],"remaining":[],"decisions":["used alpha compositing"],"next_action":"none"}'

# Mark complete and auto-transition RMI to completed
prismctl work complete RMI-MYREPO-042 --transition

# Complete multiple RMIs at once
prismctl work complete RMI-MYREPO-042 RMI-MYREPO-043 --transition

# Complete an entire phase
prismctl work complete-phase INIT-X-001/phase-2 --transition

# Or release if handing off to another session
prismctl work release RMI-MYREPO-042 \
  --handoff '{"completed":["core logic"],"remaining":["tests"],"next_action":"add unit tests"}'
```

## Bulk Status Transitions

Transition all RMIs in a phase to a new status without claiming:

```bash
# All RMIs in phase → ready
prismctl rmi update-phase INIT-X-001/phase-3 --status ready

# Only proposed RMIs → ready
prismctl rmi update-phase INIT-X-001/phase-3 --status ready --from proposed
```
