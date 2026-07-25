# Architecture

## Package Layout

```
prism-control/
├── cmd/prismctl/          CLI (thin Cobra adapter)
├── pkg/
│   ├── store/             Store interface + in-memory fake
│   │   └── doltstore/     Ent-backed Dolt implementation (server + embedded)
│   ├── service/           Shared service layer (CLI and MCP call this)
│   ├── initiative/        Initiative lifecycle + phase status derivation
│   ├── assignment/        Lease-based work claims
│   ├── evidence/          Trailer parsing + attribution
│   ├── ingest/            Git commit + changelog ingestion
│   ├── report/            Initiative report generation
│   ├── validate/          Consistency checks
│   ├── mcpserver/         MCP server (9 tools, stdio transport)
│   ├── export/            JSONL snapshots
│   ├── release/           Dependency-ordered release plans
│   ├── reposcan/          Git repo scanning (via gogit)
│   └── roadmap/           ROADMAP.md parsing and generation
├── ent/                   Ent codegen (MySQL dialect → Dolt)
├── docs/
│   ├── specs/             PRD, TRD, PLAN, ROADMAP, WALKTHROUGH
│   ├── integrations/      VisionStudio, omnidevx, devfolio
│   └── sql/               Read-only SQL views
└── schema/                Generated JSON Schemas (//go:embed)
```

## Design Principles

### Library-First

All behavior lives in `pkg/*`. The CLI (`cmd/prismctl`) and MCP server (`pkg/mcpserver`) are thin adapters over one shared service layer (`pkg/service`). Both call identical service methods — there is no CLI-only or MCP-only logic.

### Store Interface

Domain logic depends only on `pkg/store.Store`, an interface with sub-interfaces:

- `InitiativeStore` — initiatives + initiative dependencies
- `PhaseStore` — phases within initiatives
- `RMIStore` — roadmap items + RMI dependencies
- `AssignmentStore` — lease-based work claims
- `EvidenceStore` — delivery evidence
- `RepositoryStore` — repository catalog + repo dependencies

Two implementations:

- **`MemStore`** — in-memory fake for unit testing; no Dolt dependency
- **`doltstore`** — Ent-backed production implementation with both embedded and server modes

### Unit-of-Work Pattern

The `UnitOfWork` interface wraps a SQL transaction with a subsequent Dolt commit. The production implementation issues `CALL DOLT_COMMIT` on success; the in-memory fake is a no-op. This ensures every logical operation is an atomic Dolt commit.

### Plan is Authoritative

The database is the source of truth for initiative/phase/RMI structure. Ingest never mutates plan structure — mismatches between ingested evidence and the plan are validation errors to surface, not silently reconcile.

### Phase Status is Derived

Phase status is always computed from member RMIs, never stored directly. This eliminates stale-status bugs and makes the derivation rules testable in isolation.

## Data Model

```
Initiative (1) ──── (N) Phase (1) ──── (N) RoadmapItem
     │                                        │
     │                                        │
     ├── InitiativeDependency                 ├── RMIDependency
     │   (source → target)                    │   (source → target)
     │                                        │
     └── Program (optional grouping)          ├── Assignment (lease)
                                              │
                                              └── DeliveryEvidence
                                                  (commit, PR, release, changelog)
```

## Integration Points

| Consumer | Interface | Access |
|----------|-----------|--------|
| Claude Code sessions | CLI (`prismctl`) or MCP (`prismctl mcp`) | Read/write |
| VisionStudio | SQL views over Dolt server | Read-only |
| omnidevx | SQL views + JSONL export | Read-only |
| devfolio | JSONL export | Read-only |

See [Integrations](integrations/visionstudio.md) for detailed setup guides.
