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

### Script

    curl -fsSL https://raw.githubusercontent.com/AntTheLimey/pgecloudctl/main/install.sh | sh

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
| `cloud-accounts cloudformation-template` | Get AWS IAM template |
| `backups list` | List backups |
| `backups get <id>` | Get backup details |
| `backups create` | Create a backup |
| `backups delete <id>` | Delete a backup |
| `backups url <id>` | Get backup download URL |
| `backup-stores list` | List backup stores |
| `backup-stores get <id>` | Get backup store details |
| `backup-stores create` | Create a backup store |
| `backup-stores delete <id>` | Delete a backup store |
| `ssh-keys list` | List SSH keys |
| `ssh-keys get <id>` | Get SSH key details |
| `ssh-keys create` | Create an SSH key |
| `ssh-keys delete <id>` | Delete an SSH key |
| `ingresses list` | List ingresses |
| `ingresses get <id>` | Get ingress details |
| `ingresses create` | Create an ingress |
| `ingresses delete <id>` | Delete an ingress |
| `ingresses services list <id>` | List services on an ingress |
| `ingresses services register <id>` | Register a service |
| `ingresses services deregister <id> <svc-id>` | Deregister a service |
| `invites list` | List invites |
| `invites create` | Create an invite |
| `invites accept <id>` | Accept an invite |
| `invites delete <id>` | Delete an invite |
| `memberships list` | List team members |
| `memberships delete <id>` | Remove a team member |
| `clusters shares list <id>` | List cluster shares |
| `clusters shares create <id>` | Create a cluster share |
| `clusters shares get <id> <share-id>` | Get share details |
| `clusters shares delete <id> <share-id>` | Delete a share |
| `version` | Print version |

## Global Flags

| Flag | Description |
|------|-------------|
| `--api-url` | API base URL (default: api.pgedge.com) |
| `-o, --output` | Output format: table, json, yaml (default: table) |
| `--no-color` | Disable color output |
| `-v, --verbose` | Show HTTP request/response details |

## License

PostgreSQL License
