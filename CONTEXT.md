# Session Context — INIT-PRISMCONTROL-003

**Date:** 2026-08-01
**Initiative:** Unified Specs + Execution + Spend Platform

## Current State

**Initiative Status:** executing (13/17 RMIs completed, 4 cancelled)

### Completed This Session

**Phase 1 — Schema Foundation (complete):**

- RMI-101: Initiative.type enum field
- RMI-102: SpecWorkflow entity
- RMI-103: JudgeRubric entity
- RMI-104: JudgeResult entity
- RMI-120: Seed built-in workflows

**Phase 2 — CLI Commands (complete):**

- RMI-107: `prismctl spec init` — scaffolds specs from workflow
- RMI-108: `prismctl spec validate` — checks specs exist
- RMI-109: `prismctl spec judge show/record/list` — agent-driven evaluation
- RMI-121: Standardize canonical spec path (`docs/specs/initiatives/{INIT-ID}/`)

**Phase 3 — Maturity Model (complete):**

- RMI-112: CapabilityModel entity + built-in models (big-tech-essentials, big-tech-full, continuous-discovery, api-first)
- RMI-113: MaturityAssessment entity with dimension scores
- RMI-114: `prismctl maturity assess show/record/list` — agent-driven assessment
- RMI-115: `prismctl maturity report` — JSON + table output for radar charts

**Cancelled (duplicates INIT-TOKENATTRIB-001):**

- RMI-105, RMI-106, RMI-110, RMI-111 — token entities/commands already delivered

### Phase 4 — visionstudio UI extraction (not started)

| RMI | Title |
|-----|-------|
| RMI-116 | Extract shared UI components from visionstudio |
| RMI-117 | Execution dashboard panel |
| RMI-118 | Spend visualization panel |
| RMI-119 | Maturity radar chart panel |

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

# Maturity management
prismctl maturity model list
prismctl maturity model get <model-id>
prismctl maturity model seed

prismctl maturity assess show <initiative-id> <model-id>
prismctl maturity assess record <initiative-id> <model-id> --scores '{}' --overall <n> --summary "..." --assessed-by "..." --model <llm>
prismctl maturity assess list <initiative-id>

prismctl maturity report <initiative-id> [--json]
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
