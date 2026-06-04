# backups

Manage pgEdge Cloud backups for databases.

## Subcommands

### list

List backups, with optional filters by database, kind, and time range.

**Usage:** `pgecloudctl backups list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--database-id string` | No | Filter backups to a specific database ID |
| `--kind string` | No | Filter backups to a specific kind |
| `--created-after string` | No | Filter: created after this RFC3339 timestamp |
| `--created-before string` | No | Filter: created before this RFC3339 timestamp |
| `--limit int` | No | Maximum number of results to return |
| `--offset int` | No | Offset into the results for pagination |
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl backups list \
    --database-id f6a7b8c9-d0e1-2345-fabc-456789012345 \
    --created-after 2026-01-01T00:00:00Z
```

**Example output (table):**

```text
ID                                    NAME          KIND    STATUS     CREATED
c9d0e1f2-a3b4-5678-cdef-789012345678  nightly-bk1   full    complete   2026-06-01T02:00:00Z
d0e1f2a3-b4c5-6789-defa-890123456789  nightly-bk2   full    complete   2026-06-02T02:00:00Z
```

---

### get

Get details for a specific backup.

**Usage:** `pgecloudctl backups get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl backups get c9d0e1f2-a3b4-5678-cdef-789012345678
```

**Example output (table):**

```text
FIELD        VALUE
ID           c9d0e1f2-a3b4-5678-cdef-789012345678
Name         nightly-bk1
Kind         full
Status       complete
Database ID  f6a7b8c9-d0e1-2345-fabc-456789012345
Provider     s3
Created      2026-06-01T02:00:00Z
```

---

### create

Create a new on-demand backup for a database.

**Usage:** `pgecloudctl backups create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--database-id string` | Yes | Database ID to back up |
| `--provider string` | No | Backup provider |
| `--name string` | No | Optional backup name |
| `--type string` | No | Optional backup type |
| `--target-nodes strings` | No | Comma-separated list of target nodes |
| `-h, --help` | No | help for create |

**Example:**

```bash
pgecloudctl backups create \
    --database-id f6a7b8c9-d0e1-2345-fabc-456789012345 \
    --name pre-migration \
    --provider s3
```

**Example output (table):**

```text
FIELD        VALUE
ID           e1f2a3b4-c5d6-7890-efab-901234567890
Name         pre-migration
Status       running
Database ID  f6a7b8c9-d0e1-2345-fabc-456789012345
```

---

### delete

Delete a backup by ID.

**Usage:** `pgecloudctl backups delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl backups delete c9d0e1f2-a3b4-5678-cdef-789012345678 --yes
```

**Example output (table):**

```text
Backup c9d0e1f2-a3b4-5678-cdef-789012345678 deleted.
```

---

### url

Get a time-limited download URL for a completed backup.

**Usage:** `pgecloudctl backups url <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for url |

**Example:**

```bash
pgecloudctl backups url c9d0e1f2-a3b4-5678-cdef-789012345678
```

**Example output (table):**

```text
URL
https://backups.pgedge.com/c9d0e1f2-a3b4-5678-cdef-789012345678/download?token=eyJhbGc...
```
