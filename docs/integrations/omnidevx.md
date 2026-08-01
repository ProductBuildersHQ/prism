# omnidevx Integration

This document describes how omnidevx-core joins with PRISM Control to
attribute per-session token usage and effort to initiatives.

## Join Model

omnidevx-core collects Claude Code session events, each carrying an
`EventContext.SessionID` (a UUID assigned by Claude Code to each
conversation). PRISM Control tracks work assignments with a `worker`
field on each assignment row.

The join between the two systems is:

```
omnidevx Event.Context.SessionID  =  prism assignments.worker
```

When a Claude Code session claims an RMI via `prismctl work claim`, the
`--worker` value is stored in `assignments.worker`. If that value is the
Claude Code session UUID, omnidevx can attribute every token from that
session to the claimed RMI and, by extension, to its initiative.

## Worker Auto-Detection

Claude Code exposes the session UUID via the `CLAUDE_CODE_SESSION_ID`
environment variable. When claiming work, `prismctl` auto-detects this:

```bash
# The session UUID is automatically used when CLAUDE_CODE_SESSION_ID is set
prismctl work claim RMI-MYREPO-042 --lease-hours 4

# Or explicitly provide it if needed
prismctl work claim RMI-MYREPO-042 \
  --worker "$CLAUDE_CODE_SESSION_ID" \
  --lease-hours 4
```

The auto-detection enables seamless attribution: every token event from
the claiming session is automatically attributed to the claimed RMI.

## Token Attribution Reports

Use `prismctl report tokens` to generate attribution reports:

```bash
# Initiative mode: per-RMI and per-model breakdown
prismctl report tokens --initiative INIT-PRISM-001 --format markdown

# Quarter mode: all initiatives in the period
prismctl report tokens --quarter 2026-Q3

# Custom date range
prismctl report tokens --since 2026-07-01 --until 2026-07-31

# JSON output for programmatic use
prismctl report tokens --initiative INIT-PRISM-001 --format json
```

Reports show:

- **Attributed spend**: tokens/cost mapped to RMIs via session→assignment
- **Residual**: tokens matched to a repository but not a specific RMI
- **Unmanaged**: tokens from workspaces not in the registry
- **Coverage**: managed spend ÷ total spend

## Legacy Worker IDs

Prior to auto-detection, sessions used timestamp-based IDs like
`"session-$(date +%s)"`. These don't match Claude Code session UUIDs
and won't attribute correctly. Update old assignments or re-claim work
with the proper session UUID to enable attribution.

## SQL View

The `v_assignment_sessions` view (defined in `docs/sql/omnidevx-views.sql`)
flattens the assignment-to-initiative path into a single row:

| Column         | Source                              | Description                        |
|----------------|-------------------------------------|------------------------------------|
| assignment_id  | assignments.assignment_id           | Unique assignment ID               |
| rmi_id         | roadmap_items.rmi_id                | Roadmap Item ID                    |
| initiative_id  | roadmap_items.initiative_roadmap_items | Parent initiative ID            |
| worker         | assignments.worker                  | Session ID / worker identifier     |
| workspace      | assignments.workspace               | Optional workspace path            |
| status         | assignments.status                  | Assignment status                  |
| created_at     | assignments.created_at              | When the assignment was created    |
| completed_at   | assignments.completed_at            | When the assignment was completed  |

## Example Join Query

Given a set of omnidevx events loaded into a `session_tokens` table (or
computed in-memory), join against the view to get per-initiative totals:

```sql
SELECT
    vas.initiative_id,
    SUM(st.input_tokens)  AS total_input_tokens,
    SUM(st.output_tokens) AS total_output_tokens,
    SUM(st.cost_usd)      AS total_cost,
    COUNT(DISTINCT st.session_id) AS session_count
FROM v_assignment_sessions vas
JOIN session_tokens st
    ON vas.worker = st.session_id
GROUP BY vas.initiative_id;
```

In Go, the equivalent join is done in-memory: build a
`map[string]string` (session ID to initiative ID) from the view rows,
then pass it to `report.InitiativeReportFromEvents()` in omnidevx-core.
