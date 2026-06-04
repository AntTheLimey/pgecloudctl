# ssh-keys

Manage SSH public keys registered with pgEdge Cloud for node access.

## Subcommands

### list

List all SSH keys in the current tenant.

**Usage:** `pgecloudctl ssh-keys list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl ssh-keys list
```

**Example output (table):**

```text
ID                                    NAME          CREATED
a3b4c5d6-e7f8-9012-abcd-123456789012  deploy-key    2026-01-20T14:00:00Z
b4c5d6e7-f8a9-0123-bcde-234567890123  laptop-key    2026-02-05T10:30:00Z
```

---

### get

Get details for a specific SSH key.

**Usage:** `pgecloudctl ssh-keys get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl ssh-keys get a3b4c5d6-e7f8-9012-abcd-123456789012
```

**Example output (table):**

```text
FIELD       VALUE
ID          a3b4c5d6-e7f8-9012-abcd-123456789012
Name        deploy-key
Public key  ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... deploy@ci
Created     2026-01-20T14:00:00Z
```

---

### create

Register a new SSH public key.

**Usage:** `pgecloudctl ssh-keys create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name string` | Yes | SSH key name |
| `--public-key string` | Yes | SSH public key value |
| `-h, --help` | No | help for create |

**Example:**

```bash
pgecloudctl ssh-keys create \
    --name deploy-key \
    --public-key "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... deploy@ci"
```

**Example output (table):**

```text
FIELD     VALUE
ID        a3b4c5d6-e7f8-9012-abcd-123456789012
Name      deploy-key
Created   2026-01-20T14:00:00Z
```

---

### delete

Delete an SSH key by ID.

**Usage:** `pgecloudctl ssh-keys delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl ssh-keys delete a3b4c5d6-e7f8-9012-abcd-123456789012 --yes
```

**Example output (table):**

```text
SSH key a3b4c5d6-e7f8-9012-abcd-123456789012 deleted.
```
