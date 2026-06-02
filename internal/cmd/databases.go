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

// Database list flags.
var (
	dbListClusterID string
	dbListLimit     int
	dbListOffset    int
)

// Database create flags.
var (
	dbCreateName      string
	dbCreateClusterID string
	dbCreatePgVersion string
)

// Database update flags.
var (
	dbUpdateDisplayName string
	dbUpdateOptions     []string
)

// Database delete flags.
var dbDeleteYes bool

// databasesCmd is the parent command for all database subcommands. Other files
// in this package (e.g. database services commands) may attach their own
// subcommands to this var.
var databasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "Manage pgEdge Cloud databases",
}

func init() {
	rootCmd.AddCommand(databasesCmd)
	databasesCmd.AddCommand(databasesListCmd)
	databasesCmd.AddCommand(databasesGetCmd)
	databasesCmd.AddCommand(databasesCreateCmd)
	databasesCmd.AddCommand(databasesUpdateCmd)
	databasesCmd.AddCommand(databasesDeleteCmd)

	// list flags
	databasesListCmd.Flags().StringVar(&dbListClusterID, "cluster-id", "",
		"Filter by cluster ID")
	databasesListCmd.Flags().IntVar(&dbListLimit, "limit", 0,
		"Maximum number of results to return")
	databasesListCmd.Flags().IntVar(&dbListOffset, "offset", 0,
		"Offset into the results for pagination")

	// create flags
	databasesCreateCmd.Flags().StringVar(&dbCreateName, "name", "",
		"Database name")
	databasesCreateCmd.Flags().StringVar(&dbCreateClusterID, "cluster-id", "",
		"Cluster ID to deploy the database on")
	databasesCreateCmd.Flags().StringVar(&dbCreatePgVersion, "pg-version", "",
		"PostgreSQL version (e.g. 16)")
	_ = databasesCreateCmd.MarkFlagRequired("name")
	_ = databasesCreateCmd.MarkFlagRequired("cluster-id")

	// update flags
	databasesUpdateCmd.Flags().StringVar(&dbUpdateDisplayName, "display-name", "",
		"Display name for the database")
	databasesUpdateCmd.Flags().StringSliceVar(&dbUpdateOptions, "options", nil,
		"Comma-separated list of options")

	// delete flags
	databasesDeleteCmd.Flags().BoolVarP(&dbDeleteYes, "yes", "y", false,
		"Skip confirmation prompt")
}

// --- list ---

var databasesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List databases",
	RunE:  runDatabasesList,
}

func runDatabasesList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	params := &api.ListDatabasesParams{}
	if dbListClusterID != "" {
		parsed, err := uuid.Parse(dbListClusterID)
		if err != nil {
			return &exitError{
				msg:  fmt.Sprintf("invalid cluster ID %q: %v", dbListClusterID, err),
				code: ExitError,
			}
		}
		params.ClusterId = &parsed
	}
	if dbListLimit > 0 {
		params.Limit = &dbListLimit
	}
	if dbListOffset > 0 {
		params.Offset = &dbListOffset
	}

	resp, err := client.ListDatabasesWithResponse(context.Background(), params)
	if err != nil {
		return fmt.Errorf("list databases: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput == "json" {
		return output.Print(cmd.OutOrStdout(), "json", resp.JSON200, nil)
	}

	databases := resp.JSON200
	if databases == nil || len(*databases) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No databases found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*databases))
	for _, d := range *databases {
		rows = append(rows, databaseRow{
			id:        truncateID(d.Id),
			name:      d.Name,
			status:    d.Status,
			pgVersion: derefString(d.PgVersion),
			clusterID: truncateID(d.ClusterId),
			created:   formatTime(d.CreatedAt),
		})
	}

	headers := []string{"ID", "NAME", "STATUS", "PG VERSION", "CLUSTER", "CREATED"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var databasesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get database details",
	Args:  cobra.ExactArgs(1),
	RunE:  runDatabasesGet,
}

func runDatabasesGet(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &exitError{
			msg:  fmt.Sprintf("invalid database ID %q: %v", args[0], err),
			code: ExitError,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	params := &api.GetDatabaseParams{}
	resp, err := client.GetDatabaseWithResponse(context.Background(), id, params)
	if err != nil {
		return fmt.Errorf("get database: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput == "json" {
		return output.Print(cmd.OutOrStdout(), "json", resp.JSON200, nil)
	}

	d := resp.JSON200
	if d == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No database data returned.")
		return nil
	}

	rows := []output.Row{
		databaseRow{
			id:        truncateID(d.Id),
			name:      d.Name,
			status:    d.Status,
			pgVersion: derefString(d.PgVersion),
			clusterID: truncateID(d.ClusterId),
			created:   formatTime(d.CreatedAt),
		},
	}

	headers := []string{"ID", "NAME", "STATUS", "PG VERSION", "CLUSTER", "CREATED"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var databasesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new database",
	RunE:  runDatabasesCreate,
}

func runDatabasesCreate(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.CreateDatabaseJSONRequestBody{
		Name:      dbCreateName,
		ClusterId: &dbCreateClusterID,
	}
	if dbCreatePgVersion != "" {
		body.PgVersion = &dbCreatePgVersion
	}

	resp, err := client.CreateDatabaseWithResponse(context.Background(), body)
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput == "json" {
		return output.Print(cmd.OutOrStdout(), "json", resp.JSON200, nil)
	}

	d := resp.JSON200
	if d == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "Database created (no details returned).")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Database %q created (id: %s, status: %s).\n",
		d.Name, d.Id, d.Status)
	return nil
}

// --- update ---

var databasesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDatabasesUpdate,
}

func runDatabasesUpdate(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &exitError{
			msg:  fmt.Sprintf("invalid database ID %q: %v", args[0], err),
			code: ExitError,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.UpdateDatabaseJSONRequestBody{}
	if cmd.Flags().Changed("display-name") {
		body.DisplayName.Set(dbUpdateDisplayName)
	}
	if cmd.Flags().Changed("options") {
		body.Options = &dbUpdateOptions
	}

	resp, err := client.UpdateDatabaseWithResponse(context.Background(), id, body)
	if err != nil {
		return fmt.Errorf("update database: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput == "json" {
		return output.Print(cmd.OutOrStdout(), "json", resp.JSON200, nil)
	}

	d := resp.JSON200
	if d == nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Database %s updated.\n", args[0])
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Database %q updated (id: %s, status: %s).\n",
		d.Name, d.Id, d.Status)
	return nil
}

// --- delete ---

var databasesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDatabasesDelete,
}

func runDatabasesDelete(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &exitError{
			msg:  fmt.Sprintf("invalid database ID %q: %v", args[0], err),
			code: ExitError,
		}
	}

	if !dbDeleteYes {
		fmt.Fprintf(cmd.OutOrStdout(),
			"Delete database %s? This cannot be undone. [y/N]: ", args[0])
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

	resp, err := client.DeleteDatabaseWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete database: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Database %s deleted.\n", args[0])
	return nil
}

// --- row adapter ---

type databaseRow struct {
	id, name, status, pgVersion, clusterID, created string
}

func (r databaseRow) Columns() []string {
	return []string{r.id, r.name, r.status, r.pgVersion, r.clusterID, r.created}
}

// derefString returns the value of a string pointer, or "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
