package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Ingress list flags.
var (
	ingressListLimit         int
	ingressListOffset        int
	ingressListCreatedAfter  string
	ingressListCreatedBefore string
)

// Ingress create flags.
var (
	ingressCreateName      string
	ingressCreateClusterID string
	ingressCreateRegion    string
)

// Ingress delete flags.
var ingressDeleteYes bool

func init() {
	rootCmd.AddCommand(ingressesCmd)
	ingressesCmd.AddCommand(ingressesListCmd)
	ingressesCmd.AddCommand(ingressesGetCmd)
	ingressesCmd.AddCommand(ingressesCreateCmd)
	ingressesCmd.AddCommand(ingressesDeleteCmd)

	// list flags
	ingressesListCmd.Flags().IntVar(&ingressListLimit, "limit", 0,
		"Maximum number of results to return")
	ingressesListCmd.Flags().IntVar(&ingressListOffset, "offset", 0,
		"Offset into the results for pagination")
	ingressesListCmd.Flags().StringVar(&ingressListCreatedAfter,
		"created-after", "", "Filter: created after this RFC3339 timestamp")
	ingressesListCmd.Flags().StringVar(&ingressListCreatedBefore,
		"created-before", "", "Filter: created before this RFC3339 timestamp")

	// create flags
	ingressesCreateCmd.Flags().StringVar(&ingressCreateName, "name", "",
		"Ingress name")
	ingressesCreateCmd.Flags().StringVar(&ingressCreateClusterID,
		"cluster-id", "", "Cluster ID to associate with the ingress")
	ingressesCreateCmd.Flags().StringVar(&ingressCreateRegion,
		"region", "", "Cloud region for the ingress")
	_ = ingressesCreateCmd.MarkFlagRequired("name")
	_ = ingressesCreateCmd.MarkFlagRequired("cluster-id")
	_ = ingressesCreateCmd.MarkFlagRequired("region")

	// delete flags
	ingressesDeleteCmd.Flags().BoolVarP(&ingressDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
}

var ingressesCmd = &cobra.Command{
	Use:   "ingresses",
	Short: "Manage pgEdge Cloud ingresses",
}

// --- list ---

var ingressesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ingresses",
	RunE:  runIngressesList,
}

func runIngressesList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	params := &api.ListIngressesParams{}
	if ingressListLimit > 0 {
		params.Limit = &ingressListLimit
	}
	if ingressListOffset > 0 {
		params.Offset = &ingressListOffset
	}
	if ingressListCreatedAfter != "" {
		t, err := time.Parse(time.RFC3339, ingressListCreatedAfter)
		if err != nil {
			return &ExitError{
				msg:  fmt.Sprintf("invalid --created-after timestamp: %v", err),
				code: ExitGeneral,
			}
		}
		params.CreatedAfter = &t
	}
	if ingressListCreatedBefore != "" {
		t, err := time.Parse(time.RFC3339, ingressListCreatedBefore)
		if err != nil {
			return &ExitError{
				msg:  fmt.Sprintf("invalid --created-before timestamp: %v", err),
				code: ExitGeneral,
			}
		}
		params.CreatedBefore = &t
	}

	resp, err := client.ListIngressesWithResponse(context.Background(), params)
	if err != nil {
		return fmt.Errorf("list ingresses: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	ingresses := resp.JSON200
	if ingresses == nil || len(*ingresses) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No ingresses found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*ingresses))
	for _, ing := range *ingresses {
		rows = append(rows, ingressRow(ing))
	}

	headers := []string{"ID", "NAME", "STATUS", "CLUSTER ID", "REGION", "DOMAIN", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var ingressesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get ingress details",
	Args:  cobra.ExactArgs(1),
	RunE:  runIngressesGet,
}

func runIngressesGet(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid ingress ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetIngressWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get ingress: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	ing := resp.JSON200
	if ing == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No ingress data returned.")
		return nil
	}

	rows := []output.Row{ingressRow(*ing)}
	headers := []string{"ID", "NAME", "STATUS", "CLUSTER ID", "REGION", "DOMAIN", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var ingressesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new ingress",
	RunE:  runIngressesCreate,
}

func runIngressesCreate(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.CreateIngressJSONRequestBody{
		Name:      ingressCreateName,
		ClusterId: ingressCreateClusterID,
		Region:    ingressCreateRegion,
	}

	resp, err := client.CreateIngressWithResponse(context.Background(), body)
	if err != nil {
		return fmt.Errorf("create ingress: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	ing := resp.JSON200
	if ing == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "Ingress created (no details returned).")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Ingress %q created (id: %s, status: %s).\n",
		ing.Name, ing.Id, ing.Status)
	return nil
}

// --- delete ---

var ingressesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an ingress",
	Args:  cobra.ExactArgs(1),
	RunE:  runIngressesDelete,
}

func runIngressesDelete(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid ingress ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	ok, err := confirmDestructive(cmd, ingressDeleteYes,
		fmt.Sprintf("Delete ingress %s? This cannot be undone.", args[0]))
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

	resp, err := client.DeleteIngressWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete ingress: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Ingress %s deleted.\n", args[0])
	return nil
}

// --- row adapter ---

type ingressRowData struct {
	id, name, status, clusterID, region, domain, createdAt string
}

func (r ingressRowData) Columns() []string {
	return []string{
		r.id,
		r.name,
		output.ColorStatus(r.status),
		r.clusterID,
		r.region,
		r.domain,
		r.createdAt,
	}
}

func ingressRow(ing api.Ingress) ingressRowData {
	region := ""
	if ing.Region != nil {
		region = *ing.Region
	}

	domain := ""
	if ing.Domain != nil {
		domain = *ing.Domain
	}

	return ingressRowData{
		id:        ing.Id,
		name:      ing.Name,
		status:    ing.Status,
		clusterID: ing.ClusterId,
		region:    region,
		domain:    domain,
		createdAt: formatTime(ing.CreatedAt),
	}
}
