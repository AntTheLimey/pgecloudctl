# Cluster + database CLI parity — design

Date: 2026-06-30
Status: Approved (ready for implementation plan)

## Summary

Four roadmap items, picked off together because three of them share
one new piece of machinery (a structured-flag parser) and the fourth
is independent plumbing. All four are command-layer work in
`internal/cmd` — the generated client (`internal/api`) already exposes
every operation and field needed.

1. `clusters update <id>` — new command (PATCH /v1/clusters/{id}).
2. ~~`databases create --backup-store-id`~~ — REVERTED / removed from
   scope. A database inherits its backup store from the cluster; there
   is no DB-level backup-store input. See section 3.
3. Short UUID prefixes — clusters + databases `get`/`delete`/`update`.
4. `clusters create` parity — add `--backup-store-id` + `--firewall-rule`.

Out of scope this pass: node/network hardware flags on create,
short-ID resolution for resources other than clusters/databases, and
a `--spec` file mode (deferred; revisit if the AI-agent use case
demands full-fidelity input).

## Decisions

- **Flag style for nested structures: repeatable structured flags**
  (`--firewall-rule name=https,port=443,sources=0.0.0.0/0`). Each
  occurrence of the flag becomes one element in the array. Chosen over
  a `--spec` file because it fixes the real pain (no `clusters update`)
  with the least code and is discoverable in `--help`. `--spec` is a
  deliberate later item, not folded in here.
- **`clusters update` uses read-modify-write**, not raw replace. The
  PATCH body (`UpdateClusterInput`) has a non-optional `Regions`,
  which means the API replaces the whole cluster spec. To make "add
  one firewall rule" behave intuitively, the command GETs the cluster
  first, copies current state into the update body, layers the flag
  values on top, then PATCHes. The cluster GET response reuses the
  exact same `ClusterFirewallRuleSettings` / `Networks` / `Nodes`
  types as `UpdateClusterInput`, so the copy is near 1:1.
- **Repeatable flags append; they do not replace.** `--firewall-rule`
  and `--backup-store-id` add to whatever the cluster already has.
  Removing a rule is out of scope (the use case is adding one). If
  needed later, add an explicit `--firewall-rule-remove` rather than
  overloading append semantics.

## Components

### 1. Structured-flag parser (shared, new)

A small helper in `internal/cmd` that turns a repeatable
`key=value,key=value` string flag into a typed struct. Used by
`--firewall-rule` on both `clusters create` and `clusters update`.

- Registered as a `cobra` `StringArrayVar` so the flag can repeat.
- A parse function `parseFirewallRule(string)
  (api.ClusterFirewallRuleSettings, error)` splits on commas, then on
  the first `=` per pair. Recognized keys: `name`, `port`,
  `sources`, `prefix-lists`, `security-groups`. `port` parses to int;
  list-valued keys (`sources`, `prefix-lists`, `security-groups`)
  are repeated to accumulate values.
- **Separator (decided): repeat the key.** Values use `,` between
  pairs, so a list-valued key cannot also use `,` internally. Instead
  of inventing an inner separator, a list-valued key is repeated and
  accumulated — `sources=10.0.0.0/8,sources=192.168.0.0/16` appends
  both to that rule's `Sources`. No new separator to learn, and it
  reads cleanly. (Note: this means `,` always delimits pairs, never
  list items.)
- Unknown keys are a hard error naming the bad key and listing valid
  keys. Missing required `port` is a hard error.

The parser is the unit most worth testing in isolation: pure
string-in / struct-out, table-driven.

### 2. `clusters update <id>` (new command)

`internal/cmd/clusters.go`.

- Flags: `--firewall-rule` (repeatable structured), `--backup-store-id`
  (repeatable string), `--regions` (string slice, replaces regions),
  plus the standard `--wait` flags via `addWaitFlags`.
- Flow: resolve `<id>` (see component 4) → `GetClusterWithResponse`
  → build `UpdateClusterInput` from the current `Cluster` (regions,
  firewall_rules, networks, nodes, backup_store_ids, resource_tags,
  vpc_associations) → append parsed `--firewall-rule` values to
  firewall_rules → append `--backup-store-id` values to
  backup_store_ids → override regions if `--regions` given →
  `UpdateClusterWithResponse` → `checkResponse` → `trackMutation`.
- At least one mutating flag must be supplied; error otherwise so an
  empty update can't silently re-PATCH current state.

### 3. `databases create --backup-store-id` — REVERTED (not shipped)

> This section is retained for the record. It was implemented then
> reverted: a database **inherits** its backup store from its cluster,
> so there is no DB-level backup-store input. A storeless cluster
> simply can't host a database (create fails: "at least 1 repository
> must be defined for provider: pgbackrest"). The fix is the
> cluster-level `--backup-store-id` (sections above) plus the
> storeless-cluster warning on `clusters create`. Do not implement the
> below.

`internal/cmd/databases.go`.

- Today create sends no `Backups` block; the server's default lacks a
  repository, so create can 400. Add `--backup-store-id` (single
  string for now) that builds a `Backups` block with one repository
  referencing the store.
- **Open implementation detail (verify against live API in the
  plan):** the minimal valid `Backups` payload shape — specifically
  `Backups.Provider`, whether a `BackupConfig.Id` is required on
  create, and `BackupRepository.Type`. The struct path is
  `Backups{Provider, Config: [{Repositories: [{BackupStoreId}]}]}`.
  The plan must confirm required fields with a real create call (or a
  captured 200/400) before finalizing — do not guess. TDD this against
  a recorded response.
- When the flag is omitted, behavior is unchanged (no `Backups` block
  sent).

### 4. Short UUID prefix resolution (clusters + databases)

A resolver per resource type, applied wherever a command takes an
`<id>` positional arg (`get`, `delete`, `update`).

- `resolveClusterID(ctx, client, input) (uuid.UUID, error)` and
  `resolveDatabaseID(...)`.
- If `input` parses as a full UUID, return it directly (no API call).
- Otherwise list the resource (`ListClusters` / `ListDatabases`), keep
  IDs where `strings.HasPrefix(id, input)`, and require exactly one
  match. Zero matches → not-found error. More than one → ambiguous
  error listing the candidate IDs.
- Replaces the current direct `uuid.Parse(args[0])` calls in the
  clusters and databases `get`/`delete` commands, plus the new
  `update`.
- No minimum prefix length enforced (Docker-style); a 1-char prefix is
  legal if it happens to be unique.

## Error handling

- Parser errors (bad key, missing port, non-int port) return an
  `ExitError` with `ExitGeneral`, consistent with the existing invalid
  -ID handling in `clusters.go`.
- Ambiguous prefix returns an `ExitError` whose message lists the
  matching IDs so the user can disambiguate.
- All API responses continue through the existing `checkResponse`
  helper; no new error path.

## Testing

- `parseFirewallRule` — table-driven: happy path, multi-value sources,
  unknown key, missing port, non-int port, empty.
- `clusters update` — read-modify-write merge: given a fake cluster
  with one existing rule, a `--firewall-rule` appends rather than
  replaces; `--regions` overrides; empty update errors. Use the
  existing command-test pattern (see `clusters`/`backups` tests).
- Prefix resolver — unique match, no match, ambiguous match, full-UUID
  passthrough (no list call).
- `databases create --backup-store-id` — body carries the backup
  store id in the repository; omitting the flag sends no `Backups`.
- `make test` (race) and `make lint` must pass.

## Sequencing

The parser (1) lands first; `clusters update` (2) and `clusters
create` parity (4) both depend on it. The prefix resolver (4 in the
list / component 4 here) and `databases create` (3) are independent
and can land in any order.

## Roadmap impact

Closes roadmap items: "Clusters update command", "Cluster create
parity (firewall/networks/nodes)" (partial — firewall + backup-store
only), "Databases create — backup config", and "UUID prefix matching".
The deferred slices (node/network flags, `--spec`, broader short-ID
coverage) stay as Idea rows.
