# Session Context — INIT-PRISMCONTROL-003

**Date:** 2026-08-01
**Initiative:** Unified Specs + Execution + Spend Platform

## Current State

**Initiative Status:** executing (5/21 RMIs completed)

### Completed This Session

1. **RMI-PRISMCONTROL-101** — Initiative.type enum field (committed, pushed)
2. **RMI-PRISMCONTROL-102** — SpecWorkflow entity (committed, pushed)
3. **RMI-PRISMCONTROL-103** — JudgeRubric entity (committed, pushed)
4. **RMI-PRISMCONTROL-104** — JudgeResult entity (committed, pushed)
5. **RMI-PRISMCONTROL-120** — Seed built-in workflows (committed, pushed)
   - Schema: `ent/schema/initiative.go` — added `init_type`, `workflow_id` fields
   - Constants: `pkg/initiative/initiative.go` — TypeFeature/Maintenance/Migration/Compliance/Refactor
   - Store: `pkg/store/store.go`, `pkg/store/doltstore/doltstore.go` — field mappings
   - Service: `pkg/service/initiative.go` — CreateInitiative accepts initType
   - CLI: `cmd/prismctl/initiative.go` — `--type` flag
   - MCP: `pkg/mcpserver/server.go` — `init_type` field

2. **Cancelled (duplicates INIT-TOKENATTRIB-001):**
   - RMI-105, RMI-106, RMI-110, RMI-111 — token entities/commands already delivered

### Phase 1 Complete

All Phase 1 schema RMIs delivered:

- RMI-101: Initiative.type
- RMI-102: SpecWorkflow entity
- RMI-103: JudgeRubric entity
- RMI-104: JudgeResult entity
- RMI-120: Seed built-in workflows

### Phase 2-4 Summary

- **Phase 2 (CLI):** spec init/scaffold, validate, judge, path standardization
- **Phase 3 (Maturity):** CapabilityModel, MaturityAssessment, assess/report commands
- **Phase 4 (visionstudio):** Extract UI components, dashboard, spend viz, radar chart

## Key Design Decisions

### D1 — Features = Initiatives with type=feature
No separate Feature entity. Long-lived capabilities spawn new initiatives.

### D2 — Canonical spec path
```
docs/specs/initiatives/{INIT-ID}/
├── PRD.md | TRD.md | PLAN.md
├── ROADMAP.md (prismctl syncs)
└── .judge/ (judge results)
```

### D3 — Workflow gradient
| Workflow | Specs | Default for |
|----------|-------|-------------|
| quick-fix | ROADMAP | maintenance |
| pbhq-lite | PLAN, ROADMAP | refactor, migration |
| pbhq-standard | PRD, TRD, PLAN, ROADMAP | feature, compliance |
| aws-working-backwards | PRFAQ, PRD, TRD, PLAN, ROADMAP | opt-in |

### D4 — Token attribution reuses INIT-TOKENATTRIB-001
- `pkg/tokens/` — TokenSource interface, JSONL reader, Dolt source
- `pkg/report/tokens.go` — report generation
- `prismctl report tokens` — initiative and period modes

## Next Steps

### Phase 2 — CLI spec commands

| RMI | Title |
|-----|-------|
| RMI-107 | `prismctl spec init` scaffolds specs from workflow |
| RMI-108 | `prismctl spec validate` checks specs exist |
| RMI-109 | `prismctl spec judge` runs LLM evaluation |
| RMI-112 | Standardize spec path in initiative (canonical path) |

### Phase 3 — Maturity model

| RMI | Title |
|-----|-------|
| RMI-113 | Add CapabilityModel entity |
| RMI-114 | Add MaturityAssessment entity |
| RMI-115 | `prismctl assess` command |
| RMI-116 | `prismctl maturity report` command |

### Phase 4 — visionstudio UI extraction

| RMI | Title |
|-----|-------|
| RMI-117 | Extract UI components from visionstudio |
| RMI-118 | Integrate spend visualization |
| RMI-119 | Add radar chart for maturity

## Related Context

### Prior Session Work (also pushed)
- INIT-PRISMCONTROL-002: Context assembly (`pkg/contextbuild/`)
- INIT-TOKENATTRIB-001: Token attribution (`pkg/tokens/`, `prismctl report tokens`)

### Parallel Initiative (omniagent)
- INIT-OMNIAGENT-002: OpenClaw Parity Sync
- Phase 1-2 completed (10/20 RMIs)
- Phases 3-4 pending (omnimemory, cron fixes, features)

### visionspec/visionstudio Integration
- visionspec: `~/go/src/github.com/ProductBuildersHQ/visionspec/`
  - Framework definitions in `docs/frameworks/`
  - Rubrics to import into JudgeRubric table
- visionstudio: `~/go/src/github.com/ProductBuildersHQ/visionstudio/`
  - Electron app, needs `packages/ui/` extraction (Phase 4)

## Commands to Resume

```bash
# Check initiative status
prismctl initiative get INIT-PRISMCONTROL-003

# List remaining RMIs
prismctl rmi list --initiative INIT-PRISMCONTROL-003

# Start next RMI
prismctl rmi update RMI-PRISMCONTROL-102 --status in_progress

# Regenerate ent after schema changes
cd ~/go/src/github.com/ProductBuildersHQ/prism-control
go generate ./ent
go build ./...
go test ./...
```
