# Installation

## Prerequisites

- Go 1.22+ with cgo enabled
- ICU library (for embedded Dolt's regex engine)

## macOS (Homebrew)

```bash
brew install icu4c
```

Set the ICU paths for cgo:

```bash
export CGO_CPPFLAGS="-I$(brew --prefix icu4c)/include"
export CGO_LDFLAGS="-L$(brew --prefix icu4c)/lib"
```

!!! tip
    Add these exports to your shell profile (`.zshrc`, `.bashrc`) or `.envrc` so they persist across sessions.

## Build

```bash
go build -o prismctl ./cmd/prismctl/
```

Or install directly:

```bash
go install ./cmd/prismctl/
```

## Initialize the Database

```bash
prismctl db init
```

This creates `~/.productbuildershq/prism/`, initializes the embedded Dolt database, and runs schema migration. No environment variables needed — embedded mode is the default.

To use a custom location:

```bash
PRISMCTL_DATA=/path/to/data prismctl db init
```

## Verify

```bash
prismctl initiative list
prismctl validate
```

## Configuration

prismctl stores configuration at `~/.productbuildershq/prism/config.json`. The `db start` and `db serve` commands write the DSN automatically; you can also manage it manually:

```bash
prismctl config show                    # display config and resolved DSN
prismctl config set dsn <value>         # set DSN manually
prismctl config unset dsn               # clear DSN (revert to embedded)
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `PRISMCTL_DATA` | Override the embedded data directory (default: `~/.productbuildershq/prism`) |
| `PRISMCTL_DSN` | Override the DSN (takes precedence over config file) |

DSN resolution order: `--dsn` flag > `$PRISMCTL_DSN` > config file > embedded default.

See [Connection Modes](connection-modes.md) for details on embedded vs. server mode.
