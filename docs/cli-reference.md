# CLI Reference

All commands accept global flags:

| Flag | Description |
|------|-------------|
| `--data-dir` | Data directory for embedded Dolt (default: `$PRISMCTL_DATA` or `~/.productbuildershq/prism`) |
| `--dsn` | MySQL-compatible DSN for server mode (resolution: flag > `$PRISMCTL_DSN` > config file > embedded default) |

## Config

| Command | Description |
|---------|-------------|
| `config show` | Display current configuration and resolved DSN |
| `config set <key> <value>` | Set a config value (supported: `dsn`) |
| `config unset <key>` | Remove a config value |
| `config path` | Print the config file path |

## Database

| Command | Description |
|---------|-------------|
| `db init` | Initialize database and run schema migration |
| `db serve [--port N]` | Start a Dolt SQL server in the foreground |
| `db start [--port N]` | Start the Dolt SQL server in the background (default port 13306) |
| `db stop` | Stop the background Dolt SQL server |
| `db restart [--port N]` | Restart the background Dolt SQL server |
| `db status` | Check server status (PID, port, DSN) |
| `db create-views` | Create read-only SQL views for VisionStudio and other consumers |

## Dashboard

| Command | Description |
|---------|-------------|
| `dashboard [--port N]` | Start the two-level web dashboard (default port 9400) |
| `dashboard --static` | Write a one-shot HTML snapshot |

## Registry

| Command | Description |
|---------|-------------|
| `registry add --org O --name N [--path P]` | Register a repository |
| `registry list` | List all registered repositories |
| `registry scan --org-dir DIR` | Auto-discover and register repos in a directory |
| `registry deps` | Show topological dependency order |
| `registry unpushed` | List repos with uncommitted or unpushed work |

## Programs

| Command | Description |
|---------|-------------|
| `program create --id ID --name N --org O [--description D]` | Create a program |
| `program list` | List all programs |
| `program get <id>` | Show program detail |
| `program update <id> [--name N] [--description D]` | Update a program |
| `program migrate-strings` | Convert free-text program values to Program entities |

## Initiatives

| Command | Description |
|---------|-------------|
| `initiative create --id ID --title T [--org O] [--program PROG-ID] [--priority P] [--home-repo R] [--workspace W] [--spec K=V]` | Create an initiative |
| `initiative list` | List all initiatives |
| `initiative get <id>` | Show initiative detail with phase status |
| `initiative update <id> [--workspace W] [--home-repo R] [--program PROG-ID] [--description D] [--priority P]` | Update initiative fields |
| `initiative transition <id> <status>` | Transition initiative lifecycle status |
| `initiative dep add --source S --target T [--relationship R]` | Add initiative dependency |
| `initiative dep list [id]` | List initiative dependencies |

## Phases

| Command | Description |
|---------|-------------|
| `phase add --initiative I --seq N --title T [--theme TH]` | Add a phase |
| `phase list <initiative-id>` | List phases with derived status |

## Roadmap Items (RMIs)

| Command | Description |
|---------|-------------|
| `rmi create --id ID --repo R --initiative I --phase P --title T [--type TY] [--required]` | Create an RMI |
| `rmi get <id>` | Show RMI detail with dependencies |
| `rmi list --initiative I` | List RMIs for an initiative |
| `rmi update <id> [--status S] [--title T]` | Update an RMI |
| `rmi update-phase <phase-id> --status S [--from F]` | Bulk-update status of all RMIs in a phase |
| `rmi dep add --source S --target T [--relationship R]` | Add RMI dependency |
| `rmi dep list [id]` | List RMI dependencies |

## Work Assignments

| Command | Description |
|---------|-------------|
| `work ready [--repo R] [--initiative I]` | List ready, unblocked, unclaimed RMIs |
| `work claim <rmi-id> [--worker W] --lease-hours H` | Claim an RMI with a lease |
| `work claim-phase <phase-id> [--ready] --lease-hours H` | Claim all ready RMIs in a phase |
| `work renew <rmi-or-assign-id> --lease-hours H` | Extend a lease |
| `work release <rmi-or-assign-id> [--handoff JSON]` | Release a claim |
| `work update <rmi-or-assign-id> [--handoff JSON]` | Update handoff state |
| `work complete <rmi-or-assign-id>... [--transition]` | Mark work as completed |
| `work complete-phase <phase-id> [--transition]` | Complete all in-progress RMIs in a phase |
| `work status` | List all active assignments |

## Roadmap Sync

| Command | Description |
|---------|-------------|
| `roadmap import --file F --initiative I` | Import ROADMAP.md into the database |
| `roadmap generate --initiative I --output F` | Generate ROADMAP.md from database |
| `roadmap diff --file F --initiative I` | Compare ROADMAP.md against database |

## Ingestion

| Command | Description |
|---------|-------------|
| `ingest git <repo-id>` | Scan commits for `Refs:` trailers |
| `ingest changelog <repo-id>` | Import structured-changelog entries |

## Reports & Validation

| Command | Description |
|---------|-------------|
| `report initiative <id> [--format json\|markdown]` | End-to-end initiative report |
| `release plan <initiative-id>` | Dependency-ordered release plan |
| `validate [--format text\|json]` | Consistency checks across the store |
| `export [--dir D]` | JSONL snapshots of all tables |

## MCP Server

| Command | Description |
|---------|-------------|
| `mcp` | Start the MCP server (stdio transport) |
