# backup-stores

Manage pgEdge Cloud backup stores — the remote storage targets where
database backups are written.

> **Async:** `create` and `delete` spawn a background task and return as
> soon as the request is accepted. Pass `--wait` (with optional `--timeout`,
> default 300, and `--interval`, default 5) to block until the task reaches a
> terminal state — exit 0 succeeded, 1 failed, 3 timeout. Without `--wait`,
> monitor with `pgecloudctl tasks list --subject-id <id>`. `delete` also
> prompts for confirmation unless `-y/--yes` is passed.

## Subcommands

### list

List all backup stores, with optional time-range filters.

**Usage:** `pgecloudctl backup-stores list [flags]`

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
pgecloudctl backup-stores list
```

**Example output (table):**

```text
ID                                    NAME          REGION       CLOUD ACCOUNT
f2a3b4c5-d6e7-8901-fabc-012345678901  primary-bkp   us-east-1    c3d4e5f6-a7b8-9012-cdef-123456789012
```

---

### get

Get details for a specific backup store.

**Usage:** `pgecloudctl backup-stores get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl backup-stores get f2a3b4c5-d6e7-8901-fabc-012345678901
```

**Example output (table):**

```text
FIELD             VALUE
ID                f2a3b4c5-d6e7-8901-fabc-012345678901
Name              primary-bkp
Region            us-east-1
Cloud account ID  c3d4e5f6-a7b8-9012-cdef-123456789012
Created           2026-01-10T08:00:00Z
```

---

### create

Create a new backup store in a specific cloud region.

**Usage:** `pgecloudctl backup-stores create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name string` | Yes | Backup store name |
| `--cloud-account-id string` | Yes | Cloud account ID |
| `--region string` | Yes | Region for the backup store |
| `-h, --help` | No | help for create |

**Example:**

```bash
pgecloudctl backup-stores create \
    --name primary-bkp \
    --cloud-account-id c3d4e5f6-a7b8-9012-cdef-123456789012 \
    --region us-east-1
```

**Example output (table):**

```text
FIELD     VALUE
ID        f2a3b4c5-d6e7-8901-fabc-012345678901
Name      primary-bkp
Region    us-east-1
Status    creating
```

---

### delete

Delete a backup store by ID.

**Usage:** `pgecloudctl backup-stores delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl backup-stores delete f2a3b4c5-d6e7-8901-fabc-012345678901 --yes
```

**Example output (table):**

```text
Backup store f2a3b4c5-d6e7-8901-fabc-012345678901 deleted.
```
