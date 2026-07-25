# Commit Attribution

PRISM Control uses git trailers to link commits to roadmap items. This creates an auditable chain from code changes back to the initiative that motivated them.

## The `Refs:` Trailer

Every commit implementing a roadmap item carries a git trailer:

```
feat(store): add unit-of-work with Dolt commit

Refs: RMI-PRISMCONTROL-005
```

### Rules

- **Git trailer** with key `Refs:` — not in the subject line
- **RMI only** — never the initiative ID; the RMI→initiative edge is resolved at report time
- **One RMI per commit** (comma-separated if a commit genuinely serves several)
- **Conventional Commits** format for the message itself

### Fallbacks

For squash-merge workflows where trailers may be lost:

| Priority | Method | Example |
|----------|--------|---------|
| 1 | Git trailer | `Refs: RMI-MYREPO-042` |
| 2 | PR body reference | "Implements RMI-MYREPO-042" |
| 3 | Branch name | `rmi/myrepo-042-add-widget` |

## Evidence Ingestion

The `ingest git` command scans commit history for `Refs:` trailers and creates evidence rows:

```bash
prismctl ingest git github.com/org/myrepo
```

This finds all commits with `Refs: RMI-*` trailers since the last scan (tracked via `ingest_high_water` on the repository record) and creates `DeliveryEvidence` rows linking each commit to its RMI.

## Changelog Ingestion

Structured-changelog entries can also carry RMI references:

```bash
prismctl ingest changelog github.com/org/myrepo
```

This reads `CHANGELOG.json` and creates evidence rows for entries with `rmi` fields.

## Report Generation

Reports aggregate evidence by initiative:

```bash
prismctl report initiative INIT-X-001
```

The report traces: initiative → phases → RMIs → assignments → evidence (commits, PRs, changelog entries), producing a complete delivery narrative.
