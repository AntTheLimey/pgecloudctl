# invites

Manage pgEdge Cloud team invites — send, accept, and revoke invitations
for new team members.

## Subcommands

### list

List all pending invites in the current tenant.

**Usage:** `pgecloudctl invites list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl invites list
```

**Example output (table):**

```
ID                                    EMAIL                  STATUS     EXPIRES
d6e7f8a9-b0c1-2345-defa-456789012345  alice@example.com      pending    2026-06-10T00:00:00Z
e7f8a9b0-c1d2-3456-efab-567890123456  bob@example.com        pending    2026-06-15T00:00:00Z
```

---

### get

Get details for a specific invite.

**Usage:** `pgecloudctl invites get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl invites get d6e7f8a9-b0c1-2345-defa-456789012345
```

**Example output (table):**

```
FIELD    VALUE
ID       d6e7f8a9-b0c1-2345-defa-456789012345
Email    alice@example.com
Status   pending
Expires  2026-06-10T00:00:00Z
Created  2026-06-03T09:00:00Z
```

---

### create

Create and send a new invite to an email address.

**Usage:** `pgecloudctl invites create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--email string` | Yes | Email address to invite |
| `--expiration int` | No | Invite expiration in hours (optional) |
| `-h, --help` | No | help for create |

**Example:**

```bash
pgecloudctl invites create \
    --email alice@example.com \
    --expiration 168
```

**Example output (table):**

```
FIELD    VALUE
ID       d6e7f8a9-b0c1-2345-defa-456789012345
Email    alice@example.com
Status   pending
Expires  2026-06-10T00:00:00Z
```

---

### delete

Delete (revoke) a pending invite by ID.

**Usage:** `pgecloudctl invites delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl invites delete d6e7f8a9-b0c1-2345-defa-456789012345 --yes
```

**Example output (table):**

```
Invite d6e7f8a9-b0c1-2345-defa-456789012345 deleted.
```

---

### accept

Accept an invite using its ID and token (typically used by the invitee).

**Usage:** `pgecloudctl invites accept <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--token string` | Yes | Invite token |
| `-h, --help` | No | help for accept |

**Example:**

```bash
pgecloudctl invites accept d6e7f8a9-b0c1-2345-defa-456789012345 \
    --token inv_abc123xyz789
```

**Example output (table):**

```
Invite accepted. You have joined the team.
```
