# Session Context — INIT-PRISMCONTROL-003

**Date:** 2026-08-01
**Initiative:** Unified Specs + Execution + Spend Platform

## Current State

**Initiative Status:** executing (1/21 RMIs completed)

### Completed This Session

1. **RMI-PRISMCONTROL-101** — Initiative.type enum field (committed, pushed)
   - Schema: `ent/schema/initiative.go` — added `init_type`, `workflow_id` fields
   - Constants: `pkg/initiative/initiative.go` — TypeFeature/Maintenance/Migration/Compliance/Refactor
   - Store: `pkg/store/store.go`, `pkg/store/doltstore/doltstore.go` — field mappings
   - Service: `pkg/service/initiative.go` — CreateInitiative accepts initType
   - CLI: `cmd/prismctl/initiative.go` — `--type` flag
   - MCP: `pkg/mcpserver/server.go` — `init_type` field

2. **Cancelled (duplicates INIT-TOKENATTRIB-001):**
   - RMI-105, RMI-106, RMI-110, RMI-111 — token entities/commands already delivered

### Remaining Phase 1 RMIs

| RMI | Title | Status |
|-----|-------|--------|
| RMI-102 | Add SpecWorkflow entity | proposed |
| RMI-103 | Add JudgeRubric entity | proposed |
| RMI-104 | Add JudgeResult entity | proposed |
| RMI-120 | Seed built-in spec workflows | proposed |

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

### Immediate (Phase 1 completion)

1. **RMI-102: SpecWorkflow entity**
   ```go
   // ent/schema/specworkflow.go
   SpecWorkflow {
       id            string  // "pbhq-lite"
       name          string
       description   string
       specs_required []string  // JSON: ["PLAN.md", "ROADMAP.md"]
       specs_optional []string  // JSON: ["PRD.md", "TRD.md"]
       init_types    []string  // JSON: ["refactor", "migration"]
   }
   ```

2. **RMI-103: JudgeRubric entity**
   ```go
   // FK to SpecWorkflow
   JudgeRubric {
       id              string
       workflow_id     string  // FK
       spec_type       string  // "PRD.md"
       criteria        JSON    // scoring dimensions
       prompt_template string  // LLM prompt
   }
   ```

3. **RMI-104: JudgeResult entity**
   ```go
   JudgeResult {
       id              string
       initiative_id   string  // FK
       spec_path       string
       rubric_id       string  // FK
       score           float
       rationale       string
       model           string
       evaluated_at    time
   }
   ```

4. **RMI-120: Seed workflows**
   - Add to `pkg/store/doltstore/` or `cmd/prismctl/db.go` init
   - quick-fix, pbhq-lite, pbhq-standard, aws-working-backwards
   - Map Initiative.type → default workflow

### Files to Create/Modify

```
ent/schema/
├── specworkflow.go      # NEW
├── judgerubric.go       # NEW
├── judgeresult.go       # NEW
└── initiative.go        # Add edge to workflow

pkg/store/
├── store.go             # Add SpecWorkflow, JudgeRubric, JudgeResult types
└── doltstore/doltstore.go  # Add CRUD for new entities

cmd/prismctl/
├── workflow.go          # NEW: workflow list/get commands
└── spec.go              # NEW (Phase 2): spec init/validate/judge
```

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
