# VisionStudio Integration

VisionStudio connects to PRISM Control's Dolt database as a read-only SQL
consumer. This document explains how to set up the connection and what data
is available.

## Prerequisites

- `prismctl` built and on your `$PATH`
- A PRISM Control database initialized (`prismctl db init`)

## Starting the Database

Start the Dolt SQL server so VisionStudio's Go daemon can connect over the
MySQL wire protocol:

```bash
prismctl db serve --port 3306
```

This binds to `127.0.0.1:3306`. The default database name is `prismcontrol`.

## Creating the Views

Run this once (and after schema upgrades) to create the read-only views:

```bash
prismctl db create-views
```

This executes the view definitions from `docs/sql/visionstudio-views.sql`
against the running database.

## Connecting VisionStudio

In VisionStudio's daemon configuration, add a MySQL-compatible datasource
pointing at the local Dolt server:

| Parameter | Value |
|-----------|-------|
| Host | `127.0.0.1` |
| Port | `3306` |
| Database | `prismcontrol` |
| User | `root` |
| Password | (empty) |

The DSN string for a Go `database/sql` connection:

```text
root:@tcp(127.0.0.1:3306)/prismcontrol
```

All queries should be read-only. VisionStudio must never issue DDL or DML
against PRISM Control tables; all writes go through `prismctl` or the MCP
server.

## Available Views

### v_initiative_summary

Initiative-level rollup with progress metrics.

| Column | Type | Description |
|--------|------|-------------|
| initiative_id | VARCHAR(64) | Initiative primary key |
| title | VARCHAR(255) | Initiative title |
| status | VARCHAR(32) | Lifecycle status (proposed, planned, executing, ...) |
| total_rmis | INT | Total RMIs in this initiative |
| completed_rmis | INT | RMIs with status = completed |
| required_completed | INT | Required RMIs with status = completed |
| phase_count | INT | Number of phases |
| repos_involved | INT | Distinct repositories with RMIs in this initiative |

### v_phase_progress

Per-phase progress with derived status following the TRD status vocabulary.

| Column | Type | Description |
|--------|------|-------------|
| phase_id | VARCHAR(64) | Phase primary key |
| initiative_id | VARCHAR(64) | Parent initiative |
| title | VARCHAR(255) | Phase title |
| theme | VARCHAR(255) | Grouping rationale |
| sequence_number | INT | Ordering within the initiative |
| total_rmis | INT | RMIs in this phase |
| completed_count | INT | Completed RMIs |
| derived_status | VARCHAR(32) | One of: completed, partial, in_progress, blocked, planned |

Derived status rules:

- **completed** -- all required RMIs completed, no optional remain
- **partial** -- all required RMIs completed, optional RMIs remain
- **in_progress** -- at least one RMI is in_progress
- **blocked** -- at least one required RMI is blocked
- **planned** -- no RMIs started (or phase has no RMIs)

### v_rmi_detail

Flat RMI detail with denormalized foreign key references.

| Column | Type | Description |
|--------|------|-------------|
| rmi_id | VARCHAR(64) | RMI primary key |
| title | VARCHAR(255) | RMI title |
| status | VARCHAR(32) | planned, ready, in_progress, blocked, completed, cancelled |
| item_type | VARCHAR(32) | capability, task, spec, release |
| required | BOOLEAN | Whether this RMI is required for phase completion |
| repository_id | VARCHAR(128) | Repository this RMI belongs to |
| initiative_id | VARCHAR(64) | Parent initiative (nullable) |
| phase_id | VARCHAR(64) | Parent phase (nullable) |
| created_at | DATETIME | When the RMI was created |
| completed_at | DATETIME | When the RMI was completed (nullable) |

### v_active_assignments

Currently active work assignments with the related RMI title.

| Column | Type | Description |
|--------|------|-------------|
| assignment_id | VARCHAR(64) | Assignment primary key |
| rmi_id | VARCHAR(64) | Claimed RMI |
| rmi_title | VARCHAR(255) | Title of the claimed RMI |
| worker | VARCHAR(128) | Session ID of the worker |
| lease_expires_at | DATETIME | When the lease expires |
| workspace | VARCHAR(512) | Local path or worktree (nullable) |

## Example Queries

Once connected, VisionStudio can query the views directly:

```sql
-- Dashboard: all initiatives with progress
SELECT * FROM v_initiative_summary ORDER BY initiative_id;

-- Phase drill-down for a specific initiative
SELECT * FROM v_phase_progress
WHERE initiative_id = 'INIT-PRISMCONTROL-001'
ORDER BY sequence_number;

-- All RMIs that are blocked
SELECT * FROM v_rmi_detail WHERE status = 'blocked';

-- Who is working on what right now
SELECT * FROM v_active_assignments;
```

## Portability

These views use only base tables and standard SQL. They do not depend on
Dolt system tables (`dolt_history_*`, `dolt_diff_*`, `AS OF`) per the TRD
portability rule (NFR4). This means VisionStudio could connect to any
MySQL-compatible database that hosts the same schema.
