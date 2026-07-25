# Data Backup

Dolt databases support git-style remotes, making backup straightforward.

## Push to a Git Remote

```bash
cd ~/.productbuildershq/prism/prismcontrol
dolt remote add backup git@github.com:org/prism-backup.git
dolt push backup main
```

This pushes the full database (schema, data, commit history) to GitHub. The remote repository contains the complete Dolt database — you can clone it on another machine to restore.

## Automated Backup

Add to cron for daily backups:

```bash
0 2 * * * cd ~/.productbuildershq/prism/prismcontrol && dolt push backup main
```

## Restore

Clone from the remote to restore on a new machine:

```bash
mkdir -p ~/.productbuildershq/prism
cd ~/.productbuildershq/prism
dolt clone git@github.com:org/prism-backup.git prismcontrol
```

Then run `prismctl db init` to ensure the schema is up to date.

## JSONL Export

For portable snapshots that don't require Dolt:

```bash
prismctl export --dir ./exports
```

This writes one JSONL file per table plus a `manifest.json` with record counts and timestamp.
