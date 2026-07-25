# CLAUDE.md — prism-control

PRISM Control is a Product Delivery Control Plane: the canonical, queryable store (Dolt) for cross-repository initiatives, per-repo Roadmap Items (RMIs), lease-based work assignments, and delivery evidence. It is headless and library-first — no UI here (VisionStudio renders; devfolio/omnidevx compute metrics).

Specs are the source of truth for design decisions — read before implementing:

- `docs/specs/PRD.md` — problem, goals, non-goals, success metrics
- `docs/specs/TRD.md` — schema, package layout, CLI/MCP surfaces, trailer spec, testing strategy
- `docs/specs/PLAN.md` — phased build order with exit criteria and risks
- `docs/specs/ROADMAP.md` — RMI-level breakdown by themed phase

## Conventions

- **Stack:** Ent (`entgo.io/ent`, MySQL dialect → Dolt), Cobra CLI, official Go MCP SDK (`github.com/modelcontextprotocol/go-sdk`). Verify latest versions before adding to `go.mod`.
- **Library-first:** all behavior lives in `pkg/*` and the root `prismcontrol` SDK facade. `cmd/prismctl` and the MCP server are thin adapters over one shared service layer; CLI and MCP write paths must call identical service methods.
- **Unit-first testing:** domain logic is pure or depends only on the `pkg/store` interface (in-memory fake). Only the `doltstore` integration suite may require a running Dolt; guard it with a build tag.
- **Plan is authoritative; commits are evidence.** Ingest never mutates plan structure — mismatches are validation errors to surface.
- **Phases:** themed groupings of RMIs; phase status is always derived from member RMIs, never set directly.
- **JSON types:** Go structs are the source of truth; generate schemas with `invopop/jsonschema`, lint with `schemago`, embed via `//go:embed` under `schema/`.

## Commit Attribution (applies to this repo now)

Every commit implementing a roadmap item carries a git trailer referencing its RMI from `docs/specs/ROADMAP.md`:

```text
feat(store): add unit-of-work with Dolt commit

Refs: RMI-PRISMCONTROL-005
```

Trailer, not subject line. RMI only — never the initiative ID. Conventional Commits format for the message itself. This makes the Phase 5 dogfood migration retroactively attribute this project's own history.

## Connection Modes

prismctl supports two database modes:

- **Embedded (default):** Single-session access. Data stored at `~/.productbuildershq/prism`. No setup needed — just run `prismctl` commands directly. Exclusive filesystem lock means only one process at a time.

- **Server:** Multi-session access. Start a Dolt SQL server in one terminal, connect from any number of sessions via DSN. Required when multiple Claude Code sessions (e.g., in separate tmux panes) need concurrent access.

### Starting Server Mode

```bash
# Terminal 1: start the server (runs in foreground)
prismctl db serve --port 13306

# Terminal 2+: set DSN so prismctl uses the server instead of embedded
export PRISMCTL_DSN="root:@tcp(127.0.0.1:13306)/prismcontrol"
prismctl initiative list   # connects to the server
```

When `PRISMCTL_DSN` is set, all prismctl commands use the server. When unset, they use embedded mode. Both modes use the same data directory.

## Session Protocol

Claude Code sessions are first-class users of PRISM Control. Every session that implements work follows this four-step protocol. The same service layer backs both the MCP tools (`prismctl mcp`) and the CLI (`prismctl`), so the two are interchangeable.

### Dashboard

View a live dashboard of all initiatives, phases, and RMIs:

```bash
# Live server (auto-refreshes from DB every 5s) — default
prismctl dashboard

# Static HTML file (one-shot snapshot)
prismctl dashboard --static
```

### 1. Plan

Before writing code, verify the initiative, phases, and RMIs exist. `work ready` automatically checks for ROADMAP.md drift against the database and warns if discrepancies are found:

```bash
prismctl initiative list
prismctl initiative get INIT-PRISMCONTROL-001
prismctl work ready --initiative INIT-PRISMCONTROL-001
```

If new work needs to be decomposed, create RMIs first:

```bash
prismctl rmi create --id RMI-MYREPO-042 --repo github.com/org/myrepo \
  --initiative INIT-X-001 --phase INIT-X-001/phase-2 \
  --title "Add widget support" --type capability --required
```

### 2. Claim

Claim a single RMI or an entire phase before starting work. The `--worker` flag is auto-detected from `CLAUDE_CODE_SESSION_ID` when running inside Claude Code — no need to pass it manually:

```bash
# Single RMI (worker auto-detected in Claude Code sessions)
prismctl work claim RMI-MYREPO-042 --lease-hours 4

# Entire phase (claims all ready, unblocked, unclaimed RMIs)
prismctl work claim-phase INIT-X-001/phase-2 --lease-hours 4

# Entire phase, auto-transitioning proposed/planned RMIs to ready first
prismctl work claim-phase INIT-X-001/phase-2 --ready --lease-hours 4
```

Renew the lease if work takes longer (accepts RMI ID or assignment ID):

```bash
prismctl work renew RMI-MYREPO-042 --lease-hours 4
```

### 3. Execute

Work in the product repo. Every commit made under a claim carries the trailer:

```text
feat(widget): add alpha-channel overlay support

Refs: RMI-MYREPO-042
```

Trailer rules:

- Git trailer with key `Refs:` — value is one RMI ID (comma-separated if a commit genuinely serves several)
- RMI only — never the initiative ID; the RMI→initiative edge is resolved at report time
- Subject line stays clean; the trailer goes in the footer
- Squash-merge fallbacks: RMI in the PR body; branch naming `rmi/<repo>-<nnn>-<slug>`
- Attribution precedence: trailer → PR reference → branch name

### 4. Update

After completing work, update the assignment with evidence and handoff. All work commands accept RMI IDs or assignment IDs:

```bash
# Add evidence (commits, PRs, releases)
prismctl work update RMI-MYREPO-042 \
  --handoff '{"completed":["widget overlay"],"remaining":[],"decisions":["used alpha compositing"],"next_action":"none"}'

# Mark complete and auto-transition RMI to completed (accepts multiple IDs)
prismctl work complete RMI-MYREPO-042 --transition
prismctl work complete RMI-MYREPO-042 RMI-MYREPO-043 RMI-MYREPO-044 --transition

# Complete an entire phase at once
prismctl work complete-phase INIT-X-001/phase-2 --transition

# Or release if handing off to another session
prismctl work release RMI-MYREPO-042 \
  --handoff '{"completed":["core logic"],"remaining":["tests"],"next_action":"add unit tests"}'
```

#### Bulk Status Transitions

Transition all RMIs in a phase to a new status without claiming:

```bash
# All RMIs in phase → ready
prismctl rmi update-phase INIT-X-001/phase-3 --status ready

# Only proposed RMIs → ready
prismctl rmi update-phase INIT-X-001/phase-3 --status ready --from proposed
```

### Product Repo Pointer

Product repos that participate in PRISM-tracked initiatives add one line to their `CLAUDE.md`:

```markdown
## PRISM Control

This repo's roadmap items are tracked in [prism-control](https://github.com/ProductBuildersHQ/prism-control). Use `prismctl work ready --repo github.com/ProductBuildersHQ/<this-repo>` to find claimable work, and carry the `Refs: RMI-<REPOSLUG>-<NNN>` trailer on every commit.
