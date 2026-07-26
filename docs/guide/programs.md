# Programs & Initiative Dependencies

## Programs

A program is a first-class entity that groups related initiatives. Programs have their own IDs (`PROG-<SLUG>`), names, descriptions, and organization membership. Some initiatives belong to a program; others are standalone.

### Creating a Program

```bash
prismctl program create \
  --id PROG-PLATFORM \
  --name "Platform Modernization" \
  --org ProductBuildersHQ \
  --description "Cross-cutting platform upgrades"
```

### Managing Programs

```bash
# List all programs
prismctl program list

# Get program detail
prismctl program get PROG-PLATFORM

# Update a program
prismctl program update PROG-PLATFORM --name "Platform Modernization v2"
```

### Assigning Initiatives to Programs

Set the program ID when creating or updating an initiative:

```bash
# On create
prismctl initiative create \
  --id INIT-X-001 \
  --title "Core Platform" \
  --program PROG-PLATFORM

# On update
prismctl initiative update INIT-X-001 --program PROG-PLATFORM
```

### Migrating Free-Text Programs

If you have initiatives with free-text program strings from before the entity promotion, convert them to proper Program entities:

```bash
prismctl program migrate-strings
```

This slugifies existing program name strings into `PROG-<SLUG>` IDs, creates the corresponding Program entities, and updates initiative references.

### Viewing Programs

Programs appear in several places:

- **`program list`** — all programs with ID, name, and organization
- **`initiative list`** — PROGRAM column shows program ID
- **`initiative get`** — Program ID in detail output
- **Dashboard summary** — initiative cards grouped under program headings
- **Dashboard program view** — `/program/<id>` shows all initiatives in a program

## Initiative Dependencies

Initiative dependencies define ordering relationships between initiatives within or across programs.

### Adding Dependencies

```bash
# INIT-X-001 requires INIT-Y-001 to be completed first
prismctl initiative dep add \
  --source INIT-X-001 \
  --target INIT-Y-001 \
  --relationship requires

# INIT-X-001 relates to INIT-Z-001 (informational, no ordering)
prismctl initiative dep add \
  --source INIT-X-001 \
  --target INIT-Z-001 \
  --relationship relates
```

### Listing Dependencies

```bash
# All dependencies
prismctl initiative dep list

# Dependencies for a specific initiative
prismctl initiative dep list INIT-X-001
```

### Relationship Types

| Type | Meaning |
|------|---------|
| `requires` | Source initiative depends on target completing first |
| `relates` | Informational link, no ordering constraint |

### Dashboard Visualization

The dashboard shows initiative dependency edges:

- **Summary page** — all initiative dependencies in a dedicated section
- **Program view** (`/program/<id>`) — dependencies between initiatives in the program
- **Initiative detail** — dependencies involving the specific initiative
