# Deploy MCP

Deploy the pgEdge MCP server onto an existing active database.

## Prerequisites

- An active database (`pgecloudctl databases get <db-id> -o json` returns
  `"status": "active"`)
- The database's parent cluster ID
- Embedding provider credentials (OpenAI API key, or equivalent for other
  providers)

## Decision Points

- **Public vs private cluster** — if `node_location` is `private`, an ingress
  and service registration step is required after deployment
- **Write access** — pass `--allow-writes` only if agents need INSERT/UPDATE/
  DELETE permissions; omit for read-only workloads
- **Target nodes** — omit `--target-nodes` to deploy on all nodes; specify a
  subset for cost control or staged rollout

## Steps

### Step 1: Verify database status

```bash
pgecloudctl databases get <db-id> -o json
```

Expected output: JSON object with `"status": "active"`.
Capture `cluster_id` as `<cluster-id>`. Abort if status is not `active`.

### Step 2: Check cluster node location

```bash
pgecloudctl clusters get <cluster-id> -o json
```

Capture `node_location` (`public` or `private`). This determines whether
ingress steps are needed.

### Step 3: Deploy MCP

> **Security note:** Use environment variables for API keys to avoid
> exposing them in shell history.

```bash
pgecloudctl databases mcp deploy <db-id> \
  --embedding-provider <provider> \
  --embedding-model <model> \
  --embedding-api-key "$OPENAI_API_KEY" \
  -o json
```

Common provider/model combinations:

| Provider | Model                      |
|----------|----------------------------|
| openai   | text-embedding-3-small     |
| openai   | text-embedding-3-large     |

Optional flags:

| Flag               | Purpose                                          |
|--------------------|--------------------------------------------------|
| `--allow-writes`   | Permit INSERT/UPDATE/DELETE via the MCP server   |
| `--target-nodes`   | Comma-separated list of node IDs to deploy on    |
| `--init-tokens`    | Number of initial API tokens to create           |
| `--init-users`     | Number of initial database users to create       |
| `--ollama-url`     | Ollama endpoint URL (for self-hosted embeddings) |

Capture `task_id` from the response as `<mcp-task-id>`.

### Step 4: Wait for deployment

```bash
pgecloudctl tasks wait <mcp-task-id> -o json --timeout 600
```

Expected output: JSON with `"status": "completed"`. Abort if status is
`failed` — read the `error` field for detail.

### Step 5 (private clusters only): Expose the MCP service via ingress

Skip to Step 6 if `node_location` is `public`.

List services to find the MCP service ID:

```bash
pgecloudctl databases services list <db-id> -o json
```

Identify the MCP service entry and capture its `id` as `<service-id>`.

Create an ingress (skip if one already exists for this cluster):

```bash
pgecloudctl ingresses create \
  --name <ingress-name> \
  --cluster-id <cluster-id> \
  --region <region> \
  -o json
```

Capture `id` as `<ingress-id>`.

Register the service:

```bash
pgecloudctl ingresses services register <ingress-id> \
  --database-id <db-id> \
  --service-id <service-id>
```

### Step 6: Verify

```bash
pgecloudctl databases services list <db-id>
```

Expected output: an MCP service entry with `"status": "active"`.

## Verification

- `pgecloudctl databases services list <db-id> -o json` — MCP service present
  with `"status": "active"`
- If ingress was registered: `pgecloudctl ingresses services list <ingress-id>`
  — service appears with `"status": "registered"`

## Error Handling

| Exit Code | Meaning | Recovery |
|-----------|---------|----------|
| 0 | Success | Continue to next step |
| 1 | General error (invalid flags, API errors, constraints) | Check command output for details; verify flags and resource state |
| 2 | Authentication failure | Run `pgecloudctl auth login` |
| 3 | Request timeout | Retry the command; check network connectivity |
| 4 | Resource not found (database/cluster) | Verify `<db-id>` with `pgecloudctl databases list` |
