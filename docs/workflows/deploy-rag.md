# Deploy RAG

Deploy the pgEdge RAG (Retrieval-Augmented Generation) server onto an existing
active database.

## Prerequisites

- An active database (`pgecloudctl databases get <db-id> -o json` returns
  `"status": "active"`)
- The database's parent cluster ID
- Embedding and completion LLM provider credentials
- A `pipeline.json` configuration file on disk (see sample below)

## Decision Points

- **Public vs private cluster** — if `node_location` is `private`, an ingress
  and service registration step is required after deployment
- **Pipeline configuration** — `pipeline.json` defines which tables, columns,
  and retrieval strategy to use; at least one pipeline entry is required
- **Target nodes** — on single-node clusters, `--target-nodes` is optional
  (the node is auto-selected); on multi-node clusters, `--target-nodes` is
  required. Values are node names (e.g. `n1,n2`), not UUIDs — the CLI
  resolves names to host IDs automatically

## Sample pipeline.json

```json
{
  "pipelines": [
    {
      "name": "search",
      "system_prompt": "Answer questions using the provided context.",
      "hybrid_enabled": true,
      "vector_weight": 0.5,
      "tables": [
        {
          "table": "public.documents",
          "text_column": "content",
          "vector_column": "embedding"
        }
      ]
    }
  ]
}
```

Save this to `pipeline.json` and adjust `table`, `text_column`, and
`vector_column` to match the target schema before running the deploy command.

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

### Step 3: Prepare pipeline config

Confirm `pipeline.json` exists on disk and contains at least one pipeline with
valid table and column references.

### Step 4: Deploy RAG

> **Security note:** Use environment variables for API keys to avoid
> exposing them in shell history.

```bash
pgecloudctl databases rag deploy <db-id> \
  --embedding-llm-provider <provider> \
  --embedding-llm-model <embedding-model> \
  --embedding-llm-api-key "$OPENAI_API_KEY" \
  --completion-llm-provider <provider> \
  --completion-llm-model <completion-model> \
  --completion-llm-api-key "$OPENAI_API_KEY" \
  --pipeline-config pipeline.json \
  -o json
```

Common provider/model combinations:

| Role       | Provider | Model                      |
|------------|----------|----------------------------|
| Embedding  | openai   | text-embedding-3-small     |
| Embedding  | openai   | text-embedding-3-large     |
| Completion | openai   | gpt-4o                     |
| Completion | openai   | gpt-4o-mini                |

Optional flags:

| Flag               | Purpose                                              |
|--------------------|------------------------------------------------------|
| `--token-budget`   | Max tokens per RAG response                          |
| `--top-n`          | Number of retrieved chunks to pass to the LLM        |
| `--target-nodes`   | Node names to deploy on (auto-selects on single-node clusters) |

Capture `task_id` from the response as `<rag-task-id>`.

### Step 5: Wait for deployment

```bash
pgecloudctl tasks wait <rag-task-id> -o json --timeout 600
```

Expected output: JSON with `"status": "completed"`. Abort if status is
`failed` — read the `error` field for detail.

### Step 6 (private clusters only): Expose the RAG service via ingress

Skip to Step 7 if `node_location` is `public`.

List services to find the RAG service ID:

```bash
pgecloudctl databases services list <db-id> -o json
```

Identify the RAG service entry and capture its `id` as `<service-id>`.

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

### Step 7: Verify

```bash
pgecloudctl databases services list <db-id>
```

Expected output: a RAG service entry with `"status": "active"`.

## Verification

- `pgecloudctl databases services list <db-id> -o json` — RAG service present
  with `"status": "active"`
- If ingress was registered: `pgecloudctl ingresses services list <ingress-id>`
  — service appears with `"status": "registered"`

## Error Handling

| Exit Code | Meaning | Recovery |
|-----------|---------|----------|
| 0 | Success | Continue to next step |
| 1 | General error (invalid flags, API errors, constraints) | Check command output for details; verify flags and `pipeline.json` path |
| 2 | Authentication failure | Run `pgecloudctl auth login` |
| 3 | Request timeout | Retry the command; check network connectivity |
| 4 | Resource not found (database/cluster) | Verify `<db-id>` with `pgecloudctl databases list` |
