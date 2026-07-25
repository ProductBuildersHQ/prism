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

## Environment Variables

| Variable | Mode | Description |
|----------|------|-------------|
| *(none)* | Embedded (default) | Uses `~/.productbuildershq/prism` |
| `PRISMCTL_DATA` | Embedded | Override the data directory path |
| `PRISMCTL_DSN` | Server | MySQL-compatible DSN for a running Dolt SQL server |

See [Connection Modes](connection-modes.md) for details on embedded vs. server mode.
