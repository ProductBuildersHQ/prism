# PRISM Control

**Product Delivery Control Plane** — the canonical, queryable store for cross-repository initiatives, per-repo Roadmap Items (RMIs), lease-based work assignments, and delivery evidence.

PRISM Control is headless and library-first. It coordinates work across repositories and agent sessions; it does not execute work, render UIs, or compute metrics. Those responsibilities belong to:

- **Execution layer** — Claude Code sessions, Beads
- **Metrics layer** — omnidevx, devfolio
- **Visualization layer** — VisionStudio, the built-in dashboard

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Initiative** | A cross-repository project tracked through a lifecycle (proposed → planned → executing → delivery_complete → released → closed) |
| **Program** | A first-class entity (`PROG-<SLUG>`) grouping related initiatives |
| **Phase** | A themed grouping of RMIs within an initiative; status derived from member RMIs |
| **RMI** | A Roadmap Item — a single deliverable within one repository |
| **Assignment** | A lease-based work claim by an agent session on an RMI |
| **Evidence** | A commit, PR, release, or changelog entry linked to an RMI via `Refs:` trailers |

## How It Works

```
┌────────────────────────────────────────────────────────┐
│                    prism-control (Dolt)                │
│  initiatives → phases → RMIs → assignments → evidence  │
└────────────┬──────────────┬─────────────┬──────────────┘
             │              │             │
     ┌───────▼──────┐ ┌─────▼────┐ ┌──────▼──────┐
     │  prismctl    │ │ MCP      │ │ SQL views   │
     │  CLI         │ │ server   │ │ (read-only) │
     └───────┬──────┘ └─────┬────┘ └──────┬──────┘
             │              │             │
     Agent sessions    Agent sessions  VisionStudio
     (Claude Code)     (Claude Code)   dashboards
```

1. **Plan** — Find claimable work with `prismctl work ready`
2. **Claim** — Take ownership with a lease (`work claim` / `work claim-phase`)
3. **Execute** — Commit with `Refs: RMI-<REPOSLUG>-<NNN>` trailers
4. **Update** — Record evidence and hand off (`work complete` / `work release`)

## Quick Start

```bash
# Build (macOS — adjust ICU paths for your platform)
export CGO_CPPFLAGS="-I$(brew --prefix icu4c)/include"
export CGO_LDFLAGS="-L$(brew --prefix icu4c)/lib"
go build -o prismctl ./cmd/prismctl/

# Initialize
prismctl db init

# Verify
prismctl initiative list
prismctl validate

# Open the dashboard
prismctl dashboard
```

See [Installation](getting-started/install.md) for detailed setup instructions.

## Stack

- [Dolt](https://github.com/dolthub/dolt) — MySQL-compatible database with git-like version control
- [Ent](https://entgo.io/ent) — Go ORM (MySQL dialect)
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) — MCP server
- [gogit](https://github.com/grokify/gogit) — Git log parsing for commit ingestion
