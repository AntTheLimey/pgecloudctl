---
name: pgecloudctl
description: "Use this skill when the user wants to manage pgEdge Cloud
  resources — clusters, databases, backups, services (MCP/RAG), ingresses,
  SSH keys, invites, or team memberships. Triggers on: pgEdge Cloud,
  pgecloudctl, or any mention of managing pgEdge infrastructure. Also
  triggers when the user asks about deploying MCP or RAG servers on
  pgEdge."
---

# pgecloudctl Skill

## Layer 1: Setup

### Check installation

```bash
pgecloudctl --version
```

If the command is not found:

- Homebrew: `brew install AntTheLimey/tap/pgecloudctl`
- Script: `curl -fsSL https://raw.githubusercontent.com/AntTheLimey/pgecloudctl/main/install.sh | sh`
- Go: `go install github.com/AntTheLimey/pgecloudctl@latest`

The binary embeds its own complete reference: `pgecloudctl llms`
prints llms-full.txt (every command, flag, and workflow) for the
installed version. Prefer it over guessing from `--help`. This skill
can be (re)installed from any binary with `pgecloudctl skill install`.

### Check authentication

```bash
pgecloudctl auth status -o json
```

- If exit code is non-zero or `authenticated` is false, tell the user to
  run `pgecloudctl auth login`.
- Auth can use env vars (`PGEDGE_CLIENT_ID` + `PGEDGE_CLIENT_SECRET`),
  CLI flags, or the config file written by `auth login`.
- If auth is broken and the cause is unclear, run `pgecloudctl doctor` for
  diagnostics.

---

## Layer 2: Command Reference

Consult `knowledge-base.md` (in this same directory) for all command
signatures, required/optional flags, and output field names.

### Key patterns

**Always use `-o json`** when capturing output to parse IDs or status
values. Table output is for human display only.

**Async operations** — `clusters create`, `databases create`,
`databases mcp deploy`, `databases rag deploy`, and `ingresses create`
return a `task_id`. Always follow with:

```bash
pgecloudctl tasks wait <task-id> -o json --timeout 600
```

Poll until `"status": "succeeded"` (exit 0) or `"status": "failed"`
(exit 1). Exit code 3 means timed out — increase `--timeout`.

**Delete commands** require `--yes` to skip confirmation prompts in
non-interactive use.

**All IDs are full UUIDs** — never truncate.

**Exit codes:** 0=success, 1=general error, 2=auth failure, 3=timeout,
4=not found. `tasks wait` uses 0/1/3 for succeeded/failed/timed-out.

---

## Layer 3: Workflow Intelligence

### Important notes (read before any workflow)

- Firewall rules, network CIDR/subnets, and node sizing are all set
  from the CLI: `clusters create` takes `--firewall-rule`, `--network`,
  and `--node` (repeatable structured flags); `clusters update` appends
  firewall rules and backup stores to an existing cluster. No UI visit
  or raw API call needed.
- Public clusters get external hostnames automatically; no ingress needed.
- Private clusters require an ingress + service registration before the
  service endpoint is reachable externally.
- `tasks wait` is the async bridge — always use it after any create or
  deploy command before proceeding.

---

### Workflow 1: Full Stack Setup

Provision a cluster, database, and optional AI services from scratch.

```
1. CHECK AUTH
   pgecloudctl auth status -o json
   → not authenticated? Tell user: run `pgecloudctl auth login`

2. CHECK BACKUP STORES
   pgecloudctl backup-stores list -o json
   → empty array? Ask user for: cloud account ID, region
     Then: pgecloudctl backup-stores create --name <n>
           --cloud-account-id <id> --region <r> -o json
   → non-empty? Use existing; capture id

3. GATHER INPUTS
   Ask user: cluster name, public or private, target region(s)
   (cloud-account-id from step 2 or: pgecloudctl cloud-accounts list -o json)

4. CREATE CLUSTER
   pgecloudctl clusters create --name <n> --cloud-account-id <id>
     --regions <r> --node-location <public|private>
     --backup-store-id <store-id>
     --node name=n1,instance-type=r7g.medium,volume-size=30
     --network cidr=10.4.0.0/16,public-subnets=10.4.1.0/24
     --firewall-rule name=postgres,port=5432,sources=0.0.0.0/0 -o json
   Private clusters: add private-subnets=... to --network.
   Pick a CIDR that does not overlap other clusters in the account.
   Capture task_id → tasks wait <task-id> --timeout 600
   Capture result cluster_id from task output

5. FIREWALL (append later if needed)
   Rules are set at create time (step 4). To add one afterwards:
   pgecloudctl clusters update <cluster-id>
     --firewall-rule name=https,port=443,sources=0.0.0.0/0
   (append semantics — existing rules are preserved)

6. CREATE DATABASE
   pgecloudctl databases create --name <n> --cluster-id <id> -o json
   Capture task_id → tasks wait <task-id> --timeout 600
   Capture result db_id from task output

7. OPTIONAL SERVICES — ask user: MCP, RAG, or neither?

   MCP path:
     Ask: embedding provider (openai|voyage|ollama), model, API key
     (if ollama: ask for ollama-url instead of api-key)
     pgecloudctl databases mcp deploy <db-id>
       --embedding-provider <p> --embedding-model <m>
       --embedding-api-key <k> -o json
     Capture task_id → tasks wait <task-id> --timeout 600

   RAG path:
     Ask: embedding provider/model/key, completion provider/model/key,
       pipeline.json path
     pgecloudctl databases rag deploy <db-id>
       --embedding-llm-provider <p> --embedding-llm-model <m>
       --embedding-llm-api-key <k>
       --completion-llm-provider <p> --completion-llm-model <m>
       --completion-llm-api-key <k>
       --pipeline-config pipeline.json -o json
     Capture task_id → tasks wait <task-id> --timeout 600

8. INGRESS (private clusters only — skip if node-location=public)
   pgecloudctl databases services list <db-id> -o json
   Capture service id
   pgecloudctl ingresses create --name <n> --cluster-id <id>
     --region <r> -o json
   Capture ingress id
   pgecloudctl ingresses services register <ingress-id>
     --database-id <db-id> --service-id <service-id>

9. VERIFY
   pgecloudctl databases services list <db-id>
   Expected: at least one service with status=active
```

Error handling:

| Exit | Meaning | Action |
|------|---------|--------|
| 2 | Auth failure | Run `auth login` |
| 3 | tasks wait timeout | Increase --timeout; check `tasks get <id>` |
| 4 | Resource not found | Verify IDs with list commands |
| 1 (task) | Async op failed | Read `error` field; delete resource and retry |

---

### Workflow 2: Deploy MCP

Deploy MCP onto an existing active database.

```
1. VERIFY DATABASE
   pgecloudctl databases get <db-id> -o json
   → status must be "active"; capture cluster_id
   → not active? Abort; tell user to wait or check task status

2. CHECK CLUSTER TYPE
   pgecloudctl clusters get <cluster-id> -o json
   → capture node_location (public|private)

3. DEPLOY MCP
   Ask: embedding provider, model, API key (or ollama-url if ollama)
   pgecloudctl databases mcp deploy <db-id>
     --embedding-provider <p> --embedding-model <m>
     --embedding-api-key <k> -o json
   Optional flags to ask about: --allow-writes (default: read-only),
     --target-nodes, --init-tokens, --init-users
   Capture task_id → tasks wait <task-id> --timeout 600

4. INGRESS (private only)
   → See expose-service workflow below

5. VERIFY
   pgecloudctl databases services list <db-id>
   Expected: MCP service with status=active
```

Common embedding providers: openai/text-embedding-3-small,
openai/text-embedding-3-large, voyage (with api-key), ollama (with url).

---

### Workflow 3: Deploy RAG

Deploy RAG (Ellie) onto an existing active database.

```
1. VERIFY DATABASE
   pgecloudctl databases get <db-id> -o json
   → status must be "active"; capture cluster_id
   → not active? Abort

2. CHECK CLUSTER TYPE
   pgecloudctl clusters get <cluster-id> -o json
   → capture node_location

3. PREPARE PIPELINE CONFIG
   Ask user for pipeline.json path (or help them create it).
   Minimum structure:
   {
     "pipelines": [{
       "name": "search",
       "system_prompt": "Answer using the provided context.",
       "hybrid_enabled": true,
       "vector_weight": 0.5,
       "tables": [{
         "table": "public.documents",
         "text_column": "content",
         "vector_column": "embedding"
       }]
     }]
   }

4. DEPLOY RAG
   Ask: embedding provider/model/key, completion provider/model/key
   pgecloudctl databases rag deploy <db-id>
     --embedding-llm-provider <p> --embedding-llm-model <m>
     --embedding-llm-api-key <k>
     --completion-llm-provider <p> --completion-llm-model <m>
     --completion-llm-api-key <k>
     --pipeline-config pipeline.json -o json
   Optional: --top-n, --token-budget, --target-nodes
   Capture task_id → tasks wait <task-id> --timeout 600

5. INGRESS (private only)
   → See expose-service workflow below

6. VERIFY
   pgecloudctl databases services list <db-id>
   Expected: RAG service with status=active
```

Common model combinations: embedding=openai/text-embedding-3-small,
completion=openai/gpt-4o or gpt-4o-mini.

---

### Workflow 4: Expose Service

Make a deployed service reachable on a private cluster.

```
1. GET SERVICE ID
   pgecloudctl databases services list <db-id> -o json
   Capture target service id

2. GET CLUSTER ID
   pgecloudctl databases get <db-id> -o json
   Capture cluster_id and primary region

3. CHECK/CREATE INGRESS
   pgecloudctl ingresses list -o json
   → ingress for this cluster already exists? Use its id; skip creation
   → none found:
     pgecloudctl ingresses create --name <n> --cluster-id <id>
       --region <r> -o json
     Capture id as ingress-id

4. REGISTER SERVICE
   pgecloudctl ingresses services register <ingress-id>
     --database-id <db-id> --service-id <service-id>
   Exit 0 = success. Repeat for each additional service.

5. VERIFY
   pgecloudctl ingresses services list <ingress-id>
   Expected: service with status=registered
```

---

### Workflow 5: Team Onboarding

Invite a new member and confirm membership.

```
1. CREATE INVITE (requires org admin role)
   pgecloudctl invites create --email user@example.com -o json
   Optional: --expiration <hours> (e.g. 72 for 3 days)
   Capture: id (invite-id), token, invite URL

2. SHARE
   Send invite URL to the new member.
   They can accept via the pgEdge Cloud UI or via CLI (step 3).

3. ACCEPT VIA CLI (optional — for scripted onboarding)
   pgecloudctl invites accept <invite-id> --token <token>
   Expected: status=accepted

4. VERIFY
   pgecloudctl memberships list
   Expected: new member's email with status=active
```

Error handling:

| Exit | Meaning | Action |
|------|---------|--------|
| 1 | Insufficient permissions | Confirm org admin role |
| 2 | Invalid email | Fix and retry |
| 3 | Invite not found | Check `invites list` |
| 4 | Invite expired | Create a new invite |
| 5 | Token already used | Re-issue invite; tokens are single-use |
