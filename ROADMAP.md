# pgecloudctl — Roadmap

Items ranked by weighted score. Higher = do first.

**Score = (Impact × 2 + Urgency) / Effort**

### Impact

| Score | Meaning |
|-------|---------|
| 5 | Closes competitive gap or unlocks new user segment |
| 4 | Significant UX improvement for existing users |
| 3 | Useful but not blocking adoption |
| 2 | Nice-to-have, quality of life |
| 1 | Cosmetic or speculative |

### Urgency

| Score | Meaning |
|-------|---------|
| 5 | Blocks other work or missing from every competitor |
| 4 | Needed for first public release |
| 3 | Expected before recommending to customers |
| 2 | Can ship without it |
| 1 | No time pressure |

### Effort

| Size | Hours | Score |
|------|-------|-------|
| S | 1–4 | 1 |
| M | 4–16 | 2 |
| L | 16–40 | 3 |
| XL | 40+ | 4 |

---

| Item | Impact | Urgency | Effort | Score | Status | Notes |
|------|--------|---------|--------|-------|--------|-------|
| Auth (login/status/logout) | 5 | 5 | S (1) | 15.0 | Done | v0.1 — foundation for all commands |
| Clusters CRUD | 5 | 5 | M (2) | 7.5 | Done | v0.1 — core resource |
| Databases CRUD | 5 | 5 | M (2) | 7.5 | Done | v0.1 — core resource |
| Tasks list/get/wait | 5 | 5 | M (2) | 7.5 | Done | v0.1 — async bridge for AI agents |
| OpenAPI client generation | 4 | 5 | M (2) | 6.5 | Done | v0.1 — foundation for all API calls |
| Cloud accounts CRUD | 4 | 4 | M (2) | 6.0 | Done | v0.1 — needed to create clusters |
| MCP server deploy/update | 4 | 4 | M (2) | 6.0 | Done | v0.1 — key differentiator |
| RAG server deploy/update | 4 | 4 | M (2) | 6.0 | Done | v0.1 — key differentiator |
| Database services list/get/remove | 4 | 4 | S (1) | 12.0 | Done | v0.1 — service management |
| Table + JSON output | 5 | 5 | S (1) | 15.0 | Done | v0.1 — AI agents need JSON |
| CI + linting + goreleaser | 4 | 4 | M (2) | 6.0 | Done | v0.1 — pgEdge repo standards |
| Backups + backup stores | 3 | 3 | M (2) | 4.5 | Done | v0.2 |
| SSH keys | 3 | 2 | S (1) | 8.0 | Done | v0.2 |
| Ingresses + service registration | 4 | 3 | M (2) | 5.5 | Done | v0.2 — enterprise service exposure |
| Invites + memberships | 3 | 2 | M (2) | 4.0 | Done | v0.2 |
| Shares + CloudFormation template | 2 | 2 | S (1) | 6.0 | Done | v0.2 |
| YAML output format | 2 | 2 | S (1) | 6.0 | Done | v0.2 |
| --verbose HTTP tracing | 3 | 2 | S (1) | 8.0 | Done | v0.2 — debugging |
| Install script | 2 | 2 | S (1) | 6.0 | Done | v0.2 |
| UUID prefix matching | 3 | 2 | S (1) | 8.0 | Done | v0.4 — clusters/databases get/delete/update accept unique ID prefixes |
| Command-level tests | 3 | 3 | M (2) | 4.5 | Done | v0.2 — checkResponse, buildServiceList, wait loop |
| --no-color wiring + color output | 2 | 2 | S (1) | 6.0 | Done | v0.2 — flags declared but not yet functional |
| llms.txt | 4 | 3 | M (2) | 5.5 | Done | v0.3 — AI discoverability |
| Claude Code skill | 4 | 3 | M (2) | 5.5 | Done | v0.3 — Claude Code integration |
| AI workflow recipes | 3 | 2 | M (2) | 4.0 | Done | v0.3 — docs/guides for AI agents |
| pgecloudctl doctor | 3 | 2 | S (1) | 8.0 | Done | v0.3 — AI self-diagnostics |
| Multi-tenancy support | 4 | 3 | L (3) | 3.7 | Idea | Blocked on API changes |
| Clusters update command | 4 | 5 | S (1) | 13.0 | Done | v0.4 — read-modify-write; --firewall-rule/--backup-store-id/--regions |
| Cluster create parity (firewall + backup-store) | 4 | 5 | M (2) | 6.5 | Done | v0.4 — create now sends --firewall-rule + --backup-store-id; node/network flags deferred |
| ~~Databases create — backup config~~ | 4 | 5 | S (1) | 13.0 | Won't do | Misdiagnosis: a DB inherits its backup store from the cluster — there is no DB-level backup_store_id. The 400 was the cluster lacking backup_store_ids, fixed by cluster create/update parity above. |
| Re-vendor OpenAPI spec for PostgREST | 4 | 4 | S (1) | 12.0 | Idea | Blocks PostgREST cmds — vendored spec predates saas#1720; service_type enum is [mcp,rag] only, no PostgRESTServiceConfig |
| PostgREST service commands | 3 | 3 | M (2) | 4.5 | Idea | Blocked on spec re-vendor above; then mirror databases mcp/rag |
| MCP server for CLI | 3 | 2 | L (3) | 2.7 | Idea | Revisit if market demands it |

---

### Completed

**v0.1**

- Auth (login/status/logout)
- Clusters CRUD
- Databases CRUD
- Tasks list/get/wait
- OpenAPI client generation
- Cloud accounts CRUD
- MCP server deploy/update
- RAG server deploy/update
- Database services list/get/remove
- Table + JSON output
- CI + linting + goreleaser

**v0.2**

- Backups + backup stores
- SSH keys
- Ingresses + service registration
- Invites + memberships
- Shares + CloudFormation template
- YAML output format
- --verbose HTTP tracing
- Install script
- --no-color wiring + color output
- Command-level tests

**v0.3**

- llms.txt
- Claude Code skill
- AI workflow recipes
- pgecloudctl doctor

---

Post-v0.3 ideas tracked in the Idea rows above.
Multi-tenancy blocked on upstream API work.

---

### Active initiative — cluster parity + PostgREST

Found while testing PostgREST-as-a-service. Two distinct gap classes:
the **clusters** gaps below live in the Cobra command layer
(`internal/cmd`) only — the generated client (`internal/api`) already
exposes every operation and field needed. The **PostgREST** gaps are
deeper: they are blocked on the generated client, which has no
PostgREST support at all (see prerequisite below).

**Clusters create/update parity**

- `clusters create` ignores `CreateClusterInput.FirewallRules`,
  `.Networks` (cidr/subnets), `.Nodes` (instance_type, volume_size,
  volume_iops, availability_zone), and `.BackupStoreIds`. It only
  sends name, cloud_account_id, regions, node_location. Without
  `backup_store_ids`, the cluster can't host a database at all —
  `databases create` then fails `400 "backup store is not available
  to the cluster"`.
- No `clusters update` command exists, though
  `UpdateClusterWithResponse` (PATCH /v1/clusters/{id}) is generated.
  This blocked adding an `https` (:443) firewall rule to a public
  test cluster — had to fall back to raw curl.
- Open design question: how to express nested/repeated structures as
  flags — repeatable structured flags
  (`--firewall-rule name=https,port=443,sources=0.0.0.0/0`),
  convenience scalars for the single-network/single-node case, or a
  `--spec file.yaml` mapping to the input struct. Decide before
  building.
- Guardrail to encode: reject/warn on `volume_type: gp3` — it wedges
  later firewall-rule updates (rulemaster CLOUD-480).

**PostgREST service commands** (Priority 2, after integration proven)

- PREREQUISITE — re-vendor the spec. The vendored
  `openapi/pgedge.yaml` predates the PostgREST API (saas#1720). Its
  `service_type` enum is `[mcp, rag]` only and there is no
  `PostgRESTServiceConfig` schema, so the generated client carries
  zero PostgREST types (`grep postgrest internal/api/` is empty).
  Pull the updated saas OpenAPI, re-vendor `openapi/pgedge.yaml`, and
  run `make generate` BEFORE any command work — otherwise there is no
  `ServiceConfigServiceTypePostgrest` constant or config struct to
  build against.
- Add `databases postgrest deploy` / `update`, mirroring
  `databases mcp` and `databases rag`. Maps to the `services` array
  with `service_type: "postgrest"` and `postgrest_config`
  (db_schemas, db_anon_role, db_pool, max_rows, jwt_secret,
  jwt_audience, jwt_role_claim_key, cors_origins).
- Update `databases services remove` help text — it says
  "(mcp or rag)"; add postgrest.
- `services` array is REPLACE semantics: any verb must re-send
  existing services or be explicit that it replaces all.
