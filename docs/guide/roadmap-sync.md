# Roadmap Sync

PRISM Control can sync `ROADMAP.md` files bidirectionally with the database.

## Import

Import a `ROADMAP.md` into the database, creating or updating initiatives, phases, and RMIs:

```bash
prismctl roadmap import --file docs/specs/ROADMAP.md --initiative INIT-X-001
```

This parses the markdown file and creates matching records in the database. Existing records are updated; new ones are created.

## Generate

Generate a `ROADMAP.md` from database state:

```bash
prismctl roadmap generate --initiative INIT-X-001 --output docs/specs/ROADMAP.md
```

This produces a markdown file reflecting the current database state — useful for keeping documentation in sync after programmatic changes.

## Diff

Compare a `ROADMAP.md` file against the database to find drift:

```bash
prismctl roadmap diff --file docs/specs/ROADMAP.md --initiative INIT-X-001
```

Reports discrepancies between the file and database: missing RMIs, status mismatches, title changes.

## Automatic Drift Check

When using `prismctl work ready --initiative INIT-X-001`, the drift check runs automatically. If the initiative's home repo has a `ROADMAP.md` at the conventional path (`docs/specs/ROADMAP.md`), any discrepancies are printed as warnings (up to 5).

This ensures sessions notice drift before starting work — the database is always the source of truth, but the ROADMAP.md should stay in sync for human readability.
