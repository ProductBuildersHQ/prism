# Session Context — INIT-PRISMCONTROL-003

**Date:** 2026-08-01
**Initiative:** Unified Specs + Execution + Spend Platform

## Current State

**Initiative Status:** executing (9/21 RMIs completed)

### Completed This Session

**Phase 1 — Schema Foundation (complete):**

- RMI-101: Initiative.type enum field
- RMI-102: SpecWorkflow entity
- RMI-103: JudgeRubric entity
- RMI-104: JudgeResult entity
- RMI-120: Seed built-in workflows

**Phase 2 — CLI Commands (4/6 complete):**

- RMI-107: `prismctl spec init` — scaffolds specs from workflow
- RMI-108: `prismctl spec validate` — checks specs exist
- RMI-109: `prismctl spec judge show/record/list` — agent-driven evaluation
- RMI-112: Standardize canonical spec path onto initiative record

**Cancelled (duplicates INIT-TOKENATTRIB-001):**

- RMI-105, RMI-106, RMI-110, RMI-111 — token entities/commands already delivered

### Remaining Phase 2 RMIs

| RMI | Title | Notes |
|-----|-------|-------|
| RMI-??? | `prismctl spend report` | May already exist via `prismctl report tokens` |
| RMI-??? | Remaining spend CLI | Check against INIT-TOKENATTRIB-001 |

### Phase 3 — Maturity Model (not started)

| RMI | Title |
|-----|-------|
| RMI-113 | Add CapabilityModel entity |
| RMI-114 | Add MaturityAssessment entity |
| RMI-115 | `prismctl assess` command |
| RMI-116 | `prismctl maturity report` command |

### Phase 4 — visionstudio UI extraction (not started)

| RMI | Title |
|-----|-------|
| RMI-117 | Extract UI components from visionstudio |
| RMI-118 | Integrate spend visualization |
| RMI-119 | Add radar chart for maturity |

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

### D5 — Agent-driven judging (not API-driven)
Judging runs via interactive agent sessions (e.g., Claude Code) reading
specs and rubrics, not remote API calls. `spec judge show` presents the
content; `spec judge record` persists the verdict. Remote/CI judging
can be layered later without changing the contract.

## New CLI Commands

```bash
# Workflow management
prismctl workflow list
prismctl workflow get <id>
prismctl workflow seed

# Spec management
prismctl spec init <initiative-id> [--with-optional]
prismctl spec validate <initiative-id>
prismctl spec judge show <initiative-id> <spec-file>
prismctl spec judge record <initiative-id> <spec-file> --score <0-10> --rationale "..." --model <id>
prismctl spec judge list <initiative-id>
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
PRISMCTL_DSN="root:@tcp(127.0.0.1:13306)/prismcontrol" prismctl initiative get INIT-PRISMCONTROL-003

# List remaining RMIs
PRISMCTL_DSN="root:@tcp(127.0.0.1:13306)/prismcontrol" prismctl rmi list --initiative INIT-PRISMCONTROL-003

# Regenerate ent after schema changes
cd ~/go/src/github.com/ProductBuildersHQ/prism-control
go generate ./ent
go build ./...
go test ./...

# Build with dolt tag (requires ICU)
CGO_CFLAGS="-I/opt/homebrew/opt/icu4c/include" \
CGO_CXXFLAGS="-I/opt/homebrew/opt/icu4c/include" \
CGO_LDFLAGS="-L/opt/homebrew/opt/icu4c/lib" \
go build -tags dolt -o ~/go/bin/prismctl ./cmd/prismctl
```
