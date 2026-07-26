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
- **Unit-first testing:** domain logic is pure or depends only on the `pkg/store` interface (in-memory fake). Only the `doltstore` integration suite may require a running Dolt; guard it with a build tag. CI excludes `pkg/store/doltstore` and `cmd/prismctl` from test/lint/SAST because they need ICU C headers (`dolthub/go-icu-regex`) not available on standard runners — all domain logic is covered via MemStore.
- **Plan is authoritative; commits are evidence.** Ingest never mutates plan structure — mismatches are validation errors to surface.
- **Phases:** themed groupings of RMIs; phase status is always derived from member RMIs, never set directly.
- **JSON types:** Go structs are the source of truth; generate schemas with `invopop/jsonschema`, lint with `schemago`, embed via `//go:embed` under `schema/`.

## Building Locally

The embedded Dolt driver requires ICU C headers. On macOS with Homebrew:

```bash
export CGO_CPPFLAGS="-I/opt/homebrew/opt/icu4c@78/include"
export CGO_LDFLAGS="-L/opt/homebrew/opt/icu4c@78/lib"
go build -o ~/go/bin/prismctl ./cmd/prismctl/
go test -v ./...
```

Without these flags, `pkg/store/doltstore` and `cmd/prismctl` will fail to compile. The `pkg/` unit tests (which use MemStore) work without ICU.

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
# Start the server in the background (auto-saves DSN to config)
prismctl db start --port 13306

# Check server status
prismctl db status

# Terminal 2+: prismctl reads the DSN from config automatically
prismctl initiative list   # connects to the server

# Restart or stop
prismctl db restart
prismctl db stop
```

For foreground operation, use `prismctl db serve --port 13306` instead.

DSN resolution order: `--dsn` flag > `$PRISMCTL_DSN` env > config file > embedded default. Use `prismctl config show` to see the resolved value, or `prismctl config set dsn <value>` to set it manually.

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

## Development Guide

This section covers patterns for extending prism-control itself — adding entities, CLI commands, MCP tools, and dashboard views.

### Architecture Layers

Every feature touches these layers in order. When adding a new entity (like Program was added), work top-down:

| Layer | Files | Purpose |
|-------|-------|---------|
| 1. Ent schema | `ent/schema/<entity>.go` | Define fields, edges, storage keys |
| 2. Ent codegen | `ent/` (generated) | Run `go generate ./ent` |
| 3. Store interface | `pkg/store/store.go` | Domain struct + sub-interface (CRUD methods) |
| 4. MemStore | `pkg/store/memstore.go` | In-memory fake (map-based, mutex-guarded) |
| 5. DoltStore | `pkg/store/doltstore/doltstore.go` | Ent-backed production impl |
| 6. Service | `pkg/service/<entity>.go` | Business logic, timestamp management |
| 7. CLI | `cmd/prismctl/<entity>.go` | Cobra commands (thin adapter over service) |
| 8. Dashboard | `cmd/prismctl/dashboard.go` + `dashboard_templates.go` | Data loading + Go HTML templates |
| 9. MCP | `pkg/mcpserver/server.go` | Tool definition + handler |
| 10. Tests | `pkg/store/memstore_test.go`, `pkg/service/<entity>_test.go` | MemStore CRUD + service-layer tests |

### Ent Schema Conventions

Schemas live in `ent/schema/`. Key patterns:

- **ID field:** `field.String("id").StorageKey("<entity>_id").MaxLen(64)` — the StorageKey sets the DB column name
- **Edges:** `edge.To` (has-many), `edge.From` (belongs-to with `.Ref().Unique()`)
- **Timestamps:** `field.Time("created_at")`, `field.Time("updated_at")` — managed by the service layer, not Ent defaults

After any schema change, regenerate:

```bash
go generate ./ent
```

**Known issue:** Ent's transitive dependency `clipperhouse/displaywidth@v0.6.2` is broken with Go 1.26. The `go.mod` has an `exclude` directive for it — do not remove it.

### Store Interface Pattern

`pkg/store/store.go` defines the domain types and a composite interface:

```go
type Store interface {
    ProgramStore
    InitiativeStore
    PhaseStore
    RMIStore
    AssignmentStore
    EvidenceStore
    RepositoryStore
}
```

When adding a new entity:

1. Add the domain struct (e.g., `Program`) with plain Go types (no Ent imports)
2. Add a sub-interface (e.g., `ProgramStore`) with CRUD methods
3. Embed the sub-interface in `Store`
4. Implement in both `MemStore` and `DoltStore` — they must stay in sync

**MemStore pattern:** `map[string]*Entity`, `sync.RWMutex`, check-then-insert. Copy on read to prevent aliasing.

**DoltStore pattern:** `entEntityToStore` converter function, Ent client queries. Eager-load edges with `.WithEdgeName()` on Get/List.

### Build Tags

DoltStore and CLI commands that import it require ICU C headers (via `dolthub/go-icu-regex`). Files are split by build tag:

| Tag | Files | Purpose |
|-----|-------|---------|
| `//go:build dolt` | `db_dolt.go`, `registry_dolt.go`, `helpers_dolt.go` | Production code needing DoltStore |
| `//go:build !dolt` | `db_nodolt.go`, `registry_nodolt.go` | Stubs (empty functions) for CI/test without ICU |

CI runs `go test ./pkg/...` (no dolt tag) — all domain logic is tested via MemStore. Local full builds use: `go build -tags dolt ./cmd/prismctl/`.

### ID Conventions

| Entity | Format | Example |
|--------|--------|---------|
| Program | `PROG-<SLUG>` | `PROG-MARKET-SPEC` |
| Initiative | `INIT-<REPOSLUG>-<NNN>` | `INIT-PRISMCONTROL-001` |
| Phase | `<INIT-ID>/phase-<N>` | `INIT-PRISMCONTROL-001/phase-3` |
| RMI | `RMI-<REPOSLUG>-<NNN>` | `RMI-PRISMCONTROL-042` |
| Assignment | `ASSIGN-<UUID>` | auto-generated |
| Repository | `github.com/<org>/<repo>` | `github.com/ProductBuildersHQ/prism-control` |

### CLI Command Pattern

Each entity gets a file in `cmd/prismctl/` with a parent command and subcommands:

```go
func entityCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "entity", Short: "Manage entities"}
    cmd.AddCommand(entityCreateCmd(), entityListCmd(), entityGetCmd(), entityUpdateCmd())
    return cmd
}
```

Register in `rootCmd()` in `main.go`. Commands call `connectService(cmd)` to get a `*service.Service`, then delegate to service methods. Use `tabwriter` for list output.

### MCP Tool Pattern

Tools are registered in `pkg/mcpserver/server.go`. Each tool is a pair of functions:

```go
func entityListTool() *mcp.Tool {
    return &mcp.Tool{
        Name:        "entity_list",
        Description: "List all entities.",
        InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
    }
}

func entityListHandler(svc *service.Service) mcp.ToolHandler {
    return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // call svc methods, marshal to JSON, return as text content
    }
}
```

Register both in `registerTools()`. The handler calls the same service methods as the CLI.

### Dashboard Pattern

- **Data structs** in `dashboard.go`: `initData`, `phaseData`, `rmiData`, `programData` wrap store types with computed fields (progress counts, display names)
- **Data loading** in `loadDashboardData()`: queries the store, builds maps, groups by program
- **Templates** in `dashboard_templates.go`: Go `html/template` strings (`summaryHTML`, `detailHTML`, `programHTML`) with inline CSS/JS
- **Routes**: `/` (summary), `/initiative/<id>` (detail), `/program/<id>` (program view)

### Test Strategy

- **Unit tests** (`pkg/store/memstore_test.go`, `pkg/service/*_test.go`): use `NewMemStore()` — no DB, no build tags, runs in CI
- **DoltStore integration**: guarded by `//go:build dolt`, excluded from CI. Run locally with ICU flags
- **Service tests**: use a `newTestService()` helper that wraps MemStore
- **Test pattern**: create entity → verify fields → test duplicate error → update → list → verify

### Product Repo Pointer

Product repos that participate in PRISM-tracked initiatives add this block to their `CLAUDE.md`:

```markdown
## PRISM Control

This repo's roadmap items are tracked in [prism-control](https://github.com/ProductBuildersHQ/prism-control). Use `prismctl work ready --repo github.com/ProductBuildersHQ/<this-repo>` to find claimable work, and carry the `Refs: RMI-<REPOSLUG>-<NNN>` trailer on every commit.
