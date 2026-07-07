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
| Cluster create parity (firewall + backup-store) | 4 | 5 | M (2) | 6.5 | Done | v0.4 — create now sends --firewall-rule + --backup-store-id; node/network flags shipped in v0.5 |
| ~~Databases create — backup config~~ | 4 | 5 | S (1) | 13.0 | Won't do | Misdiagnosis: a DB inherits its backup store from the cluster — there is no DB-level backup_store_id. The 400 was the cluster lacking backup_store_ids, fixed by cluster create/update parity above. |
| Clusters create node/network flags | 4 | 4 | M (2) | 6.0 | Done | v0.5 — --node / --network / --instance-type / --volume-size — PR #19 |
| Embed AI docs + make unavoidable | 4 | 3 | M (2) | 5.5 | Done | v0.5 — llms + skill install commands, doctor nudge, docs-sync build guard — PR #19 |
| Remove client-side gp3 rejection | 2 | 3 | S (1) | 7.0 | Done | CLOUD-480 fixed; gp3 now valid server-side — dropped the guardrail |
| Drop CLI firewall-enum stopgap | 2 | 2 | S (1) | 6.0 | Idea | Remove validFirewallRuleName once CLOUD-542 lands the enum in the spec + re-vendor |
| Re-vendor OpenAPI spec for PostgREST | 4 | 4 | S (1) | 12.0 | Idea | Blocks PostgREST cmds — vendored copy is pre-PostgREST (saas#1720); saas source already has PostgRESTServiceConfig, so it's a straight re-vendor |
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

**v0.4** (PR #17 — cluster/database parity, live-verified 2026-07-01)

- Clusters update command (read-modify-write; --firewall-rule /
  --backup-store-id / --regions)
- Cluster create parity — --backup-store-id + --firewall-rule
- Short UUID prefixes (clusters/databases get/delete/update)
- Clusters delete --force (cascade databases + infrastructure)
- Firewall-rule name validation (http/https/postgres/ssh)
- Storeless-cluster warning on clusters create

**v0.5** (PR #19 — AI docs + cluster create parity, merged 2026-07-07)

- Embed llms.txt / llms-full.txt and the Claude Code skill in the
  binary; `llms` and `skill install` commands
- `doctor` points agents at the bundled reference; the docs-sync build
  guard fails CI if a command or flag is undocumented
- Clusters create node/network flags — `--node`, `--network`,
  `--instance-type` / `--volume-size` shorthands
- install.sh installs the skill via the embedded `skill install`

---

Post-v0.5 ideas tracked in the Idea rows above.
Multi-tenancy blocked on upstream API work.

---

### Active initiative — PostgREST service commands

The PostgREST service is **live in pgEdge Cloud** — tested, and the JWT
issue that blocked the anon/JWT test matrix is fixed and verified
(2026-07-07). What's missing is **CLI support**: `pgecloudctl` still
can't deploy or manage a PostgREST service because its generated client
carries no PostgREST types.

Cluster create/update parity shipped across v0.4–v0.5; the deferred
node/network hardware flags landed in v0.5 (PR #19). CLOUD-480 (gp3
wedging later firewall-rule updates) is now fixed, so the CLI's
client-side gp3 rejection was removed. Remaining from that thread:

- The CLI hardcodes the firewall-rule name enum
  (http/https/postgres/ssh) because the spec types it as a free-form
  string. CLOUD-542 (expanded 2026-07-07) now also covers the same gap
  on `node_location` (public/private). Drop `validFirewallRuleName`
  once the spec carries the enum and the client is regenerated.
- A `--spec file.yaml` full-fidelity create mode remains an idea.

**Next up — PostgREST service commands**

- PREREQUISITE — re-vendor the OpenAPI spec (see note below). The
  vendored `openapi/pgedge.yaml` predates the PostgREST API (saas#1720):
  its `service_type` enum is `[mcp, rag]` and there is no
  `PostgRESTServiceConfig`, so the generated client carries zero
  PostgREST types (`grep postgrest internal/api/` is empty). The saas
  spec already has them, so this is a straight re-vendor: copy it over
  and run `make generate` BEFORE any command work — otherwise there is
  no `ServiceConfigServiceTypePostgrest` constant to build against.
- Add `databases postgrest deploy` / `update`, mirroring
  `databases mcp` and `databases rag`. Maps to the `services` array
  with `service_type: "postgrest"` and `postgrest_config`
  (db_schemas, db_anon_role, db_pool, max_rows, jwt_secret,
  jwt_audience, jwt_role_claim_key, cors_origins).
- Update `databases services remove` help text — it says
  "(mcp or rag)"; add postgrest.
- `services` array is REPLACE semantics: any verb must re-send
  existing services or be explicit that it replaces all.

**Where the vendored spec comes from.** `openapi/pgedge.yaml` is a
checked-in *copy*, not authored in this repo. The source of truth is
the saas project at `internal/starfleet/oapi/pgedge.yaml`.
`make generate` runs oapi-codegen against the local copy only; it never
fetches from saas. So "re-vendor" = copy the saas spec into
`openapi/pgedge.yaml`, then `make generate`. Because the spec is
authored in saas, any spec bug must be fixed in saas **first** and then
re-vendored here — a fix applied only to pgecloudctl's copy is
overwritten on the next re-vendor.
