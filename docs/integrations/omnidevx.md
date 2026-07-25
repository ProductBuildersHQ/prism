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

## Current Limitation

The `worker` field is human-chosen. The session protocol suggests
`"session-$(date +%s)"` for convenience, but that timestamp-based ID
does not match the UUID that omnidevx reads from Claude Code JSONL
history files.

For accurate joins, sessions should use the actual Claude Code session
UUID as the worker ID when claiming work:

```bash
# The Claude Code session UUID is available in the JSONL filename or
# the sessionId field inside the session's conversation file.
prismctl work claim RMI-MYREPO-042 \
  --worker "01234567-89ab-cdef-0123-456789abcdef" \
  --lease-hours 4
```

Until all sessions adopt UUID-based worker IDs, the join is partial:
only assignments whose `worker` value is a valid Claude Code session
UUID will match omnidevx events.

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
