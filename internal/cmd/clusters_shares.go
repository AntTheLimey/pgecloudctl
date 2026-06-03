package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Cluster share create flags.
var (
	shareCreateName           string
	shareCreateCapacity       int
	shareCreateTenancy        string
	shareCreateAllowedTenants []string
)

// Cluster share delete flags.
var shareDeleteYes bool

func init() {
	clustersCmd.AddCommand(clusterSharesCmd)
	clusterSharesCmd.AddCommand(clusterSharesListCmd)
	clusterSharesCmd.AddCommand(clusterSharesGetCmd)
	clusterSharesCmd.AddCommand(clusterSharesCreateCmd)
	clusterSharesCmd.AddCommand(clusterSharesDeleteCmd)

	// create flags
	clusterSharesCreateCmd.Flags().StringVar(&shareCreateName, "name", "",
		"Share name")
	clusterSharesCreateCmd.Flags().IntVar(&shareCreateCapacity, "capacity", 0,
		"Share capacity")
	clusterSharesCreateCmd.Flags().StringVar(&shareCreateTenancy, "tenancy", "",
		"Tenancy mode: same or allowlist")
	clusterSharesCreateCmd.Flags().StringSliceVar(&shareCreateAllowedTenants,
		"allowed-tenants", nil, "Allowed tenant IDs (for allowlist tenancy)")

	// delete flags
	clusterSharesDeleteCmd.Flags().BoolVarP(&shareDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
}

var clusterSharesCmd = &cobra.Command{
	Use:   "shares",
	Short: "Manage cluster shares",
}

// --- list ---

var clusterSharesListCmd = &cobra.Command{
	Use:   "list <cluster-id>",
	Short: "List shares for a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  runClusterSharesList,
}

func runClusterSharesList(cmd *cobra.Command, args []string) error {
	clusterID, err := uuid.Parse(args[0])
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

	resp, err := client.ListClusterSharesWithResponse(context.Background(), clusterID)
	if err != nil {
		return fmt.Errorf("list cluster shares: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	shares := resp.JSON200
	if shares == nil || len(*shares) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No shares found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*shares))
	for _, s := range *shares {
		rows = append(rows, clusterShareRow{
			id:       s.Id,
			name:     s.Name,
			status:   s.Status,
			tenancy:  s.Tenancy,
			capacity: fmt.Sprintf("%d", s.Capacity),
			created:  formatTime(s.CreatedAt),
		})
	}

	headers := []string{"ID", "NAME", "STATUS", "TENANCY", "CAPACITY", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var clusterSharesGetCmd = &cobra.Command{
	Use:   "get <cluster-id> <share-id>",
	Short: "Get cluster share details",
	Args:  cobra.ExactArgs(2),
	RunE:  runClusterSharesGet,
}

func runClusterSharesGet(cmd *cobra.Command, args []string) error {
	clusterID, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid cluster ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	shareID, err := uuid.Parse(args[1])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid share ID %q: %v", args[1], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.ReadClusterShareWithResponse(
		context.Background(), clusterID, shareID,
	)
	if err != nil {
		return fmt.Errorf("get cluster share: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	s := resp.JSON200
	if s == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No share data returned.")
		return nil
	}

	rows := []output.Row{
		clusterShareRow{
			id:       s.Id,
			name:     s.Name,
			status:   s.Status,
			tenancy:  s.Tenancy,
			capacity: fmt.Sprintf("%d", s.Capacity),
			created:  formatTime(s.CreatedAt),
		},
	}

	headers := []string{"ID", "NAME", "STATUS", "TENANCY", "CAPACITY", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var clusterSharesCreateCmd = &cobra.Command{
	Use:   "create <cluster-id>",
	Short: "Create a cluster share",
	Args:  cobra.ExactArgs(1),
	RunE:  runClusterSharesCreate,
}

func runClusterSharesCreate(cmd *cobra.Command, args []string) error {
	clusterID, err := uuid.Parse(args[0])
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

	body := api.CreateClusterShareJSONRequestBody{}
	if shareCreateName != "" {
		body.Name = &shareCreateName
	}
	if shareCreateCapacity != 0 {
		body.Capacity = &shareCreateCapacity
	}
	if shareCreateTenancy != "" {
		tenancy := api.CreateClusterShareInputTenancy(shareCreateTenancy)
		body.Tenancy = &tenancy
	}
	if len(shareCreateAllowedTenants) > 0 {
		body.AllowedTenants = &shareCreateAllowedTenants
	}

	resp, err := client.CreateClusterShareWithResponse(
		context.Background(), clusterID, body,
	)
	if err != nil {
		return fmt.Errorf("create cluster share: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	s := resp.JSON200
	if s == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "Share created (no details returned).")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Share %q created (id: %s, status: %s).\n",
		s.Name, s.Id, s.Status)
	return nil
}

// --- delete ---

var clusterSharesDeleteCmd = &cobra.Command{
	Use:   "delete <cluster-id> <share-id>",
	Short: "Delete a cluster share",
	Args:  cobra.ExactArgs(2),
	RunE:  runClusterSharesDelete,
}

func runClusterSharesDelete(cmd *cobra.Command, args []string) error {
	clusterID, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid cluster ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	shareID, err := uuid.Parse(args[1])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid share ID %q: %v", args[1], err),
			code: ExitGeneral,
		}
	}

	if !shareDeleteYes {
		fmt.Fprintf(cmd.OutOrStdout(),
			"Delete share %s? This cannot be undone. [y/N]: ", args[1])
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.DeleteClusterShareWithResponse(
		context.Background(), clusterID, shareID,
	)
	if err != nil {
		return fmt.Errorf("delete cluster share: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Share %s deleted.\n", args[1])
	return nil
}

// --- row adapter ---

type clusterShareRow struct {
	id, name, status, tenancy, capacity, created string
}

func (r clusterShareRow) Columns() []string {
	return []string{
		r.id,
		r.name,
		output.ColorStatus(r.status),
		r.tenancy,
		r.capacity,
		r.created,
	}
}
