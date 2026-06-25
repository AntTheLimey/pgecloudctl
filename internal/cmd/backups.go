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

// Backup list flags.
var (
	backupListDatabaseID    string
	backupListKind          string
	backupListLimit         int
	backupListOffset        int
	backupListCreatedAfter  string
	backupListCreatedBefore string
)

// Backup create flags.
var (
	backupCreateDatabaseID  string
	backupCreateProvider    string
	backupCreateName        string
	backupCreateType        string
	backupCreateTargetNodes []string
)

// Backup delete flags.
var backupDeleteYes bool

func init() {
	rootCmd.AddCommand(backupsCmd)
	backupsCmd.AddCommand(backupsListCmd)
	backupsCmd.AddCommand(backupsGetCmd)
	backupsCmd.AddCommand(backupsCreateCmd)
	backupsCmd.AddCommand(backupsDeleteCmd)
	backupsCmd.AddCommand(backupsURLCmd)

	// list flags
	backupsListCmd.Flags().StringVar(&backupListDatabaseID, "database-id", "",
		"Filter backups to a specific database ID")
	backupsListCmd.Flags().StringVar(&backupListKind, "kind", "",
		"Filter backups to a specific kind")
	backupsListCmd.Flags().IntVar(&backupListLimit, "limit", 0,
		"Maximum number of results to return")
	backupsListCmd.Flags().IntVar(&backupListOffset, "offset", 0,
		"Offset into the results for pagination")
	backupsListCmd.Flags().StringVar(&backupListCreatedAfter,
		"created-after", "", "Filter: created after this RFC3339 timestamp")
	backupsListCmd.Flags().StringVar(&backupListCreatedBefore,
		"created-before", "", "Filter: created before this RFC3339 timestamp")

	// create flags
	backupsCreateCmd.Flags().StringVar(&backupCreateDatabaseID, "database-id", "",
		"Database ID to back up")
	backupsCreateCmd.Flags().StringVar(&backupCreateProvider, "provider", "",
		"Backup provider")
	backupsCreateCmd.Flags().StringVar(&backupCreateName, "name", "",
		"Optional backup name")
	backupsCreateCmd.Flags().StringVar(&backupCreateType, "type", "",
		"Optional backup type")
	backupsCreateCmd.Flags().StringSliceVar(&backupCreateTargetNodes, "target-nodes", nil,
		"Comma-separated list of target nodes")
	_ = backupsCreateCmd.MarkFlagRequired("database-id")
	_ = backupsCreateCmd.MarkFlagRequired("provider")

	// delete flags
	backupsDeleteCmd.Flags().BoolVarP(&backupDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
}

var backupsCmd = &cobra.Command{
	Use:   "backups",
	Short: "Manage pgEdge Cloud backups",
}

// --- list ---

var backupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backups",
	RunE:  runBackupsList,
}

func runBackupsList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	params := &api.ListBackupsParams{}
	if backupListDatabaseID != "" {
		id, err := uuid.Parse(backupListDatabaseID)
		if err != nil {
			return &ExitError{
				msg:  fmt.Sprintf("invalid database ID %q: %v", backupListDatabaseID, err),
				code: ExitGeneral,
			}
		}
		params.DatabaseId = &id
	}
	if backupListKind != "" {
		params.Kind = &backupListKind
	}
	if backupListLimit > 0 {
		params.Limit = &backupListLimit
	}
	if backupListOffset > 0 {
		params.Offset = &backupListOffset
	}
	if backupListCreatedAfter != "" {
		t, err := time.Parse(time.RFC3339, backupListCreatedAfter)
		if err != nil {
			return &ExitError{
				msg:  fmt.Sprintf("invalid --created-after timestamp: %v", err),
				code: ExitGeneral,
			}
		}
		params.CreatedAfter = &t
	}
	if backupListCreatedBefore != "" {
		t, err := time.Parse(time.RFC3339, backupListCreatedBefore)
		if err != nil {
			return &ExitError{
				msg:  fmt.Sprintf("invalid --created-before timestamp: %v", err),
				code: ExitGeneral,
			}
		}
		params.CreatedBefore = &t
	}

	resp, err := client.ListBackupsWithResponse(context.Background(), params)
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	backups := resp.JSON200
	if backups == nil || len(*backups) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No backups found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*backups))
	for _, b := range *backups {
		rows = append(rows, backupRow{
			id:         b.Id,
			databaseID: b.DatabaseId,
			name:       b.Name,
			status:     b.Status,
			kind:       b.Kind,
			size:       fmt.Sprintf("%d", b.Size),
			createdAt:  formatTime(b.CreatedAt),
		})
	}

	headers := []string{"ID", "DATABASE ID", "NAME", "STATUS", "KIND", "SIZE", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var backupsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get backup details",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupsGet,
}

func runBackupsGet(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid backup ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetBackupWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get backup: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	b := resp.JSON200
	if b == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No backup data returned.")
		return nil
	}

	rows := []output.Row{
		backupRow{
			id:         b.Id,
			databaseID: b.DatabaseId,
			name:       b.Name,
			status:     b.Status,
			kind:       b.Kind,
			size:       fmt.Sprintf("%d", b.Size),
			createdAt:  formatTime(b.CreatedAt),
		},
	}

	headers := []string{"ID", "DATABASE ID", "NAME", "STATUS", "KIND", "SIZE", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var backupsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new backup",
	RunE:  runBackupsCreate,
}

func runBackupsCreate(cmd *cobra.Command, _ []string) error {
	dbID, err := uuid.Parse(backupCreateDatabaseID)
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid database ID %q: %v", backupCreateDatabaseID, err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.BackupDatabaseJSONRequestBody{
		Provider: backupCreateProvider,
	}
	if backupCreateName != "" {
		body.Name = &backupCreateName
	}
	if backupCreateType != "" {
		body.Type = &backupCreateType
	}
	if len(backupCreateTargetNodes) > 0 {
		body.TargetNodes = &backupCreateTargetNodes
	}

	resp, err := client.BackupDatabaseWithResponse(context.Background(), dbID, body)
	if err != nil {
		return fmt.Errorf("create backup: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Backup initiated.")
	return nil
}

// --- delete ---

var backupsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a backup",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupsDelete,
}

func runBackupsDelete(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid backup ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	ok, err := confirmDestructive(cmd, backupDeleteYes,
		fmt.Sprintf("Delete backup %s? This cannot be undone.", args[0]))
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

	resp, err := client.DeleteBackupWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete backup: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Backup %s deleted.\n", args[0])
	return nil
}

// --- url ---

var backupsURLCmd = &cobra.Command{
	Use:   "url <id>",
	Short: "Get a download URL for a backup",
	Args:  cobra.ExactArgs(1),
	RunE:  runBackupsURL,
}

func runBackupsURL(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid backup ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetBackupLinkWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get backup URL: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	link := resp.JSON200
	if link == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No URL returned.")
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), link.Url)
	return nil
}

// --- row adapter ---

type backupRow struct {
	id, databaseID, name, status, kind, size, createdAt string
}

func (r backupRow) Columns() []string {
	return []string{
		r.id,
		r.databaseID,
		r.name,
		output.ColorStatus(r.status),
		r.kind,
		r.size,
		r.createdAt,
	}
}
