# memberships

Manage pgEdge Cloud team memberships — list current members and remove
members from the tenant.

## Subcommands

### list

List all current team memberships in the tenant.

**Usage:** `pgecloudctl memberships list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl memberships list
```

**Example output (table):**

```text
ID                                    EMAIL                  ROLE     JOINED
f8a9b0c1-d2e3-4567-fabc-678901234567  alice@example.com      admin    2026-01-15T09:00:00Z
a9b0c1d2-e3f4-5678-abcd-789012345678  bob@example.com        member   2026-03-01T10:00:00Z
```

---

### delete

Remove a team member from the tenant by their membership ID.

**Usage:** `pgecloudctl memberships delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl memberships delete a9b0c1d2-e3f4-5678-abcd-789012345678 --yes
```

**Example output (table):**

```text
Membership a9b0c1d2-e3f4-5678-abcd-789012345678 deleted.
```
