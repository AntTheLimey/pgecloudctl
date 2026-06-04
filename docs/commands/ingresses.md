# ingresses

Manage pgEdge Cloud ingresses — regional load balancers that route
external traffic to database services such as the MCP or RAG server.

## Subcommands

### list

List all ingresses, with optional time-range filters.

**Usage:** `pgecloudctl ingresses list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--created-after string` | No | Filter: created after this RFC3339 timestamp |
| `--created-before string` | No | Filter: created before this RFC3339 timestamp |
| `--limit int` | No | Maximum number of results to return |
| `--offset int` | No | Offset into the results for pagination |
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl ingresses list
```

**Example output (table):**

```text
ID                                    NAME          REGION       CLUSTER                                STATUS
c5d6e7f8-a9b0-1234-cdef-345678901234  prod-ingress  us-east-1    a1b2c3d4-e5f6-7890-abcd-ef1234567890   active
```

---

### get

Get details for a specific ingress.

**Usage:** `pgecloudctl ingresses get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl ingresses get c5d6e7f8-a9b0-1234-cdef-345678901234
```

**Example output (table):**

```text
FIELD       VALUE
ID          c5d6e7f8-a9b0-1234-cdef-345678901234
Name        prod-ingress
Region      us-east-1
Cluster ID  a1b2c3d4-e5f6-7890-abcd-ef1234567890
Status      active
Hostname    prod-ingress.c5d6e7f8.pgedge.io
Created     2026-02-10T11:00:00Z
```

---

### create

Create a new ingress in a specific region, associated with a cluster.

**Usage:** `pgecloudctl ingresses create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name string` | Yes | Ingress name |
| `--cluster-id string` | Yes | Cluster ID to associate with the ingress |
| `--region string` | Yes | Cloud region for the ingress |
| `-h, --help` | No | help for create |

**Example:**

```bash
pgecloudctl ingresses create \
    --name prod-ingress \
    --cluster-id a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
    --region us-east-1
```

**Example output (table):**

```text
FIELD     VALUE
ID        c5d6e7f8-a9b0-1234-cdef-345678901234
Name      prod-ingress
Region    us-east-1
Status    creating
```

---

### delete

Delete an ingress by ID.

**Usage:** `pgecloudctl ingresses delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl ingresses delete c5d6e7f8-a9b0-1234-cdef-345678901234 --yes
```

**Example output (table):**

```text
Ingress c5d6e7f8-a9b0-1234-cdef-345678901234 deleted.
```

---

## services subgroup

Manage which database services are exposed through an ingress.

### services list

List all services registered on a specific ingress.

**Usage:** `pgecloudctl ingresses services list <ingress-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl ingresses services list c5d6e7f8-a9b0-1234-cdef-345678901234
```

**Example output (table):**

```text
SERVICE ID                            DATABASE ID                           TYPE
a7b8c9d0-e1f2-3456-abcd-567890123456  f6a7b8c9-d0e1-2345-fabc-456789012345  mcp
```

---

### services register

Register a database service on an ingress to expose it externally.

**Usage:** `pgecloudctl ingresses services register <ingress-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--database-id string` | Yes | Database ID to register |
| `--service-id string` | Yes | Service ID to expose |
| `-h, --help` | No | help for register |

**Example:**

```bash
pgecloudctl ingresses services register c5d6e7f8-a9b0-1234-cdef-345678901234 \
    --database-id f6a7b8c9-d0e1-2345-fabc-456789012345 \
    --service-id a7b8c9d0-e1f2-3456-abcd-567890123456
```

**Example output (table):**

```text
Service a7b8c9d0-e1f2-3456-abcd-567890123456 registered on ingress c5d6e7f8-a9b0-1234-cdef-345678901234.
```

---

### services deregister

Deregister a service from an ingress.

**Usage:** `pgecloudctl ingresses services deregister <ingress-id> <service-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for deregister |

**Example:**

```bash
pgecloudctl ingresses services deregister \
    c5d6e7f8-a9b0-1234-cdef-345678901234 \
    a7b8c9d0-e1f2-3456-abcd-567890123456
```

**Example output (table):**

```text
Service a7b8c9d0-e1f2-3456-abcd-567890123456 deregistered from ingress c5d6e7f8-a9b0-1234-cdef-345678901234.
```
