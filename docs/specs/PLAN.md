# PRISM Control — Implementation Plan (PLAN)

**Status:** Draft
**Date:** 2026-07-23
**Initiative:** `INIT-PRISMCONTROL-001`

Execution is organized by phase (see `ROADMAP.md` for the RMI-level breakdown). Each phase has a theme, exit criteria, and builds on the previous. We review and execute by phase, not by individual RMI.

## Phase 1 — Foundation & Storage

**Theme:** Prove the stack, model the domain.

- Ent + Dolt compatibility spike: Ent/Atlas migration vs Dolt DDL, CRUD, `CALL DOLT_COMMIT` via raw SQL, `dolt_history_*` queries. Decide migration mode (Atlas vs hand-managed `schema.sql`) before the schema hardens.
- Repository scaffold: module, root SDK facade skeleton, CI (build/test/lint), golangci-lint config.
- Core domain types with `json` tags; JSON Schemas generated via `invopop/jsonschema`, validated with `schemago lint`, embedded via `//go:embed`.
- Dolt schema + Ent storage layer for `initiatives`, `repositories`, `phases`, `roadmap_items`, `rmi_dependencies`, `assignments`, `delivery_evidence`.
- Store interface + in-memory fake; unit-of-work wrapper (Ent tx + Dolt commit).

**Exit criteria:** spike decision recorded in TRD; `go test ./...` green with domain + service tests running against the fake store; small Dolt integration suite passes locally.

## Phase 2 — Coordination Core (CLI)

**Theme:** Plan, claim, and update work through `prismctl`.

- `prismctl db init|serve`; `registry add|list`.
- `initiative create|list|get|status|transition` (lifecycle timestamps stamped on transition); `phase add|list` with derived status.
- `rmi create|update|get|list` with dependency edges and acceptance criteria.
- `work ready` (ready ∧ unblocked ∧ unclaimed, `--repo`/`--initiative` filters).
- `claim` / `renew` / `release` with lease enforcement in the service layer; `claim` prints the `Refs:` trailer.
- `update` with status, evidence refs, handoff JSON.
- `export` JSONL snapshots + manifest into `exports/`.

**Exit criteria:** two terminal sessions can concurrently claim different RMIs of one initiative with no collisions; a simulated dead session's lease expires and the RMI returns to `work ready` output.

## Phase 3 — Agent Interface

**Theme:** Make Claude Code sessions first-class users.

- Session protocol written into this repo's `CLAUDE.md` (plan → claim → trailer → update → handoff); one-line pointer added to product repos' `CLAUDE.md`.
- `prismctl mcp` stdio server via the official Go MCP SDK: `initiative_list/get/create`, `rmi_create`, `work_ready`, `task_claim/release`, `task_update`, `report_initiative`. Same service methods as the CLI.
- `.mcp.json` template for product repos.

**Exit criteria:** a real Claude Code session plans a toy initiative, claims a task via MCP, commits with the trailer, and updates status — end to end without manual SQL or file editing.

## Phase 4 — Attribution & Reporting

**Theme:** Close the loop from plan to evidence to report.

- Trailer + conventional-commit parser (`pkg/evidence`), attribution precedence (trailer → PR ref → branch name).
- `ingest git` with per-repo high-water marks; `ingest changelog` for `CHANGELOG.json`.
- `report initiative` (JSON + Markdown): duration, phases, repos, commit distribution by type/scope/repo, releases, unattributed residual.
- `validate`: trailer↔RMI consistency, dangling/cyclic dependencies, expired leases, status coherence.

**Exit criteria:** report for a seeded initiative matches hand-computed expectations; ingest is idempotent (re-runs add nothing).

## Phase 5 — Pilot & Dogfood

**Theme:** Run real initiatives; fix what the pilot breaks.

- Pilot: model `INIT-OMNIDEVX-EFFECTIVENESS-001` (omnidevx, omnidevx-core, omni-github, omni-openai, devfolio) and drive it with concurrent Claude Code sessions.
- Dogfood: migrate this repo's `docs/specs/ROADMAP.md` into the running system as `INIT-PRISMCONTROL-001` (RMI IDs already match, so this is a data import).
- Protocol/tooling refinements from pilot friction; measure the success metrics from the PRD (claim collisions, unattributed-commit rate).

**Exit criteria:** PRD §8 success metrics met on the pilot; this project tracks itself.

## Phase 6 — Ecosystem Integrations (post-v1, planned)

**Theme:** Feed the consumers.

- VisionStudio: read-only SQL datasource documentation + any needed views for initiative browsing.
- omnidevx: per-initiative token/effort join via `assignments.worker` session IDs.
- devfolio: initiative dimension on velocity reports; structured-changelog RMI-ref schema extension upstream.
- Release sets + dependency-ordered release planning (tables from the ideation design).
- `ROADMAP.json` / `ROADMAP.md` projections into product repos.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Ent/Atlas migrations incompatible with Dolt DDL | Blocks storage layer | Phase-1 spike is the first RMI; fallback: hand-managed `schema.sql`, Ent no-migration mode |
| Trailer discipline slips (humans, squash merges) | Attribution quality degrades | Protocol + `claim` prints trailer; PR-body/branch fallbacks; unattributed residual surfaced in every report; CI warning → error later |
| `dolt sql-server` not running when a session starts | Sessions blocked | `prismctl db serve` helper; `exports/` JSONL as read-only fallback context |
| MCP tool schemas bloat session context | Slower/costlier sessions | Keep tool surface to the 8 core tools; everything else stays CLI-only |
| Scope creep toward Beads/releases/UI | v1 never ships | Non-goals pinned in PRD §6; release tables explicitly deferred |

## Sequencing Notes

- The pilot (Phase 5) intentionally starts only after reporting (Phase 4) so the pilot produces the first real end-to-end report.
- Commit-trailer convention starts **now** — commits to this repo reference `RMI-PRISMCONTROL-NNN` from Phase 1 onward, so the dogfood migration retroactively attributes the project's own history.
