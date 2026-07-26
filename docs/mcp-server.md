# MCP Server

PRISM Control includes an MCP (Model Context Protocol) server for agent integration. It exposes the same service layer as the CLI over stdio transport.

## Setup

Register in your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "prism-control": {
      "command": "prismctl",
      "args": ["mcp"],
      "env": {}
    }
  }
}
```

For server mode (concurrent access):

```json
{
  "mcpServers": {
    "prism-control": {
      "command": "prismctl",
      "args": ["mcp", "--dsn", "root:@tcp(127.0.0.1:13306)/prismcontrol"],
      "env": {}
    }
  }
}
```

## Available Tools

| Tool | Description |
|------|-------------|
| `program_list` | List all programs |
| `program_create` | Create a new program |
| `initiative_list` | List all initiatives |
| `initiative_get` | Get initiative detail with phase status |
| `initiative_create` | Create a new initiative (optional `program_id`) |
| `rmi_create` | Create a new roadmap item |
| `work_ready` | Find ready, unblocked, unclaimed work |
| `task_claim` | Claim an RMI with a lease |
| `task_release` | Release a work claim |
| `task_update` | Update handoff state on an assignment |
| `report_initiative` | Generate a full initiative report |

## Usage in Claude Code

Once registered, Claude Code sessions can use PRISM Control tools directly:

```
Use the prism-control MCP to find claimable work:
→ work_ready(repo: "github.com/org/myrepo")

Claim an RMI:
→ task_claim(rmi_id: "RMI-MYREPO-042", lease_hours: 4)

Complete work:
→ task_update(assignment_id: "assign-...", status: "completed")
```

The MCP server shares the same service layer as the CLI, so all operations produce identical results regardless of which interface is used.
