[![CI](https://github.com/AntTheLimey/pgecloudctl/actions/workflows/ci.yml/badge.svg)](https://github.com/AntTheLimey/pgecloudctl/actions/workflows/ci.yml)

# pgecloudctl

CLI for managing [pgEdge Cloud](https://www.pgedge.com) resources.

## Installation

### Homebrew

    brew install AntTheLimey/tap/pgecloudctl

### Go

    go install github.com/AntTheLimey/pgecloudctl@latest

### Binary

Download from
[Releases](https://github.com/AntTheLimey/pgecloudctl/releases).

## Quick Start

```bash
# Authenticate
pgecloudctl auth login

# List clusters
pgecloudctl clusters list

# Get JSON output for scripting / AI agents
pgecloudctl clusters list -o json
```

## Authentication

Create an API client in the pgEdge Cloud UI under Settings > Client.
Then either:

```bash
# Interactive login (stores credentials locally)
pgecloudctl auth login

# Or use environment variables (CI/automation)
export PGEDGE_CLIENT_ID="your-client-id"
export PGEDGE_CLIENT_SECRET="your-client-secret"
```

## Commands

| Command | Description |
|---------|-------------|
| `auth login` | Authenticate with pgEdge Cloud |
| `auth status` | Show current auth state |
| `auth logout` | Clear stored credentials |
| `clusters list` | List clusters |
| `clusters get <id>` | Get cluster details |
| `clusters create` | Create a cluster |
| `clusters delete <id>` | Delete a cluster |
| `databases list` | List databases |
| `databases get <id>` | Get database details |
| `databases create` | Create a database |
| `databases update <id>` | Update a database |
| `databases delete <id>` | Delete a database |
| `databases services list <db-id>` | List services on a database |
| `databases mcp deploy <db-id>` | Deploy MCP server |
| `databases rag deploy <db-id>` | Deploy RAG server |
| `tasks list` | List tasks |
| `tasks get <id>` | Get task status |
| `tasks wait <id>` | Wait for task completion |
| `cloud-accounts list` | List cloud accounts |
| `cloud-accounts create` | Create a cloud account |
| `version` | Print version |

## Global Flags

| Flag | Description |
|------|-------------|
| `--api-url` | API base URL (default: api.pgedge.com) |
| `-o, --output` | Output format: table, json (default: table) |
| `--no-color` | Disable color output |
| `-v, --verbose` | Show HTTP request/response details |

## License

PostgreSQL License
