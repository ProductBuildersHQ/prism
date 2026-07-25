# Programs & Initiative Dependencies

## Programs

A program is an optional grouping of related initiatives. Some initiatives are part of a program; others are standalone.

### Assigning a Program

Set the program when creating or updating an initiative:

```bash
# On create
prismctl initiative create \
  --id INIT-X-001 \
  --title "Core Platform" \
  --program "Platform Modernization"

# On update
prismctl initiative update INIT-X-001 --program "Platform Modernization"
```

### Viewing Programs

Programs appear in several places:

- **`initiative list`** — PROGRAM column in the table
- **`initiative get`** — Program field in detail output
- **Dashboard summary** — initiative cards grouped under program headings
- **Dashboard program view** — `/program/<name>` shows all initiatives in a program

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
- **Program view** — dependencies between initiatives in the program
- **Initiative detail** — dependencies involving the specific initiative
