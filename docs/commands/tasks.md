# tasks

Manage pgEdge Cloud tasks — long-running asynchronous operations such as
cluster creation, database provisioning, and service deployments.

## Subcommands

### list

List tasks, with optional filters by status, subject ID, and subject kind.

**Usage:** `pgecloudctl tasks list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--status string` | No | Filter by status (queued, running, succeeded, failed) |
| `--subject-id string` | No | Filter by subject ID |
| `--subject-kind string` | No | Filter by subject kind |
| `--limit int` | No | Maximum number of results to return |
| `--offset int` | No | Offset into the results for pagination |
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl tasks list --status running
```

**Example output (table):**

```
ID                                    KIND              STATUS     SUBJECT ID                            CREATED
b0c1d2e3-f4a5-6789-bcde-890123456789  cluster.create    running    a1b2c3d4-e5f6-7890-abcd-ef1234567890   2026-06-04T12:00:00Z
c1d2e3f4-a5b6-7890-cdef-901234567890  database.create   succeeded  f6a7b8c9-d0e1-2345-fabc-456789012345   2026-06-04T11:30:00Z
```

---

### get

Get details for a specific task, including logs and error messages.

**Usage:** `pgecloudctl tasks get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl tasks get b0c1d2e3-f4a5-6789-bcde-890123456789
```

**Example output (table):**

```
FIELD         VALUE
ID            b0c1d2e3-f4a5-6789-bcde-890123456789
Kind          cluster.create
Status        running
Subject ID    a1b2c3d4-e5f6-7890-abcd-ef1234567890
Subject kind  cluster
Created       2026-06-04T12:00:00Z
Updated       2026-06-04T12:01:30Z
```

---

### wait

Poll a task until it reaches a terminal state (succeeded or failed).
Exits with code 0 on success, 1 on failure, and 3 on timeout.

**Usage:** `pgecloudctl tasks wait <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--timeout int` | No | Maximum seconds to wait for task completion (default 300) |
| `--interval int` | No | Polling interval in seconds (default 5) |
| `-h, --help` | No | help for wait |

**Example:**

```bash
pgecloudctl tasks wait b0c1d2e3-f4a5-6789-bcde-890123456789 \
    --timeout 600 \
    --interval 10
```

**Example output (table):**

```
Waiting for task b0c1d2e3-f4a5-6789-bcde-890123456789...
Task succeeded.
```
