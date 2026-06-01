# pgEdge pgecloudctl

CLI for managing pgEdge Cloud resources via the REST API.

## Quick Start

```bash
# Install
brew install AntTheLimey/tap/pgecloudctl

# Authenticate
pgecloudctl auth login

# List clusters
pgecloudctl clusters list

# JSON output for scripting / AI agents
pgecloudctl clusters list -o json
```

## Features

- Manage clusters, databases, cloud accounts, and tasks
- Deploy and configure MCP and RAG services on databases
- Table and JSON output formats
- Async task polling with `tasks wait`
- Environment variable and config file authentication
