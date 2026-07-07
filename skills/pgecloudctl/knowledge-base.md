# pgecloudctl Command Reference

## Authentication
Three sources (priority): env vars > flags > config file.
- Env: `PGEDGE_CLIENT_ID` + `PGEDGE_CLIENT_SECRET`
- Flags: `--client-id` + `--client-secret`
- Config: `~/.pgecloudctl/config.json` (written by `auth login`)

## Exit Codes
0=success, 1=general error, 2=auth failure, 3=timeout, 4=not found

## Global Flags
`--api-url` (default: https://api.pgedge.com), `-o/--output`
(table|json|yaml), `--no-color`, `-v/--verbose`, `--client-id`,
`--client-secret`

Always use `-o json` when capturing output for downstream processing.

## Destructive operations (`-y/--yes`)

All `delete` commands, plus `databases services remove` and `ingresses
services deregister`, prompt for confirmation. They take `-y/--yes` to skip
the prompt. In a non-interactive context (piped, or run by an agent — no
TTY) they **fail with exit 1 unless `--yes` is passed**, rather than hanging
or silently aborting. So scripts and agents must pass `--yes` to proceed.
Note `services remove` is irrecoverable — it discards the service's config
and credentials.

## Asynchronous operations (`--wait`)

Operations that spawn a background task — `create` and `delete` for
databases, clusters, ingresses, and backup-stores, plus the service
mutations (`mcp`/`rag deploy`/`update`, `services remove`) — accept
`--wait` (with `--timeout`, default 300, and `--interval`, default 5).
Without `--wait` the command returns as soon as the request is accepted,
not when the work completes. With `--wait` it blocks until the task reaches
a terminal state and the exit code reflects the outcome (0 succeeded,
1 failed, 3 timeout). In table output, the non-wait path prints the
`tasks list --subject-id <id>` command to monitor progress. Metadata
commands (ssh-keys, cloud-accounts, invites, memberships, cluster shares)
are synchronous and have no `--wait`.

---

## Commands

### auth
- `auth login` — interactive; no required flags
- `auth status` — no flags; output fields: AUTHENTICATED, SOURCE, EXPIRES
- `auth logout` — no flags; removes config.json + token.json

### cloud-accounts (alias: ca)
- `cloud-accounts list` — no required flags;
  output fields: id, name, type, status
- `cloud-accounts get <id>` — output fields: id, name, type, role_arn,
  status, description, created_at
- `cloud-accounts create` — required: `--name`, `--type` (aws|azure|gcp);
  AWS also requires `--role-arn`;
  GCP also requires `--project-id`, `--service-account`;
  Azure also requires `--azure-client-id`, `--azure-client-secret`,
  `--subscription-id`, `--tenant-id`;
  optional: `--description`, `--resource-group` (azure)
- `cloud-accounts delete <id>` — optional: `--yes`
- `cloud-accounts cloudformation-template` — prints AWS IAM CloudFormation
  template; no flags

### clusters
- `clusters list` — optional: `--limit`, `--offset`;
  output fields: id, name, status, regions
- `clusters get <id>` — output fields: id, name, status, node_location,
  regions, cloud_account_id, created_at
- `clusters create` — required: `--name`, `--cloud-account-id`,
  `--regions`, `--node-location` (public|private);
  optional: `--backup-store-id` (repeatable; required to host a DB),
  `--firewall-rule` (repeatable;
  `name=postgres,port=5432,sources=0.0.0.0/0`; name one of
  http|https|postgres|ssh),
  `--network` (repeatable, one per region;
  `region=...,cidr=10.4.0.0/16,`
  `public-subnets=10.4.1.0/24,private-subnets=10.4.128.0/24`),
  `--node` (repeatable;
  `name=n1,region=...,instance-type=r7g.medium,volume-size=30`),
  `--instance-type` + `--volume-size` (shorthand for --node, one node
  per region), `--wait`/`--timeout`/`--interval`;
  region= may be omitted on single-region clusters;
  response includes task_id
- `clusters update <id>` — optional: `--firewall-rule` (append),
  `--backup-store-id` (append), `--regions` (replace),
  `--wait`/`--timeout`/`--interval`; at least one flag required
- `clusters delete <id>` — optional: `--yes`, `--force` (cascade
  databases + infrastructure), `--wait`/`--timeout`/`--interval`

### clusters shares
- `clusters shares list <cluster-id>` — output fields: id, name, tenancy,
  capacity
- `clusters shares get <cluster-id> <share-id>` — output fields: id, name,
  tenancy, capacity, allowed_tenants
- `clusters shares create <cluster-id>` — required: `--name`, `--tenancy`
  (same|allowlist); optional: `--capacity`, `--allowed-tenants`
- `clusters shares delete <cluster-id> <share-id>` — optional: `--yes`

### databases
- `databases list` — optional: `--cluster-id`, `--limit`, `--offset`;
  output fields: id, name, pg_version, status, cluster_id
- `databases get <id>` — output fields: id, name, display_name, pg_version,
  status, cluster_id, created_at
- `databases create` — required: `--name`, `--cluster-id`;
  optional: `--pg-version` (default 16);
  response includes task_id
- `databases update <id>` — optional: `--display-name`, `--options`
- `databases delete <id>` — optional: `--yes`

### databases services

> **WARNING: Destructive API behavior.** The pgEdge Cloud API treats the
> `services` field as fully declarative — whatever you send REPLACES all
> existing services. The CLI uses a read-modify-write pattern to preserve
> existing services, but direct API callers must include all services in
> every update or risk destroying running services.

Service mutations (`mcp deploy/update`, `rag deploy/update`, `services
remove`) are asynchronous — they spawn a background task and return once the
request is accepted. Add `--wait` (with optional `--timeout`, default 300,
and `--interval`, default 5) to block until the task reaches a terminal
state; exit code is then 0 succeeded, 1 failed, 3 timeout. Without `--wait`,
monitor with `pgecloudctl tasks list --subject-id <db-id>`.

- `databases services list <db-id>` — output fields: id, type, status
- `databases services get <db-id> <service-id>` — output fields: id, type,
  status, endpoint
- `databases services remove <db-id> <type>` — type is mcp or rag;
  prompts for confirmation, `-y/--yes` to skip (required when non-interactive)

### databases mcp
- `databases mcp deploy <db-id>` — optional: `--embedding-provider`
  (ollama|openai|voyage), `--embedding-model`, `--embedding-api-key`,
  `--ollama-url` (required when provider=ollama), `--allow-writes`,
  `--target-nodes` (node names, e.g. n1,n2; auto-selects on single-node
  clusters; required on multi-node clusters), `--init-tokens`,
  `--init-users`;
  response includes task_id; output fields: id, type, status, endpoint
- `databases mcp update <db-id>` — same flags as deploy

### databases rag
- `databases rag deploy <db-id>` — optional: `--embedding-llm-provider`,
  `--embedding-llm-model`, `--embedding-llm-api-key`,
  `--completion-llm-provider`, `--completion-llm-model`,
  `--completion-llm-api-key`, `--pipeline-config` (path to a JSON file —
  either a bare array of pipelines or a `{"pipelines": [...]}` object),
  `--target-nodes` (node names, e.g. n1,n2; auto-selects on single-node
  clusters; required on multi-node clusters), `--top-n`, `--token-budget`;
  response includes task_id; output fields: id, type, status, endpoint
- `databases rag update <db-id>` — same flags as deploy

### backups
- `backups list` — optional: `--database-id`, `--kind`, `--created-after`,
  `--created-before`, `--limit`, `--offset`;
  output fields: id, name, kind, status, created_at
- `backups get <id>` — output fields: id, name, kind, status, database_id,
  provider, created_at
- `backups create` — required: `--database-id`; optional: `--provider`,
  `--name`, `--type`, `--target-nodes`
- `backups delete <id>` — optional: `--yes`
- `backups url <id>` — prints time-limited download URL; no flags

### backup-stores
- `backup-stores list` — optional: `--created-after`, `--created-before`,
  `--limit`, `--offset`;
  output fields: id, name, region, cloud_account_id
- `backup-stores get <id>` — output fields: id, name, region,
  cloud_account_id, created_at
- `backup-stores create` — required: `--name`, `--cloud-account-id`,
  `--region`
- `backup-stores delete <id>` — optional: `--yes`

### ssh-keys
- `ssh-keys list` — output fields: id, name, created_at
- `ssh-keys get <id>` — output fields: id, name, public_key, created_at
- `ssh-keys create` — required: `--name`, `--public-key`
- `ssh-keys delete <id>` — optional: `--yes`

### ingresses
- `ingresses list` — optional: `--created-after`, `--created-before`,
  `--limit`, `--offset`;
  output fields: id, name, region, cluster_id, status
- `ingresses get <id>` — output fields: id, name, region, cluster_id,
  status, hostname, created_at
- `ingresses create` — required: `--name`, `--cluster-id`, `--region`;
  response includes id
- `ingresses delete <id>` — optional: `--yes`

### ingresses services
- `ingresses services list <ingress-id>` — output fields: service_id,
  database_id, type
- `ingresses services register <ingress-id>` — required: `--database-id`,
  `--service-id`
- `ingresses services deregister <ingress-id> <service-id>` — no flags

### invites
- `invites list` — output fields: id, email, status, expires_at
- `invites get <id>` — output fields: id, email, status, expires_at,
  created_at
- `invites create` — required: `--email`; optional: `--expiration` (hours);
  response includes id, token
- `invites delete <id>` — optional: `--yes`
- `invites accept <id>` — required: `--token`

### memberships
- `memberships list` — output fields: id, email, role, joined_at
- `memberships delete <id>` — optional: `--yes`

### tasks
- `tasks list` — optional: `--status` (queued|running|succeeded|failed),
  `--subject-id`, `--subject-kind`, `--limit`, `--offset`;
  output fields: id, kind, status, subject_id, created_at
- `tasks get <id>` — output fields: id, kind, status, subject_id,
  subject_kind, created_at, updated_at
- `tasks wait <id>` — optional: `--timeout` (seconds, default 300),
  `--interval` (seconds, default 5);
  exit 0=succeeded, 1=failed, 3=timed out

### doctor
- `doctor` — no flags; runs 9 checks: Version, Latest version, Auth,
  API connectivity, Config, Environment, Shell, Install method, Skill;
  each check reports ok|warning|error

### llms
- `llms` — prints the complete embedded reference (llms-full.txt) for
  the installed version; no flags, no auth required

### skill
- `skill install` — optional: `--dir` (default
  ~/.claude/skills/pgecloudctl); installs SKILL.md + knowledge-base.md
  from the binary; idempotent, upgrades in place
