# databases

Manage pgEdge Cloud databases, along with MCP servers, RAG servers, and
other services deployed alongside them.

## Subcommands

### list

List databases, optionally filtered to a specific cluster.

**Usage:** `pgecloudctl databases list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--cluster-id string` | No | Filter by cluster ID |
| `--limit int` | No | Maximum number of results to return |
| `--offset int` | No | Offset into the results for pagination |
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl databases list \
    --cluster-id a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Example output (table):**

```text
ID                                    NAME      PG VERSION    STATUS    CLUSTER
f6a7b8c9-d0e1-2345-fabc-456789012345  mydb      16            active    a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

---

### get

Get details for a specific database.

**Usage:** `pgecloudctl databases get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl databases get f6a7b8c9-d0e1-2345-fabc-456789012345
```

**Example output (table):**

```text
FIELD         VALUE
ID            f6a7b8c9-d0e1-2345-fabc-456789012345
Name          mydb
Display name  My Production DB
PG version    16
Status        active
Cluster ID    a1b2c3d4-e5f6-7890-abcd-ef1234567890
Created       2026-02-01T09:00:00Z
```

---

### create

Create a new PostgreSQL database on an existing cluster.

**Usage:** `pgecloudctl databases create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name string` | Yes | Database name |
| `--cluster-id string` | Yes | Cluster ID to deploy the database on |
| `--pg-version string` | Yes | PostgreSQL version (e.g. 16) |
| `-h, --help` | No | help for create |

**Example:**

```bash
pgecloudctl databases create \
    --name mydb \
    --cluster-id a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
    --pg-version 16
```

**Example output (table):**

```text
FIELD       VALUE
ID          f6a7b8c9-d0e1-2345-fabc-456789012345
Name        mydb
PG version  16
Status      creating
```

---

### update

Update mutable properties of an existing database.

**Usage:** `pgecloudctl databases update <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--display-name string` | No | Display name for the database |
| `--options strings` | No | Comma-separated list of options |
| `-h, --help` | No | help for update |

**Example:**

```bash
pgecloudctl databases update f6a7b8c9-d0e1-2345-fabc-456789012345 \
    --display-name "My Production DB"
```

**Example output (table):**

```text
Database f6a7b8c9-d0e1-2345-fabc-456789012345 updated.
```

---

### delete

Delete a database by ID.

**Usage:** `pgecloudctl databases delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl databases delete f6a7b8c9-d0e1-2345-fabc-456789012345 --yes
```

**Example output (table):**

```text
Database f6a7b8c9-d0e1-2345-fabc-456789012345 deleted.
```

---

## services subgroup

Manage services (MCP, RAG) deployed alongside a database.

> **WARNING: Destructive API behavior.** The pgEdge Cloud API treats the
> `services` field as fully declarative — whatever you send REPLACES all
> existing services on the database. If you deploy a RAG server without
> including the existing MCP server in the request, the MCP server will be
> destroyed. The CLI handles this automatically via a read-modify-write
> pattern, but direct API callers must include all existing services in
> every update.

### Asynchronous operations and `--wait`

Service mutations — `mcp deploy`, `mcp update`, `rag deploy`, `rag update`,
and `services remove` — are asynchronous. The API accepts the request and
spawns a background task, so by default the command returns as soon as the
request is accepted, not when the work completes. Exit 0 means "accepted".

These commands accept three shared flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--wait` | false | Block until the spawned task reaches a terminal state |
| `--timeout int` | 300 | Max seconds to wait when `--wait` is set |
| `--interval int` | 5 | Polling interval in seconds when `--wait` is set |

With `--wait`, the exit code reflects the real outcome: 0 if the task
succeeded, 1 if it failed, 3 if the timeout was exceeded. Without `--wait`,
the command prints the `tasks list --subject-id <db-id>` invocation you can
use to monitor progress yourself.

### services list

List all services deployed on a database.

**Usage:** `pgecloudctl databases services list <db-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl databases services list f6a7b8c9-d0e1-2345-fabc-456789012345
```

**Example output (table):**

```text
ID                                    TYPE    STATUS
a7b8c9d0-e1f2-3456-abcd-567890123456  mcp     active
b8c9d0e1-f2a3-4567-bcde-678901234567  rag     active
```

---

### services get

Get details of a specific service deployed on a database.

**Usage:** `pgecloudctl databases services get <db-id> <service-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl databases services get \
    f6a7b8c9-d0e1-2345-fabc-456789012345 \
    a7b8c9d0-e1f2-3456-abcd-567890123456
```

**Example output (table):**

```text
FIELD     VALUE
ID        a7b8c9d0-e1f2-3456-abcd-567890123456
Type      mcp
Status    active
Endpoint  https://mcp.a7b8c9d0.pgedge.io
```

---

### services remove

Remove a service type (mcp or rag) from a database.

**Usage:** `pgecloudctl databases services remove <db-id> <type> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for remove |

**Example:**

```bash
pgecloudctl databases services remove \
    f6a7b8c9-d0e1-2345-fabc-456789012345 \
    mcp
```

**Example output (table):**

```text
Service mcp removed from database f6a7b8c9-d0e1-2345-fabc-456789012345.
```

---

## mcp subgroup

Deploy and configure the pgEdge MCP server on a database. The MCP server
exposes a Model Context Protocol endpoint that AI agents can connect to
for structured database access.

### mcp deploy

Deploy an MCP server alongside an existing database.

**Usage:** `pgecloudctl databases mcp deploy <db-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--allow-writes` | No | Grant the MCP service read-write access (WARNING: allows LLM to modify data) |
| `--embedding-provider string` | No | Embedding provider: ollama, openai, or voyage |
| `--embedding-model string` | No | Embedding model identifier (required when --embedding-provider is set) |
| `--embedding-api-key string` | No | API key for the embedding provider (required for openai and voyage) |
| `--ollama-url string` | No | Endpoint URL for an Ollama server (required when --embedding-provider is ollama) |
| `--target-nodes strings` | No | Node names to deploy on (e.g. n1,n2). Auto-selects if cluster has one node |
| `--init-tokens string` | No | Bearer token forwarded to the MCP server as INIT_TOKENS |
| `--init-users string` | No | Comma-separated username:password pairs forwarded as INIT_USERS |
| `-h, --help` | No | help for deploy |

Node names are resolved to host UUIDs automatically. On single-node
clusters, `--target-nodes` can be omitted and the node is auto-selected.
On multi-node clusters, `--target-nodes` is required.

**Example:**

```bash
pgecloudctl databases mcp deploy f6a7b8c9-d0e1-2345-fabc-456789012345 \
    --embedding-provider openai \
    --embedding-model text-embedding-3-small \
    --embedding-api-key "$OPENAI_API_KEY"
```

**Example output (table):**

```text
FIELD     VALUE
ID        a7b8c9d0-e1f2-3456-abcd-567890123456
Type      mcp
Status    deploying
Endpoint  https://mcp.a7b8c9d0.pgedge.io
```

---

### mcp update

Update the MCP server configuration on a database. Accepts the same flags
as `mcp deploy`.

**Usage:** `pgecloudctl databases mcp update <db-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--allow-writes` | No | Grant the MCP service read-write access (WARNING: allows LLM to modify data) |
| `--embedding-provider string` | No | Embedding provider: ollama, openai, or voyage |
| `--embedding-model string` | No | Embedding model identifier (required when --embedding-provider is set) |
| `--embedding-api-key string` | No | API key for the embedding provider (required for openai and voyage) |
| `--ollama-url string` | No | Endpoint URL for an Ollama server (required when --embedding-provider is ollama) |
| `--target-nodes strings` | No | Node names to deploy on (e.g. n1,n2). Auto-selects if cluster has one node |
| `--init-tokens string` | No | Bearer token forwarded to the MCP server as INIT_TOKENS |
| `--init-users string` | No | Comma-separated username:password pairs forwarded as INIT_USERS |
| `-h, --help` | No | help for update |

**Example:**

```bash
pgecloudctl databases mcp update f6a7b8c9-d0e1-2345-fabc-456789012345 \
    --allow-writes \
    --target-nodes node1,node2
```

**Example output (table):**

```text
MCP service updated on database f6a7b8c9-d0e1-2345-fabc-456789012345.
```

---

## rag subgroup

Deploy and configure the pgEdge RAG server (Ellie) on a database. The RAG
server provides retrieval-augmented generation pipelines backed by pgvector.

### rag deploy

Deploy a RAG server alongside an existing database.

**Usage:** `pgecloudctl databases rag deploy <db-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--embedding-llm-provider string` | No | Embedding LLM provider (e.g. openai, voyage) |
| `--embedding-llm-model string` | No | Embedding LLM model identifier |
| `--embedding-llm-api-key string` | No | API key for the embedding LLM provider |
| `--completion-llm-provider string` | No | Completion LLM provider (e.g. openai) |
| `--completion-llm-model string` | No | Completion LLM model identifier |
| `--completion-llm-api-key string` | No | API key for the completion LLM provider |
| `--pipeline-config string` | No | Path to a JSON file containing pipeline definitions |
| `--target-nodes strings` | No | Node names to deploy on (e.g. n1,n2). Auto-selects if cluster has one node |
| `--top-n int` | No | Default number of results to retrieve per pipeline |
| `--token-budget int` | No | Default max completion tokens across all pipelines |
| `-h, --help` | No | help for deploy |

Node names are resolved to host UUIDs automatically. On single-node
clusters, `--target-nodes` can be omitted and the node is auto-selected.
On multi-node clusters, `--target-nodes` is required.

**Example:**

```bash
pgecloudctl databases rag deploy f6a7b8c9-d0e1-2345-fabc-456789012345 \
    --embedding-llm-provider openai \
    --embedding-llm-model text-embedding-3-small \
    --embedding-llm-api-key "$OPENAI_API_KEY" \
    --completion-llm-provider openai \
    --completion-llm-model gpt-4o \
    --completion-llm-api-key "$OPENAI_API_KEY" \
    --top-n 5 \
    --token-budget 2048
```

**Example output (table):**

```text
FIELD     VALUE
ID        b8c9d0e1-f2a3-4567-bcde-678901234567
Type      rag
Status    deploying
Endpoint  https://rag.b8c9d0e1.pgedge.io
```

---

### rag update

Update the RAG server configuration on a database. Accepts the same flags
as `rag deploy`.

**Usage:** `pgecloudctl databases rag update <db-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--embedding-llm-provider string` | No | Embedding LLM provider (e.g. openai, voyage) |
| `--embedding-llm-model string` | No | Embedding LLM model identifier |
| `--embedding-llm-api-key string` | No | API key for the embedding LLM provider |
| `--completion-llm-provider string` | No | Completion LLM provider (e.g. openai) |
| `--completion-llm-model string` | No | Completion LLM model identifier |
| `--completion-llm-api-key string` | No | API key for the completion LLM provider |
| `--pipeline-config string` | No | Path to a JSON file containing pipeline definitions |
| `--target-nodes strings` | No | Node names to deploy on (e.g. n1,n2). Auto-selects if cluster has one node |
| `--top-n int` | No | Default number of results to retrieve per pipeline |
| `--token-budget int` | No | Default max completion tokens across all pipelines |
| `-h, --help` | No | help for update |

**Example:**

```bash
pgecloudctl databases rag update f6a7b8c9-d0e1-2345-fabc-456789012345 \
    --top-n 10 \
    --token-budget 4096
```

**Example output (table):**

```text
RAG service updated on database f6a7b8c9-d0e1-2345-fabc-456789012345.
```
