# Product Repo Setup

Repos participating in PRISM-tracked initiatives need two things: a registry entry and a `CLAUDE.md` pointer.

## Register the Repository

```bash
prismctl registry add \
  --org plexusone \
  --name myrepo \
  --path ~/go/src/github.com/plexusone/myrepo
```

This creates a catalog entry so `prismctl` can resolve repo IDs, scan for commits, and detect local paths.

To discover and register repos automatically:

```bash
prismctl registry scan --org-dir ~/go/src/github.com/plexusone
```

## Add the CLAUDE.md Pointer

Add a section to the repo's `CLAUDE.md` so Claude Code sessions know how to find claimable work:

```markdown
## PRISM Control

This repo's roadmap items are tracked in [prism-control](https://github.com/ProductBuildersHQ/prism-control).
Use `prismctl work ready --repo github.com/org/<this-repo>` to find claimable work,
and carry the `Refs: RMI-<REPOSLUG>-<NNN>` trailer on every commit.
```

Replace `<REPOSLUG>` with the uppercase repo slug used in RMI IDs (e.g., `MARKETSPEC`, `OMNISIGNAL`, `PRISMCONTROL`).

## Verify

```bash
# Check the repo is registered
prismctl registry list | grep myrepo

# Check for claimable work
prismctl work ready --repo github.com/org/myrepo
```

## Repository Utilities

```bash
# Show dependency order across all registered repos
prismctl registry deps

# Find repos with uncommitted or unpushed work
prismctl registry unpushed
```
