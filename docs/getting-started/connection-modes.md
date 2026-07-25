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

On startup, `db serve` automatically saves the DSN to `~/.productbuildershq/prism/config.json`. All other `prismctl` sessions pick it up immediately — no environment variables needed.

### Connecting

```bash
# Terminal 2+: prismctl reads the DSN from config automatically
prismctl initiative list   # connects to the server
```

### DSN Resolution Order

1. `--dsn` CLI flag
2. `$PRISMCTL_DSN` environment variable
3. Config file (`~/.productbuildershq/prism/config.json`)
4. Embedded mode (default)

### Managing the Config

```bash
prismctl config show                    # display config and resolved DSN
prismctl config set dsn <value>         # set DSN manually
prismctl config unset dsn               # clear DSN (revert to embedded)
prismctl config path                    # print config file location
```

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
