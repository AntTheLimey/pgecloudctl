# Full-Stack Setup

End-to-end provisioning of a pgEdge Cloud cluster, database, and optional
AI services from a clean slate.

## Prerequisites

- `pgecloudctl` authenticated (`pgecloudctl auth status` returns non-error)
- A cloud account ID (retrieve with `pgecloudctl cloud-accounts list -o json`)
- Target region string (e.g. `us-west-2`, `eastus`)
- For MCP/RAG: OpenAI API key (or other embedding provider credentials)
- For RAG: a `pipeline.json` file on disk

## Decision Points

- **Public vs private cluster** — `--node-location public` skips ingress steps;
  `--node-location private` requires creating an ingress and registering the
  service before connections are possible
- **MCP, RAG, or neither** — determines which optional service-deployment steps
  to run
- **Backup store exists** — if `backup-stores list` returns results, skip
  creation; otherwise create one before the cluster

## Steps

### Step 1: Verify authentication

```bash
pgecloudctl auth status -o json
```

Expected output: JSON with `"status": "authenticated"`.
Abort if the command returns a non-zero exit code.

### Step 2: Check for an existing backup store

```bash
pgecloudctl backup-stores list -o json
```

If the array is non-empty, capture `id` from the first result and skip to
Step 4.

### Step 3: Create a backup store (if needed)

```bash
pgecloudctl backup-stores create \
  --name <backup-store-name> \
  --cloud-account-id <cloud-account-id> \
  --region <region> \
  -o json
```

Capture `id` from the response as `<backup-store-id>`.

### Step 4: Create the cluster

```bash
pgecloudctl clusters create \
  --name <cluster-name> \
  --cloud-account-id <cloud-account-id> \
  --regions <region> \
  --node-location <public|private> \
  -o json
```

Capture `task_id` from the response as `<cluster-task-id>`.

### Step 5: Wait for cluster provisioning

```bash
pgecloudctl tasks wait <cluster-task-id> -o json --timeout 600
```

Expected output: JSON with `"status": "completed"`.
Capture `result.id` (or `resource_id`) as `<cluster-id>`.

### Step 6: Configure firewall rules

Firewall rule management is not yet available in `pgecloudctl`. Open the
pgEdge Cloud UI at `https://app.pgedge.com`, navigate to the cluster, and
configure allowed CIDR ranges under **Network → Firewall**.

### Step 7: Create the database

```bash
pgecloudctl databases create \
  --name <database-name> \
  --cluster-id <cluster-id> \
  -o json
```

Capture `task_id` as `<database-task-id>`.

### Step 8: Wait for database provisioning

```bash
pgecloudctl tasks wait <database-task-id> -o json --timeout 600
```

Expected output: JSON with `"status": "completed"`.
Capture `result.id` (or `resource_id`) as `<db-id>`.

### Step 9 (optional): Deploy MCP

Skip this step if MCP is not required.

> **Security note:** Use environment variables for API keys to avoid
> exposing them in shell history.

```bash
pgecloudctl databases mcp deploy <db-id> \
  --embedding-provider openai \
  --embedding-model text-embedding-3-small \
  --embedding-api-key "$OPENAI_API_KEY" \
  -o json
```

Additional optional flags: `--allow-writes`, `--target-nodes <nodes>`,
`--init-tokens <n>`, `--init-users <n>`.

Capture `task_id` and run:

```bash
pgecloudctl tasks wait <mcp-task-id> -o json --timeout 600
```

### Step 10 (optional): Deploy RAG

Skip this step if RAG is not required. Requires `pipeline.json` on disk (see
the [deploy-rag workflow](deploy-rag.md) for a sample config).

```bash
pgecloudctl databases rag deploy <db-id> \
  --embedding-llm-provider openai \
  --embedding-llm-model text-embedding-3-small \
  --embedding-llm-api-key "$OPENAI_API_KEY" \
  --completion-llm-provider openai \
  --completion-llm-model gpt-4o \
  --completion-llm-api-key "$OPENAI_API_KEY" \
  --pipeline-config pipeline.json \
  -o json
```

Additional optional flags: `--token-budget <n>`, `--top-n <n>`,
`--target-nodes <nodes>`.

Capture `task_id` and run:

```bash
pgecloudctl tasks wait <rag-task-id> -o json --timeout 600
```

### Step 11 (private clusters only): Expose the service via ingress

Skip to Step 12 if `--node-location public` was used in Step 4.

List services to get the service ID:

```bash
pgecloudctl databases services list <db-id> -o json
```

Capture `id` from the relevant service as `<service-id>`.

Create an ingress:

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

### Step 12: Verify

```bash
pgecloudctl databases services list <db-id>
```

Expected output: at least one service with `status: active`.

## Verification

- `pgecloudctl clusters get <cluster-id> -o json` → `"status": "active"`
- `pgecloudctl databases get <db-id> -o json` → `"status": "active"`
- If ingress was created: `pgecloudctl ingresses services list <ingress-id>`
  → service appears with `"status": "registered"`

## Error Handling

| Exit Code | Meaning | Recovery |
|-----------|---------|----------|
| 0 | Success | Continue to next step |
| 1 | General error (invalid flags, API errors, constraints) | Check command output for details; verify flags and resource state |
| 2 | Authentication failure | Run `pgecloudctl auth login` |
| 3 | Request timeout | Retry the command; check network connectivity |
| 4 | Resource not found | Verify IDs with the relevant `list` command |
