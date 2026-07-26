# Dashboard

PRISM Control includes a two-level web dashboard for visualizing initiatives, phases, and RMIs.

## Starting the Dashboard

```bash
# Live server (default) — re-queries DB on every page load
prismctl dashboard

# Custom port
prismctl dashboard --port 8080

# Static HTML file (one-shot snapshot, summary only)
prismctl dashboard --static
```

The server auto-refreshes every 5 seconds and opens your browser automatically.

## Summary Page

The landing page (`/`) shows:

- **Stats bar** — total initiatives, programs, phases, RMIs, and completion count
- **RMI status distribution** — bar chart showing proposed, ready, in-progress, completed counts
- **Initiative cards by program** — grouped under program headings with progress bars
- **Standalone initiatives** — initiatives not assigned to any program
- **Initiative dependency edges** — if any initiative dependencies are defined

Each initiative card displays:

- Title and status badge
- Initiative ID
- Progress bar with completion percentage
- Priority, home repo, workspace
- Repository pills showing RMI count per repo

## Initiative Detail

Click an initiative card to drill down (`/initiative/<id>`):

- Breadcrumb navigation back to summary (and program if applicable)
- Full initiative header with all metadata
- Repository pills
- Overall progress bar
- Collapsible phases with status badges per phase
- Individual RMI rows with status, type, repo, and required flag
- Tooltips showing claim/completion timestamps
- Initiative and RMI dependency edges

## Program View

Click a program name to see all initiatives in that program (`/program/<id>`):

- Program ID and name
- Program-level stats (initiatives, phases, RMIs, completion)
- Initiative dependency graph within the program
- Full phase/RMI detail for every initiative, with collapsible phases

## Keyboard Shortcuts

- Click a phase header to collapse/expand its RMI list
- Click the toggle-all arrow (top-left of the RMI header row) to collapse/expand all phases in an initiative
- Collapse state is preserved in localStorage across page refreshes
