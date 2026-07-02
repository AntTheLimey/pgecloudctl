# Cluster + database CLI parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `clusters update`, cluster create parity
(`--backup-store-id` + `--firewall-rule`), `databases create
--backup-store-id`, and Docker-style short UUID prefixes for clusters
and databases.

**Architecture:** All work is in the Cobra command layer
(`internal/cmd`); the generated client already exposes every operation
and field. Nested structures use repeatable structured flags. The
update command does read-modify-write (GET current cluster, append
flags, PATCH) so the API's replace-the-whole-spec semantics don't wipe
existing state. Pure helpers (parser, merge, prefix-match) are
extracted and unit-tested in isolation; commands wire them to IO.

**Tech Stack:** Go, Cobra, oapi-codegen client, `github.com/google/uuid`,
table-driven tests with `net/http/httptest`.

## Global Constraints

- gofmt mandatory; `make lint` (golangci-lint) must pass.
- `make test` runs with the race detector and must pass.
- Table-driven tests preferred.
- Tabs for Go indentation (gofmt default).
- No `panic` in production code.
- Conventional commit style: `feat:`, `fix:`, `test:`, etc. No
  self-attribution lines in commits.
- `--firewall-rule` MUST be registered with `StringArrayVar`, never
  `StringSliceVar` — the value contains commas and cobra's slice
  parser would split them and break pair grouping.

---

### Task 1: Firewall-rule flag parser

Pure string-to-struct parser used by both `clusters update` and
`clusters create`. List-valued keys (`sources`, `prefix-lists`,
`security-groups`) repeat the key to add elements; pairs are
comma-separated.

**Files:**
- Modify: `internal/cmd/clusters.go` (add `parseFirewallRule`,
  `appendStrPtr` near the bottom, before the row adapter)
- Test: `internal/cmd/clusters_test.go` (new file)

**Interfaces:**
- Produces: `parseFirewallRule(s string)
  (api.ClusterFirewallRuleSettings, error)` and
  `appendStrPtr(p *[]string, v string) *[]string`.

- [ ] **Step 1: Write the failing test**

Create `internal/cmd/clusters_test.go`:

```go
package cmd

import (
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
)

func strSlice(p *[]string) []string {
	if p == nil {
		return nil
	}
	return *p
}

func TestParseFirewallRule(t *testing.T) {
	name := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}

	t.Run("happy path with one source", func(t *testing.T) {
		r, err := parseFirewallRule("name=https,port=443,sources=0.0.0.0/0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name(r.Name) != "https" {
			t.Errorf("name = %q, want https", name(r.Name))
		}
		if r.Port != 443 {
			t.Errorf("port = %d, want 443", r.Port)
		}
		if got := strSlice(r.Sources); len(got) != 1 || got[0] != "0.0.0.0/0" {
			t.Errorf("sources = %v, want [0.0.0.0/0]", got)
		}
	})

	t.Run("repeated key accumulates sources", func(t *testing.T) {
		r, err := parseFirewallRule("port=22,sources=10.0.0.0/8,sources=192.168.0.0/16")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := strSlice(r.Sources); len(got) != 2 {
			t.Errorf("sources = %v, want 2 elements", got)
		}
	})

	errCases := []struct {
		name string
		in   string
	}{
		{"missing port", "name=https,sources=0.0.0.0/0"},
		{"non-int port", "port=https"},
		{"unknown key", "port=443,colour=red"},
		{"not key=value", "port=443,bogus"},
	}
	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseFirewallRule(tc.in); err == nil {
				t.Errorf("parseFirewallRule(%q) = nil error, want error", tc.in)
			}
		})
	}
}

var _ = api.ClusterFirewallRuleSettings{} // keep api import if unused above
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/ -run TestParseFirewallRule -v`
Expected: FAIL — `undefined: parseFirewallRule`.

- [ ] **Step 3: Write minimal implementation**

In `internal/cmd/clusters.go`, add these imports if missing
(`strconv`, `strings`) and add the functions before
`type clusterRow struct`:

```go
// parseFirewallRule parses a repeatable structured flag value of the
// form "name=https,port=443,sources=0.0.0.0/0" into a
// ClusterFirewallRuleSettings. Pairs are comma-separated; list-valued
// keys (sources, prefix-lists, security-groups) are repeated to add
// elements. port is required.
func parseFirewallRule(s string) (api.ClusterFirewallRuleSettings, error) {
	var rule api.ClusterFirewallRuleSettings
	portSet := false
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			return rule, fmt.Errorf(
				"firewall-rule: %q is not key=value", pair)
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		switch k {
		case "name":
			name := v
			rule.Name = &name
		case "port":
			p, err := strconv.Atoi(v)
			if err != nil {
				return rule, fmt.Errorf(
					"firewall-rule: port %q is not an integer", v)
			}
			rule.Port = p
			portSet = true
		case "sources":
			rule.Sources = appendStrPtr(rule.Sources, v)
		case "prefix-lists":
			rule.PrefixLists = appendStrPtr(rule.PrefixLists, v)
		case "security-groups":
			rule.SecurityGroups = appendStrPtr(rule.SecurityGroups, v)
		default:
			return rule, fmt.Errorf(
				"firewall-rule: unknown key %q (valid: name, port, "+
					"sources, prefix-lists, security-groups)", k)
		}
	}
	if !portSet {
		return rule, fmt.Errorf("firewall-rule: port is required")
	}
	return rule, nil
}

// appendStrPtr appends v to the slice behind p, allocating if p is nil.
func appendStrPtr(p *[]string, v string) *[]string {
	if p == nil {
		return &[]string{v}
	}
	*p = append(*p, v)
	return p
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cmd/ -run TestParseFirewallRule -v`
Expected: PASS (all subtests).

- [ ] **Step 5: Commit**

```bash
gofmt -w internal/cmd/clusters.go internal/cmd/clusters_test.go
git add internal/cmd/clusters.go internal/cmd/clusters_test.go
git commit -m "feat: add firewall-rule structured flag parser"
```

---

### Task 2: Short UUID prefix resolver

Resolve a full UUID or a unique prefix to a cluster/database UUID. Pure
match logic is extracted so it can be tested without IO; the per-resource
resolvers wrap a List call.

**Files:**
- Create: `internal/cmd/resolve.go`
- Test: `internal/cmd/resolve_test.go`

**Interfaces:**
- Produces:
  - `resolveIDPrefix(input string, ids []string, kind string)
    (string, error)`
  - `resolveClusterID(ctx context.Context, client
    *api.ClientWithResponses, input string) (uuid.UUID, error)`
  - `resolveDatabaseID(ctx context.Context, client
    *api.ClientWithResponses, input string) (uuid.UUID, error)`

- [ ] **Step 1: Write the failing test**

Create `internal/cmd/resolve_test.go`:

```go
package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
)

func TestResolveIDPrefix(t *testing.T) {
	ids := []string{
		"a1b2c3d4-0000-0000-0000-000000000001",
		"a1ffffff-0000-0000-0000-000000000002",
		"b9000000-0000-0000-0000-000000000003",
	}
	t.Run("unique prefix", func(t *testing.T) {
		got, err := resolveIDPrefix("b9", ids, "cluster")
		if err != nil || got != ids[2] {
			t.Fatalf("got (%q, %v), want (%q, nil)", got, err, ids[2])
		}
	})
	t.Run("ambiguous prefix", func(t *testing.T) {
		if _, err := resolveIDPrefix("a1", ids, "cluster"); err == nil {
			t.Errorf("expected ambiguous error")
		}
	})
	t.Run("no match", func(t *testing.T) {
		if _, err := resolveIDPrefix("zz", ids, "cluster"); err == nil {
			t.Errorf("expected not-found error")
		}
	})
}

func TestResolveClusterID(t *testing.T) {
	full := "a1b2c3d4-1111-2222-3333-444455556666"
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.Cluster{{Id: full}})
	}

	t.Run("full uuid skips API", func(t *testing.T) {
		// nil client proves no list call is made for a full UUID.
		got, err := resolveClusterID(context.Background(), nil, full)
		if err != nil || got.String() != full {
			t.Fatalf("got (%q, %v), want (%q, nil)", got, err, full)
		}
	})

	t.Run("prefix resolves via list", func(t *testing.T) {
		client := newTestClient(t, handler)
		got, err := resolveClusterID(context.Background(), client, "a1b2")
		if err != nil || got.String() != full {
			t.Fatalf("got (%q, %v), want (%q, nil)", got, err, full)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/ -run 'TestResolveIDPrefix|TestResolveClusterID' -v`
Expected: FAIL — `undefined: resolveIDPrefix` /
`undefined: resolveClusterID`.

- [ ] **Step 3: Write minimal implementation**

Create `internal/cmd/resolve.go`:

```go
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/google/uuid"
)

// resolveIDPrefix returns the single ID in ids equal to or prefixed by
// input. Errors if input matches zero or more than one ID. kind names
// the resource for error messages (e.g. "cluster").
func resolveIDPrefix(input string, ids []string, kind string) (string, error) {
	var matches []string
	for _, id := range ids {
		if strings.HasPrefix(id, input) {
			matches = append(matches, id)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", &ExitError{
			msg:  fmt.Sprintf("no %s matches ID prefix %q", kind, input),
			code: ExitNotFound,
		}
	default:
		return "", &ExitError{
			msg: fmt.Sprintf("ambiguous %s ID prefix %q matches: %s",
				kind, input, strings.Join(matches, ", ")),
			code: ExitGeneral,
		}
	}
}

// resolveClusterID returns the cluster UUID for a full UUID or a unique
// ID prefix. A full UUID is returned without any API call.
func resolveClusterID(ctx context.Context, client *api.ClientWithResponses,
	input string) (uuid.UUID, error) {
	if id, err := uuid.Parse(input); err == nil {
		return id, nil
	}
	resp, err := client.ListClustersWithResponse(ctx, &api.ListClustersParams{})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("list clusters: %w", err)
	}
	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return uuid.UUID{}, err
	}
	var ids []string
	if resp.JSON200 != nil {
		for _, c := range *resp.JSON200 {
			ids = append(ids, c.Id)
		}
	}
	matched, err := resolveIDPrefix(input, ids, "cluster")
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(matched)
}

// resolveDatabaseID returns the database UUID for a full UUID or a
// unique ID prefix. A full UUID is returned without any API call.
func resolveDatabaseID(ctx context.Context, client *api.ClientWithResponses,
	input string) (uuid.UUID, error) {
	if id, err := uuid.Parse(input); err == nil {
		return id, nil
	}
	resp, err := client.ListDatabasesWithResponse(ctx, &api.ListDatabasesParams{})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("list databases: %w", err)
	}
	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return uuid.UUID{}, err
	}
	var ids []string
	if resp.JSON200 != nil {
		for _, d := range *resp.JSON200 {
			ids = append(ids, d.Id)
		}
	}
	matched, err := resolveIDPrefix(input, ids, "database")
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(matched)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cmd/ -run 'TestResolveIDPrefix|TestResolveClusterID' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -w internal/cmd/resolve.go internal/cmd/resolve_test.go
git add internal/cmd/resolve.go internal/cmd/resolve_test.go
git commit -m "feat: add short UUID prefix resolver for clusters/databases"
```

---

### Task 3: Wire resolver into get/delete commands

Replace the direct `uuid.Parse(args[0])` calls in the clusters and
databases `get`/`delete` commands with the resolver. Each call site
must move the parse to *after* `newAPIClient()` because the resolver
needs a client.

**Files:**
- Modify: `internal/cmd/clusters.go` (`runClustersGet` ~line 135,
  `runClustersDelete` ~line 243)
- Modify: `internal/cmd/databases.go` (`runDatabasesGet` ~line 160,
  `runDatabasesDelete` ~line 271; leave `runDatabasesUpdate` ~line 327
  for this task too)

**Interfaces:**
- Consumes: `resolveClusterID`, `resolveDatabaseID` from Task 2.

- [ ] **Step 1: Update `runClustersGet`**

Current (clusters.go ~135):

```go
func runClustersGet(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid cluster ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetClusterWithResponse(context.Background(), id)
```

Replace the opening with:

```go
func runClustersGet(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	id, err := resolveClusterID(context.Background(), client, args[0])
	if err != nil {
		return err
	}

	resp, err := client.GetClusterWithResponse(context.Background(), id)
```

- [ ] **Step 2: Update `runClustersDelete`**

Apply the same transform at clusters.go ~243: move `client, err :=
newAPIClient()` above the ID resolution, and replace the `uuid.Parse`
block with `id, err := resolveClusterID(context.Background(), client,
args[0])`. Keep the existing `confirmDestructive` call — it uses
`args[0]` for the prompt text, which is fine.

- [ ] **Step 3: Update databases get/delete/update**

In databases.go, apply the same transform at ~160 (`runDatabasesGet`),
~271 (`runDatabasesDelete`), and ~327 (`runDatabasesUpdate`): create
the client first, then `id, err := resolveDatabaseID(
context.Background(), client, args[0])`, replacing each
`uuid.Parse(args[0])` block.

- [ ] **Step 4: Verify build and existing tests pass**

Run: `go build ./... && go test ./internal/cmd/ -v`
Expected: build succeeds; all existing tests PASS. If `uuid` becomes an
unused import in either file, remove it (gofmt/goimports).

- [ ] **Step 5: Commit**

```bash
gofmt -w internal/cmd/clusters.go internal/cmd/databases.go
git add internal/cmd/clusters.go internal/cmd/databases.go
git commit -m "feat: accept short UUID prefixes on cluster/db get and delete"
```

---

### Task 4: `clusters update` command

New command doing read-modify-write. `buildClusterUpdate` is the pure
merge function (existing state + appended flags); `runClustersUpdate`
wires GET → merge → PATCH → task tracking.

**Files:**
- Modify: `internal/cmd/clusters.go` (flag vars, `init`, command var,
  `runClustersUpdate`, `buildClusterUpdate`)
- Test: `internal/cmd/clusters_test.go`

**Interfaces:**
- Consumes: `parseFirewallRule` (Task 1), `resolveClusterID` (Task 2),
  `addWaitFlags`, `newestSubjectTaskID`, `trackMutation` (existing).
- Produces: `buildClusterUpdate(c *api.Cluster, addRules
  []api.ClusterFirewallRuleSettings, addStoreIDs, regions []string)
  api.UpdateClusterInput`.

- [ ] **Step 1: Write the failing test**

Append to `internal/cmd/clusters_test.go`:

```go
func TestBuildClusterUpdate(t *testing.T) {
	existingRule := api.ClusterFirewallRuleSettings{Port: 22}
	c := &api.Cluster{
		Regions:        []string{"us-east-1"},
		FirewallRules:  &[]api.ClusterFirewallRuleSettings{existingRule},
		BackupStoreIds: &[]string{"bs-existing"},
	}
	newRule := api.ClusterFirewallRuleSettings{Port: 443}

	t.Run("appends rules and stores, keeps regions", func(t *testing.T) {
		in := buildClusterUpdate(c,
			[]api.ClusterFirewallRuleSettings{newRule},
			[]string{"bs-new"}, nil)
		if in.FirewallRules == nil || len(*in.FirewallRules) != 2 {
			t.Fatalf("firewall rules = %v, want 2", in.FirewallRules)
		}
		if in.BackupStoreIds == nil || len(*in.BackupStoreIds) != 2 {
			t.Fatalf("backup stores = %v, want 2", in.BackupStoreIds)
		}
		if len(in.Regions) != 1 || in.Regions[0] != "us-east-1" {
			t.Errorf("regions = %v, want [us-east-1]", in.Regions)
		}
	})

	t.Run("regions override when supplied", func(t *testing.T) {
		in := buildClusterUpdate(c, nil, nil, []string{"eu-west-1"})
		if len(in.Regions) != 1 || in.Regions[0] != "eu-west-1" {
			t.Errorf("regions = %v, want [eu-west-1]", in.Regions)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/ -run TestBuildClusterUpdate -v`
Expected: FAIL — `undefined: buildClusterUpdate`.

- [ ] **Step 3: Implement `buildClusterUpdate`**

In `internal/cmd/clusters.go`, add before the row adapter:

```go
// buildClusterUpdate produces an UpdateClusterInput from the cluster's
// current state, then layers on requested changes. Firewall rules and
// backup store IDs are appended to existing values; regions replace
// only when supplied. Fresh slices are built to avoid aliasing the
// cluster's own slices.
func buildClusterUpdate(c *api.Cluster,
	addRules []api.ClusterFirewallRuleSettings,
	addStoreIDs, regions []string) api.UpdateClusterInput {
	in := api.UpdateClusterInput{
		Regions:         c.Regions,
		Networks:        c.Networks,
		Nodes:           c.Nodes,
		ResourceTags:    c.ResourceTags,
		VpcAssociations: c.VpcAssociations,
	}

	rules := []api.ClusterFirewallRuleSettings{}
	if c.FirewallRules != nil {
		rules = append(rules, *c.FirewallRules...)
	}
	rules = append(rules, addRules...)
	if len(rules) > 0 {
		in.FirewallRules = &rules
	}

	stores := []string{}
	if c.BackupStoreIds != nil {
		stores = append(stores, *c.BackupStoreIds...)
	}
	stores = append(stores, addStoreIDs...)
	if len(stores) > 0 {
		in.BackupStoreIds = &stores
	}

	if len(regions) > 0 {
		in.Regions = regions
	}
	return in
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cmd/ -run TestBuildClusterUpdate -v`
Expected: PASS.

- [ ] **Step 5: Add flags, command, and runner**

In `internal/cmd/clusters.go`, add to the flag var block:

```go
// Cluster update flags.
var (
	clusterUpdateFirewallRules  []string
	clusterUpdateBackupStoreIDs []string
	clusterUpdateRegions        []string
)
```

In `init()`, after the delete flags, register the command and flags
(note `StringArrayVar` for `--firewall-rule`):

```go
	clustersCmd.AddCommand(clustersUpdateCmd)
	clustersUpdateCmd.Flags().StringArrayVar(&clusterUpdateFirewallRules,
		"firewall-rule", nil,
		"Firewall rule to append, e.g. "+
			"name=https,port=443,sources=0.0.0.0/0 (repeatable)")
	clustersUpdateCmd.Flags().StringSliceVar(&clusterUpdateBackupStoreIDs,
		"backup-store-id", nil,
		"Backup store ID to attach (repeatable)")
	clustersUpdateCmd.Flags().StringSliceVar(&clusterUpdateRegions,
		"regions", nil, "Replace the cluster's regions")
	addWaitFlags(clustersUpdateCmd)
```

Add the command var and runner (place after `clustersDeleteCmd` /
`runClustersDelete`):

```go
var clustersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a cluster (append firewall rules / backup stores)",
	Args:  cobra.ExactArgs(1),
	RunE:  runClustersUpdate,
}

func runClustersUpdate(cmd *cobra.Command, args []string) error {
	if len(clusterUpdateFirewallRules) == 0 &&
		len(clusterUpdateBackupStoreIDs) == 0 &&
		len(clusterUpdateRegions) == 0 {
		return &ExitError{
			msg: "clusters update: specify at least one of " +
				"--firewall-rule, --backup-store-id, --regions",
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	id, err := resolveClusterID(context.Background(), client, args[0])
	if err != nil {
		return err
	}

	rules := make([]api.ClusterFirewallRuleSettings, 0,
		len(clusterUpdateFirewallRules))
	for _, raw := range clusterUpdateFirewallRules {
		r, err := parseFirewallRule(raw)
		if err != nil {
			return &ExitError{msg: err.Error(), code: ExitGeneral}
		}
		rules = append(rules, r)
	}

	getResp, err := client.GetClusterWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get cluster: %w", err)
	}
	if err := checkResponse(getResp.StatusCode(),
		string(getResp.Body)); err != nil {
		return err
	}
	if getResp.JSON200 == nil {
		return &ExitError{msg: "cluster not found", code: ExitNotFound}
	}

	body := buildClusterUpdate(getResp.JSON200, rules,
		clusterUpdateBackupStoreIDs, clusterUpdateRegions)

	var priorTaskID string
	if waitFlag {
		priorTaskID, err = newestSubjectTaskID(
			context.Background(), client, id.String())
		if err != nil {
			return err
		}
	}

	resp, err := client.UpdateClusterWithResponse(
		context.Background(), id, body)
	if err != nil {
		return fmt.Errorf("update cluster: %w", err)
	}
	if err := checkResponse(resp.StatusCode(),
		string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s updated.\n", id)
	return trackMutation(cmd, client, id.String(), priorTaskID)
}
```

- [ ] **Step 6: Verify build + full package tests**

Run: `go build ./... && go test ./internal/cmd/ -v`
Expected: build succeeds; all tests PASS.

- [ ] **Step 7: Commit**

```bash
gofmt -w internal/cmd/clusters.go internal/cmd/clusters_test.go
git add internal/cmd/clusters.go internal/cmd/clusters_test.go
git commit -m "feat: add clusters update command (read-modify-write)"
```

---

### Task 5: `clusters create` parity flags

Add `--backup-store-id` and `--firewall-rule` to `clusters create`,
reusing the Task 1 parser.

**Files:**
- Modify: `internal/cmd/clusters.go` (flag vars, `init`,
  `runClustersCreate` ~line 188)

**Interfaces:**
- Consumes: `parseFirewallRule` (Task 1).

- [ ] **Step 1: Add flag vars**

In the create flag var block, add:

```go
	clusterCreateBackupStoreIDs []string
	clusterCreateFirewallRules  []string
```

- [ ] **Step 2: Register flags in `init`**

After the existing create flag registrations and before the
`MarkFlagRequired` calls:

```go
	clustersCreateCmd.Flags().StringSliceVar(
		&clusterCreateBackupStoreIDs, "backup-store-id", nil,
		"Backup store ID to attach (repeatable; required to host a DB)")
	clustersCreateCmd.Flags().StringArrayVar(
		&clusterCreateFirewallRules, "firewall-rule", nil,
		"Firewall rule, e.g. "+
			"name=https,port=443,sources=0.0.0.0/0 (repeatable)")
```

- [ ] **Step 3: Populate the request body**

In `runClustersCreate`, after the `body := api.CreateClusterJSONRequestBody{...}`
literal and before `client.CreateClusterWithResponse`:

```go
	if len(clusterCreateBackupStoreIDs) > 0 {
		body.BackupStoreIds = &clusterCreateBackupStoreIDs
	}
	for _, raw := range clusterCreateFirewallRules {
		r, err := parseFirewallRule(raw)
		if err != nil {
			return &ExitError{msg: err.Error(), code: ExitGeneral}
		}
		if body.FirewallRules == nil {
			body.FirewallRules = &[]api.ClusterFirewallRuleSettings{}
		}
		*body.FirewallRules = append(*body.FirewallRules, r)
	}
```

- [ ] **Step 4: Write a test for the body-building**

The create runner does IO, but the flag-to-body mapping is the same
parser already covered in Task 1. Add a focused test that the create
flags produce the expected body fields by exercising `parseFirewallRule`
through the same path. Append to `clusters_test.go`:

```go
func TestClusterCreateFirewallParsing(t *testing.T) {
	// Mirrors the loop in runClustersCreate.
	raw := []string{"name=https,port=443,sources=0.0.0.0/0"}
	var rules []api.ClusterFirewallRuleSettings
	for _, s := range raw {
		r, err := parseFirewallRule(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rules = append(rules, r)
	}
	if len(rules) != 1 || rules[0].Port != 443 {
		t.Fatalf("rules = %v, want one rule on port 443", rules)
	}
}
```

- [ ] **Step 5: Verify build + tests**

Run: `go build ./... && go test ./internal/cmd/ -v`
Expected: build succeeds; all tests PASS.

- [ ] **Step 6: Commit**

```bash
gofmt -w internal/cmd/clusters.go internal/cmd/clusters_test.go
git add internal/cmd/clusters.go internal/cmd/clusters_test.go
git commit -m "feat: add --backup-store-id and --firewall-rule to clusters create"
```

---

### Task 6: `databases create --backup-store-id` — REVERTED, do not implement

> **Do not implement this task.** It was built then reverted: a
> database inherits its backup store from its cluster, so there is no
> DB-level backup-store input. The real fix is the cluster-level
> `--backup-store-id` (Task 5) plus a storeless-cluster warning on
> `clusters create`. The steps below are retained only for the record.

Add `--backup-store-id` to `databases create`, building a `Backups`
block with one repository referencing the store. The exact minimal
payload (provider value, whether the server tolerates an empty
`BackupConfig.Id`) is the one unknown — the unit test pins the CLI's
output, and a manual verification step confirms server acceptance.

**Files:**
- Modify: `internal/cmd/databases.go` (flag var, `init`,
  `runDatabasesCreate` ~line 216, add `buildDatabaseBackups`)
- Test: `internal/cmd/databases_create_test.go` (new file)

**Interfaces:**
- Produces: `buildDatabaseBackups(storeID string) *api.Backups`.

- [ ] **Step 1: Write the failing test**

Create `internal/cmd/databases_create_test.go`:

```go
package cmd

import (
	"testing"
)

func TestBuildDatabaseBackups(t *testing.T) {
	b := buildDatabaseBackups("bs-abc123")
	if b == nil {
		t.Fatal("buildDatabaseBackups returned nil")
	}
	if b.Config == nil || len(*b.Config) != 1 {
		t.Fatalf("config = %v, want 1 entry", b.Config)
	}
	repos := (*b.Config)[0].Repositories
	if repos == nil || len(*repos) != 1 {
		t.Fatalf("repositories = %v, want 1 entry", repos)
	}
	got := (*repos)[0].BackupStoreId
	if got == nil || *got != "bs-abc123" {
		t.Fatalf("backup_store_id = %v, want bs-abc123", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmd/ -run TestBuildDatabaseBackups -v`
Expected: FAIL — `undefined: buildDatabaseBackups`.

- [ ] **Step 3: Implement `buildDatabaseBackups`**

In `internal/cmd/databases.go`, add near the create command:

```go
// buildDatabaseBackups builds a Backups block with a single repository
// pointing at the given backup store. NOTE: Provider is set to
// "pgbackrest" and BackupConfig.Id is left empty pending live-API
// verification (see plan Task 6, Step 7) — adjust if the API rejects
// these.
func buildDatabaseBackups(storeID string) *api.Backups {
	return &api.Backups{
		Provider: "pgbackrest",
		Config: &[]api.BackupConfig{{
			Repositories: &[]api.BackupRepository{{
				BackupStoreId: &storeID,
			}},
		}},
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cmd/ -run TestBuildDatabaseBackups -v`
Expected: PASS.

- [ ] **Step 5: Add the flag and wire into the runner**

In `internal/cmd/databases.go`, add to the create flag var block:

```go
	dbCreateBackupStoreID string
```

Register in `init` (with the other `databasesCreateCmd` flags):

```go
	databasesCreateCmd.Flags().StringVar(&dbCreateBackupStoreID,
		"backup-store-id", "",
		"Backup store ID to use for this database's backups")
```

In `runDatabasesCreate`, after the `if dbCreatePgVersion != "" {...}`
block and before `client.CreateDatabaseWithResponse`:

```go
	if dbCreateBackupStoreID != "" {
		body.Backups = buildDatabaseBackups(dbCreateBackupStoreID)
	}
```

- [ ] **Step 6: Verify build + tests**

Run: `go build ./... && go test ./internal/cmd/ -v`
Expected: build succeeds; all tests PASS.

- [ ] **Step 7: Live-API verification (manual, REQUIRED)**

The unit test only proves the CLI sends the store ID. Confirm the
server accepts the payload before considering this done:

```bash
go run ./cmd/pgecloudctl databases create \
  --name verify-backups --cluster-id <cluster-with-store> \
  --backup-store-id <store-id> --verbose
```

Expected: 2xx and a database id. If it returns 400, inspect the body:
- Wrong `Provider` value → set the correct one in
  `buildDatabaseBackups` and re-run the unit test.
- Complaint about `BackupConfig.Id` (empty) → either populate it from
  the cluster's config or confirm the server generates it; adjust
  `buildDatabaseBackups` accordingly.
Record the working payload shape in a commit message.

- [ ] **Step 8: Commit**

```bash
gofmt -w internal/cmd/databases.go internal/cmd/databases_create_test.go
git add internal/cmd/databases.go internal/cmd/databases_create_test.go
git commit -m "feat: add --backup-store-id to databases create"
```

---

### Task 7: Final verification + help-text sweep

**Files:**
- Modify: `internal/cmd/databases_services.go:104` (help text, if not
  already updated) — confirm it still reads sensibly; PostgREST is out
  of scope here, so no change required unless it's stale.

- [ ] **Step 1: Run the full suite with race + lint**

Run: `make test && make lint`
Expected: all tests PASS under `-race`; golangci-lint reports no issues.

- [ ] **Step 2: Smoke-test the new help output**

Run:
```bash
go run ./cmd/pgecloudctl clusters update --help
go run ./cmd/pgecloudctl clusters create --help
go run ./cmd/pgecloudctl databases create --help
```
Expected: `--firewall-rule`, `--backup-store-id`, `--regions` appear
with the documented usage strings.

- [ ] **Step 3: Commit any lint fixes**

```bash
git add -A
git commit -m "chore: lint and help-text cleanup for cluster/db parity"
```

---

## Notes for the implementer

- `api.UUID` (the client's id parameter type) is satisfied by
  `github.com/google/uuid`.`UUID`; existing code passes `uuid.Parse`
  results directly, so the resolver returning `uuid.UUID` is correct.
- The cluster GET response (`api.Cluster`) reuses the exact
  `ClusterFirewallRuleSettings` / `Networks` / `Nodes` types as
  `UpdateClusterInput`, which is why `buildClusterUpdate` copies them
  directly.
- Node/network hardware flags, a `--spec` file mode, and short-ID
  resolution for resources beyond clusters/databases are intentionally
  out of scope (see the design doc).
