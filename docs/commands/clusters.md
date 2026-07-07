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

Create a new cluster across one or more cloud regions. Networks
(CIDR + subnets), node sizing, firewall rules, and backup stores can
all be set at create time — a cluster is fully specifiable from the
CLI.

**Usage:** `pgecloudctl clusters create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name string` | Yes | Cluster name |
| `--cloud-account-id string` | Yes | Cloud account ID |
| `--regions strings` | Yes | Comma-separated list of regions |
| `--node-location string` | Yes | Node location (public or private) |
| `--backup-store-id strings` | No | Backup store ID to attach (repeatable; required to host a DB) |
| `--firewall-rule stringArray` | No | Firewall rule (repeatable); name one of http, https, postgres, ssh |
| `--network stringArray` | No | Network settings (repeatable, one per region) |
| `--node stringArray` | No | Node settings (repeatable) |
| `--instance-type string` | No | Instance type for all nodes (shorthand for `--node`) |
| `--volume-size int` | No | Volume size in GB for all nodes (shorthand for `--node`) |
| `--wait` | No | Block until the create task reaches a terminal state |
| `--timeout int` | No | Max seconds to wait when `--wait` is set (default 300) |
| `--interval int` | No | Polling interval in seconds when `--wait` is set (default 5) |
| `-h, --help` | No | help for create |

The structured flags take comma-separated key=value pairs; repeat a
list-valued key to add elements. On single-region clusters `region=`
may be omitted:

- `--firewall-rule name=https,port=443,sources=0.0.0.0/0`
- `--network region=us-east-1,cidr=10.4.0.0/16,public-subnets=10.4.1.0/24,private-subnets=10.4.128.0/24`
- `--node name=n1,region=us-east-1,instance-type=r7g.medium,volume-size=30`
  (also accepts volume-iops, volume-type, availability-zone)

`volume-type=gp3` is rejected — gp3 nodes wedge later firewall-rule
updates and leave the cluster degraded (CLOUD-480). Omit volume-type
to use the default (gp2). `--node` and the
`--instance-type`/`--volume-size` shorthand are mutually exclusive;
the shorthand creates one node per region, named n1, n2, ...

**Example:**

```bash
pgecloudctl clusters create \
    --name prod-cluster \
    --cloud-account-id c3d4e5f6-a7b8-9012-cdef-123456789012 \
    --regions us-east-1 \
    --node-location public \
    --backup-store-id f2a3b4c5-d6e7-8901-fabc-012345678901 \
    --node name=n1,instance-type=r7g.medium,volume-size=30 \
    --network cidr=10.4.0.0/16,public-subnets=10.4.1.0/24 \
    --firewall-rule name=postgres,port=5432,sources=0.0.0.0/0
```

**Example output (table):**

```text
Cluster "prod-cluster" created (id: a1b2c3d4-e5f6-7890-abcd-ef1234567890, status: creating).
Track progress: pgecloudctl tasks list --subject-id a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

---

### update

Update a cluster in place. Firewall rules and backup stores append to
the cluster's existing values (read-modify-write); regions replace the
current list when supplied. At least one flag is required.

**Usage:** `pgecloudctl clusters update <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--firewall-rule stringArray` | No | Firewall rule to append (repeatable) |
| `--backup-store-id strings` | No | Backup store ID to attach (repeatable) |
| `--regions strings` | No | Replace the cluster's regions |
| `--wait` | No | Block until the update task reaches a terminal state |
| `--timeout int` | No | Max seconds to wait when `--wait` is set (default 300) |
| `--interval int` | No | Polling interval in seconds when `--wait` is set (default 5) |
| `-h, --help` | No | help for update |

**Example:**

```bash
pgecloudctl clusters update a1b2c3d4-e5f6-7890-abcd-ef1234567890 \
    --firewall-rule name=https,port=443,sources=0.0.0.0/0
```

**Example output (table):**

```text
Cluster a1b2c3d4-e5f6-7890-abcd-ef1234567890 updated.
```

---

### delete

Delete a cluster by ID.

**Usage:** `pgecloudctl clusters delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `--force` | No | Also delete all databases and cloud infrastructure, bypassing status and database-existence checks |
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
