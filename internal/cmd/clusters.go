package cmd

import (
	"context"
	"fmt"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
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
)

// Cluster delete flags.
var clusterDeleteYes bool

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
	_ = clustersCreateCmd.MarkFlagRequired("name")
	_ = clustersCreateCmd.MarkFlagRequired("cloud-account-id")
	_ = clustersCreateCmd.MarkFlagRequired("regions")
	_ = clustersCreateCmd.MarkFlagRequired("node-location")
	addWaitFlags(clustersCreateCmd)

	// delete flags
	clustersDeleteCmd.Flags().BoolVarP(&clusterDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
	addWaitFlags(clustersDeleteCmd)
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

	resp, err := client.CreateClusterWithResponse(context.Background(), body)
	if err != nil {
		return fmt.Errorf("create cluster: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	c := resp.JSON200
	switch {
	case flagOutput != "table":
		if err := output.Print(cmd.OutOrStdout(), flagOutput, c, nil); err != nil {
			return err
		}
	case c == nil:
		fmt.Fprintln(cmd.OutOrStdout(), "Cluster created (no details returned).")
	default:
		fmt.Fprintf(cmd.OutOrStdout(),
			"Cluster %q created (id: %s, status: %s).\n",
			c.Name, c.Id, c.Status)
	}

	if c == nil {
		return nil
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
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid cluster ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	ok, err := confirmDestructive(cmd, clusterDeleteYes,
		fmt.Sprintf("Delete cluster %s? This cannot be undone.", args[0]))
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

	var priorTaskID string
	if waitFlag {
		priorTaskID, err = newestSubjectTaskID(
			context.Background(), client, id.String())
		if err != nil {
			return err
		}
	}

	params := &api.DeleteClusterParams{}
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

// --- row adapter ---

type clusterRow struct {
	id, name, status, regions, created string
}

func (r clusterRow) Columns() []string {
	return []string{r.id, r.name, output.ColorStatus(r.status), r.regions, r.created}
}
