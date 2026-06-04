# auth

Manage authentication credentials for pgEdge Cloud.

## Subcommands

### login

Authenticate with pgEdge Cloud using a client ID and secret. Credentials
can be passed via `--client-id` / `--client-secret` flags or the
`PGEDGE_CLIENT_ID` / `PGEDGE_CLIENT_SECRET` environment variables.
A token is fetched and cached in the config directory.

**Usage:** `pgecloudctl auth login [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for login |

**Example:**

```bash
pgecloudctl auth login \
    --client-id my-client-id \
    --client-secret my-client-secret
```

**Example output (table):**

```
Logged in successfully.
```

---

### status

Show current authentication state, including token source and expiry.

**Usage:** `pgecloudctl auth status [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for status |

**Example:**

```bash
pgecloudctl auth status
```

**Example output (table):**

```
AUTHENTICATED    SOURCE    EXPIRES
true             config    2026-12-31T23:59:59Z
```

---

### logout

Clear stored credentials and cached tokens from the config directory.

**Usage:** `pgecloudctl auth logout [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for logout |

**Example:**

```bash
pgecloudctl auth logout
```

**Example output (table):**

```
Logged out successfully.
```
