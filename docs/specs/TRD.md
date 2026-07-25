# PRISM Control — Technical Requirements Document (TRD)

**Status:** Draft
**Date:** 2026-07-23
**Initiative:** `INIT-PRISMCONTROL-001`

## 1. Architecture Overview

```text
                        ┌──────────────────────────────┐
                        │   Consumers                  │
                        │   VisionStudio (SQL, RO)     │
                        │   devfolio / omnidevx (join) │
                        │   future web UI (SQL, RO)    │
                        └──────────────┬───────────────┘
                                       │
┌───────────────┐   ┌──────────────────▼───────────────┐
│ Claude Code   │──►│  prism-control                   │
│ sessions      │   │                                  │
│ (MCP + CLI)   │   │  SDK: root package facade        │
└───────────────┘   │  pkg/*: domain + service (logic) │
                    │  store interface ─► Ent ─► Dolt  │
┌───────────────┐   │                                  │
│ humans / CI   │──►│  adapters: cmd/prismctl (Cobra)  │
│ (CLI)         │   │            prismctl mcp (go-sdk) │
└───────────────┘   └──────────────────────────────────┘
```

### Principles

1. **Library-first.** All behavior lives in importable Go packages (the SDK). `cmd/prismctl` and the MCP server are thin adapters over one shared service layer; CLI and MCP write paths call identical service methods so protocol rules (lease checks, status transitions) cannot diverge.
2. **Headless.** No UI in this repo. Consumers read the Dolt SQL server or generated artifacts.
3. **Plan is authoritative; commits are evidence.** Ingest never mutates plan structure. A commit referencing an RMI that belongs to a different repo is a validation error, not new data.
4. **Unit-testable core.** Domain logic (dependency graphs, lease expiry, trailer parsing, report computation, status derivation) is pure or depends only on the store interface, so it tests without Dolt.

## 2. Storage

### Dolt

- One canonical Dolt database for the whole portfolio (not per-repo).
- Local `dolt sql-server` on the developer machine; all sessions and VisionStudio connect via MySQL wire protocol. Claims are plain SQL transactions — no file/git locking.
- Every logical write = SQL transaction + `CALL DOLT_COMMIT(...)` with a structured message (`actor`, `operation`, `ids`). This yields a full audit history and enables `dolt_history_*` / `dolt_diff_*` / `AS OF` queries for progress-over-time reporting.
- **Portability rule (NFR4):** reports and consumer-facing queries use plain SQL over base tables plus explicit lifecycle timestamps. Dolt system tables are an enhancement (time-series views), never a dependency.

### Backup and portability

- Dolt remote push (DoltHub, file remote, or git remote via `refs/dolt/data`).
- `prismctl export` writes JSONL snapshots (one file per table) plus a manifest (`dolt_commit`, `exported_at`, record counts) into `exports/`, committed to this git repo. Snapshots double as offline agent context and CI validation input.

## 3. Data Model

ID conventions:

- Initiative: `INIT-<NAME>-<NNN>` (e.g., `INIT-PRISMCONTROL-001`).
- RMI: `RMI-<REPOSLUG>-<NNN>` where `REPOSLUG` is the uppercased alphanumeric repo name (e.g., `RMI-PRISMCONTROL-014`, `RMI-OMNIDEVXCORE-003`). IDs are allocated centrally and never reused.
- Assignment: `ASSIGN-<NNNN>`.

### Tables (v1)

```sql
CREATE TABLE initiatives (
    initiative_id     VARCHAR(64) PRIMARY KEY,
    organization      VARCHAR(128) NOT NULL,      -- future tenancy scoping
    title             VARCHAR(255) NOT NULL,
    description       TEXT,
    status            VARCHAR(32) NOT NULL,       -- lifecycle below
    priority          VARCHAR(32),
    -- explicit lifecycle timestamps (do not rely on dolt history)
    created_at        DATETIME NOT NULL,
    planned_at        DATETIME,
    executing_at      DATETIME,
    delivery_complete_at DATETIME,
    released_at       DATETIME,
    closed_at         DATETIME,
    updated_at        DATETIME NOT NULL
);

CREATE TABLE repositories (
    repository_id     VARCHAR(128) PRIMARY KEY,   -- e.g. github.com/plexusone/omnidevx
    organization      VARCHAR(128) NOT NULL,
    repository_name   VARCHAR(128) NOT NULL,
    default_branch    VARCHAR(128) NOT NULL DEFAULT 'main',
    local_path        VARCHAR(512),               -- absolute path on disk (for ingest/scan)
    go_module         VARCHAR(256),               -- go.mod module path (from gitscan)
    domain            VARCHAR(128),
    status            VARCHAR(32) NOT NULL,
    ingest_high_water VARCHAR(128),               -- last scanned commit SHA
    UNIQUE (organization, repository_name)
);

CREATE TABLE repository_dependencies (
    source_repository_id VARCHAR(128) NOT NULL,   -- source depends on target
    target_repository_id VARCHAR(128) NOT NULL,
    dependency_type      VARCHAR(32) NOT NULL,     -- go_module
    PRIMARY KEY (source_repository_id, target_repository_id, dependency_type),
    FOREIGN KEY (source_repository_id) REFERENCES repositories(repository_id),
    FOREIGN KEY (target_repository_id) REFERENCES repositories(repository_id)
);

CREATE TABLE phases (
    phase_id          VARCHAR(64) PRIMARY KEY,    -- e.g. INIT-PRISMCONTROL-001/phase-1
    initiative_id     VARCHAR(64) NOT NULL,
    sequence_number   INT NOT NULL,
    title             VARCHAR(255) NOT NULL,
    theme             VARCHAR(255),               -- the grouping rationale
    -- NOTE: no status column; phase status is ALWAYS derived from member RMIs
    FOREIGN KEY (initiative_id) REFERENCES initiatives(initiative_id)
);

CREATE TABLE roadmap_items (
    rmi_id            VARCHAR(64) PRIMARY KEY,
    repository_id     VARCHAR(128) NOT NULL,
    initiative_id     VARCHAR(64),
    phase_id          VARCHAR(64),
    title             VARCHAR(255) NOT NULL,
    description       TEXT,
    item_type         VARCHAR(32) NOT NULL,       -- capability|task|spec|release
    status            VARCHAR(32) NOT NULL,
    priority          VARCHAR(32),
    required          BOOLEAN NOT NULL DEFAULT TRUE,
    sequence_number   INT,
    acceptance_criteria JSON,                     -- array of strings
    created_at        DATETIME NOT NULL,
    completed_at      DATETIME,
    updated_at        DATETIME NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(repository_id),
    FOREIGN KEY (initiative_id) REFERENCES initiatives(initiative_id),
    FOREIGN KEY (phase_id) REFERENCES phases(phase_id)
);

CREATE TABLE rmi_dependencies (
    source_rmi_id     VARCHAR(64) NOT NULL,       -- source depends on target
    target_rmi_id     VARCHAR(64) NOT NULL,
    relationship      VARCHAR(32) NOT NULL,       -- requires|relates
    PRIMARY KEY (source_rmi_id, target_rmi_id, relationship)
);

CREATE TABLE assignments (
    assignment_id     VARCHAR(64) PRIMARY KEY,
    rmi_id            VARCHAR(64) NOT NULL,
    worker            VARCHAR(128) NOT NULL,      -- session ID; MUST match omnidevx claudecode session identifier
    status            VARCHAR(32) NOT NULL,       -- active|released|expired|completed
    lease_expires_at  DATETIME NOT NULL,
    workspace         VARCHAR(512),               -- local path / worktree
    handoff           JSON,                       -- completed/remaining/decisions/next_action
    created_at        DATETIME NOT NULL,
    updated_at        DATETIME NOT NULL,
    FOREIGN KEY (rmi_id) REFERENCES roadmap_items(rmi_id)
);

CREATE TABLE delivery_evidence (
    evidence_id       VARCHAR(128) PRIMARY KEY,
    rmi_id            VARCHAR(64) NOT NULL,
    evidence_type     VARCHAR(32) NOT NULL,       -- commit|pr|release|changelog|test
    reference         VARCHAR(512) NOT NULL,      -- SHA, URL, tag, changelog entry ID
    commit_type       VARCHAR(32),                -- conventional commit type (commits only)
    commit_scope      VARCHAR(64),
    occurred_at       DATETIME,
    created_at        DATETIME NOT NULL,
    FOREIGN KEY (rmi_id) REFERENCES roadmap_items(rmi_id)
);
```

Deferred to a later phase (designed in the ideation doc, not built in v1): `release_sets`, `component_releases`, `release_dependencies`, `release_validations`.

### Status vocabularies

- **Initiative lifecycle:** `proposed → planned → executing → delivery_complete → releasing → released → closed` (plus `cancelled`). Each forward transition stamps its timestamp column.
- **RMI:** `planned | ready | in_progress | blocked | completed | cancelled`.
- **Phase (derived only):** `completed` (all required RMIs completed) · `in_progress` (any RMI active) · `blocked` (a required RMI blocked) · `planned` (none started) · `partial` (required done, optional remain).
- **Assignment:** `active | released | expired | completed`.

## 4. Ent Layer

- Ent (`entgo.io/ent`) generates the typed query layer against the schema above; MySQL dialect targets Dolt.
- **Phase-1 spike (RMI-PRISMCONTROL-001) — COMPLETED.** Result: Ent's MySQL dialect is fully compatible with Dolt 2.2.2. DDL (`CREATE TABLE`), CRUD, SQL transactions, `CALL DOLT_COMMIT` / `CALL DOLT_ADD`, `dolt_history_*` queries, and `AS OF` time-travel queries all work. No Atlas/migration fallback needed; Ent codegen + standard MySQL driver (`go-sql-driver/mysql`) connects to `dolt sql-server` without issues. The unit-of-work pattern (SQL tx commit → `DOLT_ADD` + `DOLT_COMMIT`) is confirmed viable.
- **Unit of work:** a service-layer wrapper runs `fn(tx)` inside an Ent transaction then issues the Dolt commit. All mutating service methods go through it.

## 5. Package Layout (SDK)

Module: `github.com/ProductBuildersHQ/prism-control`. Root package `prismcontrol` is the SDK facade (mirrors the omnidevx pattern): `prismcontrol.New(cfg) (*Client, error)` exposing the service API consumed by CLI, MCP, and external importers.

```text
prism-control/
├── prismcontrol.go        # SDK facade: Client, config, constructors
├── cmd/prismctl/          # Cobra CLI (thin adapter)
├── pkg/
│   ├── initiative/        # initiative + phase domain: lifecycle, derived phase status
│   ├── rmi/               # RMI domain: IDs, statuses, dependency graph, ready-work logic
│   ├── assignment/        # leases: claim, renew, release, expiry
│   ├── evidence/          # evidence types; trailer spec + parser
│   ├── ingest/            # git-history scan (high-water marks), CHANGELOG.json ingest
│   ├── report/            # initiative report computation + JSON/Markdown rendering
│   ├── registry/          # repository catalog
│   ├── reposcan/          # bulk registry import via gogit/scanner (gitscan integration)
│   ├── export/            # JSONL snapshots + manifest
│   ├── store/             # storage interface + in-memory fake (for unit tests)
│   │   └── doltstore/     # Ent implementation, unit-of-work, Dolt commits
│   └── mcpserver/         # MCP tools over the service layer (importable)
├── ent/                   # Ent schema + generated code
├── schema/                # generated JSON Schemas (//go:embed) for export/report types
├── exports/               # JSONL snapshots (generated, committed)
└── docs/specs/            # PRD, TRD, PLAN, ROADMAP
```

JSON-emitting types (report, export, handoff) follow the Go-first schema workflow: structs are the source of truth, `invopop/jsonschema` generates `schema/*.schema.json`, `schemago lint` validates, schemas embedded via `//go:embed`.

## 6. CLI Surface (`prismctl`, Cobra)

```text
prismctl db init|serve|commit-log         # bootstrap Dolt, run sql-server helper, audit log
prismctl registry add|list|scan|deps|unpushed  # repository catalog + gitscan integration
prismctl initiative create|list|get|status|transition
prismctl phase add|list                   # themed groupings within an initiative
prismctl rmi create|update|get|list       # RMIs + dependency edges (--depends-on)
prismctl work ready [--repo|--initiative] # ready ∧ unblocked ∧ unclaimed
prismctl claim <rmi> --worker <session> --lease 4h    # prints the Refs: trailer
prismctl release|renew <assignment>
prismctl update <rmi> --status ... --evidence pr:URL --handoff file.json
prismctl ingest git --repo <path> [--all] # trailer scan from high-water mark
prismctl ingest changelog --repo <path>   # CHANGELOG.json (structured-changelog)
prismctl report initiative <id> [--format json|markdown]
prismctl validate                         # consistency checks (see §9)
prismctl export                           # JSONL snapshots to exports/
prismctl mcp                              # stdio MCP server
```

## 7. MCP Server (`prismctl mcp`)

Official Go SDK (`github.com/modelcontextprotocol/go-sdk`), stdio transport, registered per session via `.mcp.json`. Tools mirror protocol verbs and call the same service methods as the CLI. Keep the surface small — schemas cost context in every session:

| Tool | Purpose |
|------|---------|
| `initiative_list` / `initiative_get` | Browse initiatives, derived phase progress |
| `initiative_create` / `rmi_create` | Plan: decompose initiative into phased, per-repo RMIs |
| `work_ready` | Ready ∧ unblocked ∧ unclaimed, filterable by repo |
| `task_claim` / `task_release` | Lease claim; returns the `Refs:` trailer to use |
| `task_update` | Status, evidence refs, handoff notes |
| `report_initiative` | End-to-end report (JSON) |

## 8. Session Protocol and Commit Attribution

The protocol is documented in this repo's `CLAUDE.md` (product repos' `CLAUDE.md` get one pointer line):

1. **Plan** — create/verify the initiative, phases, and per-repo RMIs before implementation work starts (duration measurement begins at creation).
2. **Claim** — `task_claim` with the session identifier and a lease (default 4h). The response includes the exact trailer line.
3. **Execute** — work in the product repo. Every commit made under a claim carries the trailer.
4. **Update** — `task_update` with status, evidence, and a compact handoff (`completed` / `remaining` / `decisions` / `next_action`); release or renew the lease.

### Trailer specification

```text
feat(compositor): add alpha-channel overlay support

Refs: RMI-VIDEOASCODE-019
```

- Git trailer with key `Refs:`; value is one RMI ID (comma-separated if a commit genuinely serves several). Conventional-Commits-compliant (footers follow git trailer format).
- RMI only — never the initiative ID; the RMI→initiative edge is resolved in the database at report time.
- Subject line stays clean; extraction is `git log --format='%H|%(trailers:key=Refs,valueonly)'`.
- Squash-merge backups: RMI in the PR body; branch naming `rmi/<repo>-<nnn>-<slug>`. Attribution precedence: trailer → PR reference → branch name.
- Not every commit requires a trailer; the report exposes an **unattributed residual** (commits in participating repos during the initiative window without attribution) as the data-quality signal.

### Ingest

- Scan-once, query-forever: `prismctl ingest git` walks commits since `repositories.ingest_high_water`, parses trailers and conventional-commit type/scope, writes `delivery_evidence` rows, advances the high-water mark. All downstream queries are SQL; consumers never touch git.
- Changelog ingest maps `CHANGELOG.json` (structured-changelog) entries to RMIs via the trailer-populated reference (schema extension to structured-changelog: optional RMI ref field).

## 9. Validation (`prismctl validate`)

- Trailer references a non-existent RMI, or an RMI whose repo ≠ the commit's repo (evidence never mutates the plan — these are errors to surface).
- Dangling `rmi_dependencies` edges; dependency cycles.
- Expired `active` leases (auto-expire + report).
- Status coherence: initiative `delivery_complete` while required RMIs are open; `completed` RMIs lacking required evidence.
- Duplicate IDs; ID-format violations.

## 10. Initiative Report

`prismctl report initiative INIT-X` (also MCP `report_initiative`). Computation in `pkg/report` from store data only (portable per NFR4).

```json
{
  "initiative_id": "INIT-OMNIDEVX-EFFECTIVENESS-001",
  "duration": {
    "created": "…", "executing": "…", "delivery_complete": "…",
    "days_executing": 34
  },
  "phases": [
    {"phase_id": "phase-1", "theme": "…", "status": "completed", "rmis_completed": 5, "rmis_total": 5}
  ],
  "repos": {"count": 5, "list": ["…"]},
  "rmis": {"total": 14, "completed": 14, "required_completed": 12},
  "commits": {
    "total": 87,
    "by_type": {"feat": 31, "fix": 18, "test": 15, "docs": 12, "refactor": 8, "chore": 3},
    "by_repo": {"omnidevx-core": 29},
    "unattributed_in_window": 4
  },
  "releases": [{"repo": "omnidevx-core", "version": "v0.4.0"}]
}
```

## 11. Testing Strategy

Unit-first, enabled by the layering:

- **Pure domain logic** (highest coverage): dependency graph / ready-work / cycle detection (`pkg/rmi`), lease expiry (`pkg/assignment`), trailer + conventional-commit parsing (`pkg/evidence`), report computation (`pkg/report`), derived phase status (`pkg/initiative`) — table-driven tests, no I/O.
- **Service layer:** tested against the in-memory fake store (`pkg/store`), covering protocol rules (claim conflicts, invalid transitions, evidence validation).
- **Store implementation:** a deliberately small integration suite against a real local Dolt (schema migration, unit-of-work + `DOLT_COMMIT`, history queries) — the only tests requiring Dolt, skippable via build tag when unavailable.
- **Adapters:** CLI/MCP tests assert argument→service-call mapping, not behavior (behavior is tested once, in the service layer).

Error handling follows the global standard: return errors up; panic only on invariant violations; `t.Fatal` in tests; never discard errors.

## 12. Dependencies

Per the dependency-verification standard, confirm latest versions at implementation time (`go list -m -versions`, GitHub releases) — do not copy versions from this document:

- `entgo.io/ent` — ORM/codegen (MySQL dialect → Dolt)
- `github.com/spf13/cobra` — CLI
- `github.com/modelcontextprotocol/go-sdk` — MCP server
- `github.com/invopop/jsonschema` + `github.com/grokify/schemago` — schema generation/lint
- `github.com/grokify/gogit` — repository scanning, dependency graph, topological sort (gitscan integration)
- `github.com/grokify/structured-changelog` — changelog parsing (via `schangelog` conventions)
- Dolt binary (external runtime dependency, not a Go module)

## 13. gitscan Integration (`pkg/reposcan`)

`github.com/grokify/gogit/scanner` is imported as a Go library (not called via CLI). Key integration points:

- **Bulk registry import** (`prismctl registry scan <path>`): `scanner.ScanDirectoryWithProgress` scans an org directory (e.g., `~/go/src/github.com/plexusone`) and auto-registers all git repos found — populating `repository_name`, `organization`, `local_path`, `go_module`, and `status` fields. Repos already in the registry are updated (path, module, status); new repos are inserted.
- **Dependency graph** (`prismctl registry deps`): `scanner.TopologicalSort` + the `Dependencies` field from go.mod parsing populate `repository_dependencies` rows. Only dependencies between repos within the registered set are stored (external module deps are ignored). This feeds the release-ordering feature (Phase 6, RMI-026).
- **Impact analysis**: `scanner.GetTransitiveDependents(seeds, all)` answers "if I change repo X, which downstream repos in the registry need attention?" — available as a service method and future MCP tool.
- **Unpushed detection** (`prismctl registry unpushed`): `RepoResult.HasUncommittedChanges` / `HasUnpushedCommits` flags repos with un-pushed work. Combined with active assignments, this surfaces "claimed but not pushed" as a validation warning.
- **Activity filtering**: `RepoResult.LatestModTime` enables `--since` filtering on scan output, so only recently active repos surface during planning.

### Data flow

```text
~/go/src/github.com/{org}/
        │
        ▼
  gogit/scanner.ScanDirectoryWithProgress
        │
        ▼
  pkg/reposcan.Import(results) ──► repositories table
        │                          repository_dependencies table
        ▼
  prismctl registry scan ──► Dolt commit
```

## 14. Multi-Tenancy Readiness (not implemented in v1)

- `organization` columns on `initiatives` and `repositories` from day one; single-tenant v1 uses one constant value.
- Portability rule NFR4 keeps consumer queries valid against a plain MySQL read replica for a future hosted UI.
- No auth in v1; the Dolt SQL server binds to localhost.
