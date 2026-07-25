# Connection Modes

PRISM Control supports two database modes. Choose based on how many processes need concurrent access.

## Embedded Mode (Default)

Single-session access. The full Dolt database engine is compiled into the `prismctl` binary — no external `dolt` install or server process required.

- Data stored at `~/.productbuildershq/prism`
- Exclusive filesystem lock — only one process at a time
- No setup needed; just run `prismctl` commands directly

```bash
prismctl initiative list   # just works
```

**When to use:** Solo development, one `prismctl` at a time. If you only have one Claude Code session working on PRISM-tracked repos, embedded mode is simplest.

## Server Mode

Multi-session access. Start a Dolt SQL server in one terminal, connect from any number of sessions via DSN.

### Starting the Server

```bash
# Terminal 1: start the server (runs in foreground)
prismctl db serve --port 13306
```

The server prints the connection string on startup:

```
Connect with: export PRISMCTL_DSN="root:@tcp(127.0.0.1:13306)/prismcontrol"
```

### Connecting

```bash
# Terminal 2+: set DSN so prismctl uses the server
export PRISMCTL_DSN="root:@tcp(127.0.0.1:13306)/prismcontrol"
prismctl initiative list   # connects to the server
```

When `PRISMCTL_DSN` is set, all `prismctl` commands use the server. When unset, they use embedded mode. Both modes use the same data directory.

**When to use:** Multiple Claude Code sessions (e.g., in separate tmux panes) need concurrent access to the same database. Required when several agent sessions are working on different initiatives simultaneously.

## Workspace Convention

Each initiative has an optional `workspace` field for tracking where work is orchestrated:

```bash
prismctl initiative create \
  --id INIT-X-001 \
  --title "My Initiative" \
  --workspace "my-tmux-session"
```

One tmux session per initiative, with Claude Code sessions in different panes/windows for each repo.
