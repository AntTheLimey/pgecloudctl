# Expose Service

Create an ingress and register a database service to make it reachable on a
private cluster.

## Prerequisites

- A private cluster with at least one active database
- The database ID (`<db-id>`)
- At least one deployed service on the database (MCP, RAG, or other)

## Decision Points

- **Ingress already exists** — if an ingress for this cluster was previously
  created, skip Step 3 and use the existing `<ingress-id>`
- **Multiple services** — run Steps 4 and 5 once per service that needs to be
  registered

## Steps

### Step 1: List database services

```bash
pgecloudctl databases services list <db-id> -o json
```

Expected output: a JSON array of service objects. Identify the target service
and capture its `id` as `<service-id>`.

### Step 2: Get the cluster ID

```bash
pgecloudctl databases get <db-id> -o json
```

Capture `cluster_id` as `<cluster-id>`. Also note the cluster's primary region
for use in Step 3.

### Step 3: Create the ingress

Skip this step if an ingress for `<cluster-id>` already exists. To check:

```bash
pgecloudctl ingresses list -o json
```

If no matching ingress is found, create one:

```bash
pgecloudctl ingresses create \
  --name <ingress-name> \
  --cluster-id <cluster-id> \
  --region <region> \
  -o json
```

Capture `id` from the response as `<ingress-id>`.

### Step 4: Register the service

```bash
pgecloudctl ingresses services register <ingress-id> \
  --database-id <db-id> \
  --service-id <service-id>
```

No output is returned on success (exit code 0). Repeat this step for each
additional service that needs to be exposed.

### Step 5: Verify

```bash
pgecloudctl ingresses services list <ingress-id>
```

Expected output: the registered service appears in the list with
`"status": "registered"`.

## Verification

- `pgecloudctl ingresses services list <ingress-id>` — target service present
  with `"status": "registered"`
- Connection to the service endpoint succeeds from an allowed network

## Error Handling

| Exit Code | Meaning                                     | Recovery                                                    |
|-----------|---------------------------------------------|-------------------------------------------------------------|
| 1         | Database or service not found               | Verify IDs with `databases list` and `databases services list` |
| 2         | Invalid or missing flag                     | Check `--cluster-id`, `--region`, `--database-id`, `--service-id` |
| 3         | Ingress already exists for this cluster     | Use existing ingress ID; skip Step 3                        |
| 4         | Service not in a registerable state         | Ensure the service deployment task completed successfully   |
| 1 (task)  | Ingress creation task failed                | Read `error` field; verify region availability and retry    |
