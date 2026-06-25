# clusters

Manage pgEdge Cloud clusters, including multi-region node groups and
cluster shares for multi-tenant deployments.

> **Async:** `create` and `delete` spawn a background task and return as
> soon as the request is accepted. Pass `--wait` (with optional `--timeout`,
> default 300, and `--interval`, default 5) to block until the task reaches a
> terminal state — exit 0 succeeded, 1 failed, 3 timeout. Without `--wait`,
> monitor with `pgecloudctl tasks list --subject-id <id>`. `delete` also
> prompts for confirmation unless `-y/--yes` is passed.

## Subcommands

### list

List all clusters in the current tenant.

**Usage:** `pgecloudctl clusters list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--limit int` | No | Maximum number of results to return |
| `--offset int` | No | Offset into the results for pagination |
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl clusters list --limit 20
```

**Example output (table):**

```text
ID                                    NAME          STATUS    REGIONS
a1b2c3d4-e5f6-7890-abcd-ef1234567890  prod-cluster  active    us-east-1,eu-west-1
b2c3d4e5-f6a7-8901-bcde-f12345678901  dev-cluster   active    us-west-2
```

---

### get

Get details for a specific cluster.

**Usage:** `pgecloudctl clusters get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl clusters get a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Example output (table):**

```text
FIELD               VALUE
ID                  a1b2c3d4-e5f6-7890-abcd-ef1234567890
Name                prod-cluster
Status              active
Node location       public
Regions             us-east-1, eu-west-1
Cloud account ID    c3d4e5f6-a7b8-9012-cdef-123456789012
Created             2026-01-15T10:00:00Z
```

---

### create

Create a new cluster across one or more cloud regions.

**Usage:** `pgecloudctl clusters create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name string` | Yes | Cluster name |
| `--cloud-account-id string` | Yes | Cloud account ID |
| `--regions strings` | Yes | Comma-separated list of regions |
| `--node-location string` | Yes | Node location (public or private) |
| `-h, --help` | No | help for create |

**Example:**

```bash
pgecloudctl clusters create \
    --name prod-cluster \
    --cloud-account-id c3d4e5f6-a7b8-9012-cdef-123456789012 \
    --regions us-east-1,eu-west-1 \
    --node-location public
```

**Example output (table):**

```text
FIELD               VALUE
ID                  a1b2c3d4-e5f6-7890-abcd-ef1234567890
Name                prod-cluster
Status              creating
```

---

### delete

Delete a cluster by ID.

**Usage:** `pgecloudctl clusters delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl clusters delete a1b2c3d4-e5f6-7890-abcd-ef1234567890 --yes
```

**Example output (table):**

```text
Cluster a1b2c3d4-e5f6-7890-abcd-ef1234567890 deleted.
```

---

## shares subgroup

Manage cluster shares for multi-tenant database hosting.

### shares list

List all shares belonging to a cluster.

**Usage:** `pgecloudctl clusters shares list <cluster-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl clusters shares list a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Example output (table):**

```text
ID                                    NAME        TENANCY     CAPACITY
d4e5f6a7-b8c9-0123-defa-234567890123  shared-s1   same        10
e5f6a7b8-c9d0-1234-efab-345678901234  shared-s2   allowlist   5
```

---

### shares get

Get details for a specific cluster share.

**Usage:** `pgecloudctl clusters shares get <cluster-id> <share-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl clusters shares get \
    a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
    d4e5f6a7-b8c9-0123-defa-234567890123
```

**Example output (table):**

```text
FIELD              VALUE
ID                 d4e5f6a7-b8c9-0123-defa-234567890123
Name               shared-s1
Tenancy            same
Capacity           10
Allowed tenants
```

---

### shares create

Create a new share on a cluster.

**Usage:** `pgecloudctl clusters shares create <cluster-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name string` | Yes | Share name |
| `--tenancy string` | Yes | Tenancy mode: same or allowlist |
| `--capacity int` | No | Share capacity |
| `--allowed-tenants strings` | No | Allowed tenant IDs (for allowlist tenancy) |
| `-h, --help` | No | help for create |

**Example:**

```bash
pgecloudctl clusters shares create a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
    --name shared-s1 \
    --tenancy same \
    --capacity 10
```

**Example output (table):**

```text
FIELD     VALUE
ID        d4e5f6a7-b8c9-0123-defa-234567890123
Name      shared-s1
Tenancy   same
Capacity  10
```

---

### shares delete

Delete a cluster share by ID.

**Usage:** `pgecloudctl clusters shares delete <cluster-id> <share-id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl clusters shares delete \
    a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
    d4e5f6a7-b8c9-0123-defa-234567890123 \
    --yes
```

**Example output (table):**

```text
Share d4e5f6a7-b8c9-0123-defa-234567890123 deleted.
```
