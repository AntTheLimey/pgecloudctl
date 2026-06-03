package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Backup store list flags.
var (
	backupStoreListLimit         int
	backupStoreListOffset        int
	backupStoreListCreatedAfter  string
	backupStoreListCreatedBefore string
)

// Backup store create flags.
var (
	backupStoreCreateName           string
	backupStoreCreateCloudAccountID string
	backupStoreCreateRegion         string
)

// Backup store delete flags.
var backupStoreDeleteYes bool

func init() {
	rootCmd.AddCommand(backupStoresCmd)
	backupStoresCmd.AddCommand(backupStoresListCmd)
	backupStoresCmd.AddCommand(backupStoresGetCmd)
	backupStoresCmd.AddCommand(backupStoresCreateCmd)
	backupStoresCmd.AddCommand(backupStoresDeleteCmd)

	// list flags
	backupStoresListCmd.Flags().IntVar(&backupStoreListLimit, "limit", 0,
		"Maximum number of results to return")
	backupStoresListCmd.Flags().IntVar(&backupStoreListOffset, "offset", 0,
		"Offset into the results for pagination")
	backupStoresListCmd.Flags().StringVar(&backupStoreListCreatedAfter,
		"created-after", "", "Filter: created after this RFC3339 timestamp")
	backupStoresListCmd.Flags().StringVar(&backupStoreListCreatedBefore,
		"created-before", "", "Filter: created before this RFC3339 timestamp")

	// create flags
	backupStoresCreateCmd.Flags().StringVar(&backupStoreCreateName, "name", "",
		"Backup store name")
	backupStoresCreateCmd.Flags().StringVar(&backupStoreCreateCloudAccountID,
		"cloud-account-id", "", "Cloud account ID")
	backupStoresCreateCmd.Flags().StringVar(&backupStoreCreateRegion, "region", "",
		"Region for the backup store")
	_ = backupStoresCreateCmd.MarkFlagRequired("name")
	_ = backupStoresCreateCmd.MarkFlagRequired("cloud-account-id")

	// delete flags
	backupStoresDeleteCmd.Flags().BoolVarP(&backupStoreDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
}

var backupStoresCmd = &cobra.Command{
	Use:   "backup-stores",
	Short: "Manage pgEdge Cloud backup stores",
}

// --- list ---

var backupStoresListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backup stores",
	RunE:  runBackupStoresList,
}

func runBackupStoresList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	params := &api.ListBackupStoresParams{}
	if backupStoreListLimit > 0 {
		params.Limit = &backupStoreListLimit
	}
	if backupStoreListOffset > 0 {
		params.Offset = &backupStoreListOffset
	}
	if backupStoreListCreatedAfter != "" {
		t, err := time.Parse(time.RFC3339, backupStoreListCreatedAfter)
		if err != nil {
			return &ExitError{
				msg:  fmt.Sprintf("invalid --created-after timestamp: %v", err),
				code: ExitGeneral,
			}
		}
		params.CreatedAfter = &t
	}
	if backupStoreListCreatedBefore != "" {
		t, err := time.Parse(time.RFC3339, backupStoreListCreatedBefore)
		if err != nil {
			return &ExitError{
				msg:  fmt.Sprintf("invalid --created-before timestamp: %v", err),
				code: ExitGeneral,
			}
		}
		params.CreatedBefore = &t
	}

	resp, err := client.ListBackupStoresWithResponse(context.Background(), params)
	if err != nil {
		return fmt.Errorf("list backup stores: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	stores := resp.JSON200
	if stores == nil || len(*stores) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No backup stores found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*stores))
	for _, s := range *stores {
		rows = append(rows, backupStoreRow{
			id:             s.Id,
			name:           s.Name,
			status:         s.Status,
			cloudAccountID: s.CloudAccountId,
			createdAt:      formatTime(s.CreatedAt),
		})
	}

	headers := []string{"ID", "NAME", "STATUS", "CLOUD ACCOUNT ID", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var backupStoresGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get backup store details",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupStoresGet,
}

func runBackupStoresGet(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid backup store ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetBackupStoreWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get backup store: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	s := resp.JSON200
	if s == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No backup store data returned.")
		return nil
	}

	rows := []output.Row{
		backupStoreRow{
			id:             s.Id,
			name:           s.Name,
			status:         s.Status,
			cloudAccountID: s.CloudAccountId,
			createdAt:      formatTime(s.CreatedAt),
		},
	}

	headers := []string{"ID", "NAME", "STATUS", "CLOUD ACCOUNT ID", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var backupStoresCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new backup store",
	RunE:  runBackupStoresCreate,
}

func runBackupStoresCreate(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.CreateBackupStoreJSONRequestBody{
		Name:           backupStoreCreateName,
		CloudAccountId: backupStoreCreateCloudAccountID,
	}
	if backupStoreCreateRegion != "" {
		body.Region = &backupStoreCreateRegion
	}

	resp, err := client.CreateBackupStoreWithResponse(context.Background(), body)
	if err != nil {
		return fmt.Errorf("create backup store: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	s := resp.JSON200
	if s == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "Backup store created (no details returned).")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Backup store %q created (id: %s, status: %s).\n",
		s.Name, s.Id, s.Status)
	return nil
}

// --- delete ---

var backupStoresDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a backup store",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupStoresDelete,
}

func runBackupStoresDelete(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid backup store ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	if !backupStoreDeleteYes {
		fmt.Fprintf(cmd.OutOrStdout(),
			"Delete backup store %s? This cannot be undone. [y/N]: ", args[0])
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

	resp, err := client.DeleteBackupStoreWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete backup store: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Backup store %s deleted.\n", args[0])
	return nil
}

// --- row adapter ---

type backupStoreRow struct {
	id, name, status, cloudAccountID, createdAt string
}

func (r backupStoreRow) Columns() []string {
	return []string{
		r.id,
		r.name,
		output.ColorStatus(r.status),
		r.cloudAccountID,
		r.createdAt,
	}
}
