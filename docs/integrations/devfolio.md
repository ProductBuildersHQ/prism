# DevFolio Integration

DevFolio generates contributor profiles enriched with PRISM Control initiative and RMI data. This document describes the data pipeline from prism-control to devfolio and how structured-changelog entries participate.

## Architecture

```
prism-control (Dolt DB)
    |
    | prismctl export --format jsonl
    v
export.jsonl
    |
    | devfolio datasource/prism.Load()
    v
devfolio contributor profiles
```

## JSONL Export Format

`prismctl export` produces a newline-delimited JSON file. Each line is a self-contained record with a `kind` discriminator:

### Initiative record

```json
{"kind":"initiative","initiative":{"id":"INIT-PRISMCONTROL-001","title":"PRISM Control MVP","description":"Build the canonical product delivery control plane","status":"in_progress"},"exportedAt":"2026-07-20T10:00:00Z"}
```

### RMI record

```json
{"kind":"rmi","rmi":{"id":"RMI-PRISMCONTROL-005","repo":"github.com/ProductBuildersHQ/prism-build","initiative":"INIT-PRISMCONTROL-001","phase":"INIT-PRISMCONTROL-001/phase-1","title":"Unit-of-work with Dolt commit","type":"capability","status":"in_progress","required":true,"assignedTo":"session-1234"},"exportedAt":"2026-07-20T10:00:00Z"}
```

### Fields

**Initiative fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Initiative ID (e.g. `INIT-PRISMCONTROL-001`) |
| `title` | string | Human-readable title |
| `description` | string | Optional longer description |
| `status` | string | `planned`, `in_progress`, `completed` |

**RMI fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | RMI ID (e.g. `RMI-PRISMCONTROL-005`) |
| `repo` | string | Full repository path |
| `initiative` | string | Parent initiative ID |
| `phase` | string | Phase path (e.g. `INIT-X-001/phase-2`) |
| `title` | string | Human-readable title |
| `type` | string | `capability`, `task`, `bug`, etc. |
| `status` | string | `proposed`, `planned`, `ready`, `in_progress`, `completed` |
| `required` | bool | Whether this RMI is required for phase completion |
| `assignedTo` | string | Current assignee (worker ID or empty) |
| `completedAt` | string | RFC 3339 timestamp when completed |

## DevFolio Datasource

The `datasource/prism` package in devfolio reads the JSONL export:

```go
import "github.com/plexusone/devfolio/datasource/prism"

data, err := prism.LoadFile("export.jsonl")
if err != nil {
    return err
}

// Group RMIs by initiative
byInit := data.RMIsByInitiative()

// Group RMIs by repository
byRepo := data.RMIsByRepo()
```

This enables devfolio to show which initiatives and RMIs a contributor worked on, linking commits (via `Refs:` trailers) to their initiative context.

## Structured-Changelog RMI References

The `structured-changelog` Entry type includes PRISM-aware fields:

```go
import "github.com/grokify/structured-changelog/changelog"

entry := changelog.NewEntry("Add streaming support").
    WithRMI("RMI-MYREPO-042").
    WithInitiative("INIT-X-001").
    WithCommit("abc123").
    WithAuthor("@dev")
```

These fields serialize as:

```json
{
  "description": "Add streaming support",
  "rmi": "RMI-MYREPO-042",
  "initiative": "INIT-X-001",
  "commit": "abc123",
  "author": "@dev"
}
```

When devfolio processes changelogs, the `rmi` and `initiative` fields let it join changelog entries back to PRISM Control's initiative/phase hierarchy.

## End-to-End Pipeline

1. **Export** -- Run `prismctl export --format jsonl -o export.jsonl` to snapshot initiatives and RMIs.
2. **Load** -- DevFolio's `datasource/prism.LoadFile()` parses the JSONL into typed Go structs.
3. **Correlate** -- DevFolio matches contributor commits (carrying `Refs: RMI-*` trailers) against the loaded RMI data to attribute work to initiatives.
4. **Enrich** -- Structured-changelog entries with `rmi` and `initiative` fields provide additional attribution for changelog-driven profiles.
5. **Render** -- The enriched contributor profile includes initiative participation alongside commit, PR, and review statistics.
