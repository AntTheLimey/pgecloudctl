package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/spf13/cobra"
)

// Cluster list flags.
var (
	clusterListLimit  int
	clusterListOffset int
)

// Cluster create flags.
var (
	clusterCreateName           string
	clusterCreateCloudAccountID string
	clusterCreateRegions        []string
	clusterCreateNodeLocation   string
	clusterCreateBackupStoreIDs []string
	clusterCreateFirewallRules  []string
)

// Cluster delete flags.
var (
	clusterDeleteYes   bool
	clusterDeleteForce bool
)

// Cluster update flags.
var (
	clusterUpdateFirewallRules  []string
	clusterUpdateBackupStoreIDs []string
	clusterUpdateRegions        []string
)

func init() {
	rootCmd.AddCommand(clustersCmd)
	clustersCmd.AddCommand(clustersListCmd)
	clustersCmd.AddCommand(clustersGetCmd)
	clustersCmd.AddCommand(clustersCreateCmd)
	clustersCmd.AddCommand(clustersDeleteCmd)

	// list flags
	clustersListCmd.Flags().IntVar(&clusterListLimit, "limit", 0,
		"Maximum number of results to return")
	clustersListCmd.Flags().IntVar(&clusterListOffset, "offset", 0,
		"Offset into the results for pagination")

	// create flags
	clustersCreateCmd.Flags().StringVar(&clusterCreateName, "name", "",
		"Cluster name")
	clustersCreateCmd.Flags().StringVar(&clusterCreateCloudAccountID,
		"cloud-account-id", "", "Cloud account ID")
	clustersCreateCmd.Flags().StringSliceVar(&clusterCreateRegions,
		"regions", nil, "Comma-separated list of regions")
	clustersCreateCmd.Flags().StringVar(&clusterCreateNodeLocation,
		"node-location", "", "Node location (public or private)")
	clustersCreateCmd.Flags().StringSliceVar(
		&clusterCreateBackupStoreIDs, "backup-store-id", nil,
		"Backup store ID to attach (repeatable; required to host a DB)")
	clustersCreateCmd.Flags().StringArrayVar(
		&clusterCreateFirewallRules, "firewall-rule", nil,
		"Firewall rule (repeatable). name must be one of "+
			"http, https, postgres, ssh. "+
			"e.g. name=https,port=443,sources=0.0.0.0/0")
	_ = clustersCreateCmd.MarkFlagRequired("name")
	_ = clustersCreateCmd.MarkFlagRequired("cloud-account-id")
	_ = clustersCreateCmd.MarkFlagRequired("regions")
	_ = clustersCreateCmd.MarkFlagRequired("node-location")
	addWaitFlags(clustersCreateCmd)

	// delete flags
	clustersDeleteCmd.Flags().BoolVarP(&clusterDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
	clustersDeleteCmd.Flags().BoolVar(&clusterDeleteForce, "force", false,
		"Also delete all databases and cloud infrastructure, "+
			"bypassing status and database-existence checks")
	addWaitFlags(clustersDeleteCmd)

	clustersCmd.AddCommand(clustersUpdateCmd)
	clustersUpdateCmd.Flags().StringArrayVar(&clusterUpdateFirewallRules,
		"firewall-rule", nil,
		"Firewall rule to append (repeatable). name must be one of "+
			"http, https, postgres, ssh. "+
			"e.g. name=https,port=443,sources=0.0.0.0/0")
	clustersUpdateCmd.Flags().StringSliceVar(&clusterUpdateBackupStoreIDs,
		"backup-store-id", nil,
		"Backup store ID to attach (repeatable)")
	clustersUpdateCmd.Flags().StringSliceVar(&clusterUpdateRegions,
		"regions", nil, "Replace the cluster's regions")
	addWaitFlags(clustersUpdateCmd)
}

var clustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Manage pgEdge Cloud clusters",
}

// --- list ---

var clustersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE:  runClustersList,
}

func runClustersList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	params := &api.ListClustersParams{}
	if clusterListLimit > 0 {
		params.Limit = &clusterListLimit
	}
	if clusterListOffset > 0 {
		params.Offset = &clusterListOffset
	}

	resp, err := client.ListClustersWithResponse(context.Background(), params)
	if err != nil {
		return fmt.Errorf("list clusters: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	clusters := resp.JSON200
	if clusters == nil || len(*clusters) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No clusters found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*clusters))
	for _, c := range *clusters {
		rows = append(rows, clusterRow{
			id:      c.Id,
			name:    c.Name,
			status:  c.Status,
			regions: joinStrings(c.Regions),
			created: formatTime(c.CreatedAt),
		})
	}

	headers := []string{"ID", "NAME", "STATUS", "REGIONS", "CREATED"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var clustersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get cluster details",
	Args:  cobra.ExactArgs(1),
	RunE:  runClustersGet,
}

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
	if err != nil {
		return fmt.Errorf("get cluster: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	c := resp.JSON200
	if c == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No cluster data returned.")
		return nil
	}

	rows := []output.Row{
		clusterRow{
			id:      c.Id,
			name:    c.Name,
			status:  c.Status,
			regions: joinStrings(c.Regions),
			created: formatTime(c.CreatedAt),
		},
	}

	headers := []string{"ID", "NAME", "STATUS", "REGIONS", "CREATED"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var clustersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	RunE:  runClustersCreate,
}

func runClustersCreate(cmd *cobra.Command, _ []string) error {
	if w := backupStoreWarning(clusterCreateBackupStoreIDs); w != "" {
		fmt.Fprintln(cmd.ErrOrStderr(), w)
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.CreateClusterJSONRequestBody{
		Name:           clusterCreateName,
		CloudAccountId: &clusterCreateCloudAccountID,
		Regions:        clusterCreateRegions,
		NodeLocation:   clusterCreateNodeLocation,
	}

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

	resp, err := client.CreateClusterWithResponse(context.Background(), body)
	if err != nil {
		return fmt.Errorf("create cluster: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	c := resp.JSON200
	if c == nil {
		// Accepted, but no body to read an id from — nothing to track.
		if flagOutput == "table" {
			fmt.Fprintln(cmd.OutOrStdout(),
				"Cluster created (no details returned).")
		}
		return nil
	}

	if flagOutput != "table" {
		if err := output.Print(cmd.OutOrStdout(), flagOutput, c, nil); err != nil {
			return err
		}
	} else {
		fmt.Fprintf(cmd.OutOrStdout(),
			"Cluster %q created (id: %s, status: %s).\n",
			c.Name, c.Id, c.Status)
	}
	return trackMutation(cmd, client, c.Id, "")
}

// --- delete ---

var clustersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  runClustersDelete,
}

func runClustersDelete(cmd *cobra.Command, args []string) error {
	prompt := fmt.Sprintf("Delete cluster %s? This cannot be undone.", args[0])
	if clusterDeleteForce {
		prompt = fmt.Sprintf(
			"Force-delete cluster %s AND all its databases and cloud "+
				"infrastructure? This cannot be undone.", args[0])
	}
	ok, err := confirmDestructive(cmd, clusterDeleteYes, prompt)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	id, err := resolveClusterID(context.Background(), client, args[0])
	if err != nil {
		return err
	}

	var priorTaskID string
	if waitFlag {
		priorTaskID, err = newestSubjectTaskID(
			context.Background(), client, id.String())
		if err != nil {
			return err
		}
	}

	params := &api.DeleteClusterParams{}
	if clusterDeleteForce {
		params.Force = &clusterDeleteForce
	}
	resp, err := client.DeleteClusterWithResponse(context.Background(), id, params)
	if err != nil {
		return fmt.Errorf("delete cluster: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Cluster %s deleted.\n", args[0])
	return trackMutation(cmd, client, id.String(), priorTaskID)
}

// --- update ---

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
	name := ""
	if rule.Name != nil {
		name = *rule.Name
	}
	switch {
	case name == "":
		return rule, fmt.Errorf(
			"firewall-rule: name is required (one of: %s)",
			validFirewallRuleNames)
	case !validFirewallRuleName(name):
		return rule, fmt.Errorf(
			"firewall-rule: name %q is not a valid rule type (one of: %s)",
			name, validFirewallRuleNames)
	}
	return rule, nil
}

// validFirewallRuleNames is the set of rule names the Cloud API accepts for a
// firewall rule (saas internal/starfleet/clusters/cluster_validator.go). The
// OpenAPI spec types this field as a free-form string with no enum, so the CLI
// hardcodes the set to give a clear client-side error instead of an opaque API
// 400. Keep in sync with the server; tracked by CLOUD spec bug.
const validFirewallRuleNames = "http, https, postgres, ssh"

func validFirewallRuleName(name string) bool {
	switch name {
	case "postgres", "https", "http", "ssh":
		return true
	default:
		return false
	}
}

// backupStoreWarning returns a caution when a cluster is being created with no
// backup store. A storeless cluster provisions fine but cannot host a database
// (database create fails: "at least 1 repository must be defined for provider:
// pgbackrest"), so it is a dead-end until a store is attached. Returns "" when
// at least one store is given. The CLI only warns — the API stays permissive so
// the create-then-attach flow (clusters update --backup-store-id) remains valid.
func backupStoreWarning(storeIDs []string) string {
	if len(storeIDs) > 0 {
		return ""
	}
	return "warning: cluster has no backup store; it cannot host a " +
		"database until one is attached " +
		"(--backup-store-id or `clusters update`)"
}

// appendStrPtr appends v to the slice behind p, allocating if p is nil.
func appendStrPtr(p *[]string, v string) *[]string {
	if p == nil {
		return &[]string{v}
	}
	*p = append(*p, v)
	return p
}

// --- row adapter ---

type clusterRow struct {
	id, name, status, regions, created string
}

func (r clusterRow) Columns() []string {
	return []string{r.id, r.name, output.ColorStatus(r.status), r.regions, r.created}
}
