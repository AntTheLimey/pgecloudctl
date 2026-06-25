package cmd

import (
	"context"
	"fmt"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/spf13/cobra"
)

var dbServicesRemoveYes bool

func init() {
	databasesCmd.AddCommand(dbServicesCmd)
	dbServicesCmd.AddCommand(dbServicesListCmd)
	dbServicesCmd.AddCommand(dbServicesGetCmd)
	dbServicesCmd.AddCommand(dbServicesRemoveCmd)
	dbServicesRemoveCmd.Flags().BoolVarP(&dbServicesRemoveYes, "yes", "y",
		false, "Skip confirmation prompt")
	addServiceWaitFlags(dbServicesRemoveCmd)
}

var dbServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage services deployed alongside a database",
}

// --- list ---

var dbServicesListCmd = &cobra.Command{
	Use:   "list <db-id>",
	Short: "List services deployed on a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBServicesList,
}

func runDBServicesList(cmd *cobra.Command, args []string) error {
	_, db, err := fetchDatabase(args[0])
	if err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, db.Services, nil)
	}

	if db.Services == nil || len(*db.Services) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No services found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*db.Services))
	for _, svc := range *db.Services {
		rows = append(rows, serviceRow(svc))
	}

	headers := []string{"SERVICE ID", "TYPE", "STATE", "PORT", "DOMAIN"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var dbServicesGetCmd = &cobra.Command{
	Use:   "get <db-id> <service-id>",
	Short: "Get details of a specific service",
	Args:  cobra.ExactArgs(2),
	RunE:  runDBServicesGet,
}

func runDBServicesGet(cmd *cobra.Command, args []string) error {
	_, db, err := fetchDatabase(args[0])
	if err != nil {
		return err
	}

	svcID := args[1]

	if db.Services != nil {
		for _, svc := range *db.Services {
			if svc.ServiceId == svcID {
				if flagOutput != "table" {
					return output.Print(cmd.OutOrStdout(), flagOutput, &svc, nil)
				}
				rows := []output.Row{serviceRow(svc)}
				headers := []string{"SERVICE ID", "TYPE", "STATE", "PORT", "DOMAIN"}
				return output.Print(cmd.OutOrStdout(), "table", rows, headers)
			}
		}
	}

	return &ExitError{
		msg:  fmt.Sprintf("service %q not found on database %s", svcID, args[0]),
		code: ExitNotFound,
	}
}

// --- remove ---

var dbServicesRemoveCmd = &cobra.Command{
	Use:   "remove <db-id> <type>",
	Short: "Remove a service type from a database (mcp or rag)",
	Args:  cobra.ExactArgs(2),
	RunE:  runDBServicesRemove,
}

func runDBServicesRemove(cmd *cobra.Command, args []string) error {
	dbID := args[0]
	svcType := args[1]

	ok, err := confirmDestructive(cmd, dbServicesRemoveYes, fmt.Sprintf(
		"Remove %q service from database %s? This cannot be undone; its "+
			"configuration and credentials are unrecoverable.",
		svcType, dbID))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, db, err := fetchDatabase(dbID)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(dbID)
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid database ID %q: %v", dbID, err),
			code: ExitGeneral,
		}
	}

	remaining := make([]api.ServiceConfig, 0)
	if db.Services != nil {
		for _, svc := range *db.Services {
			if string(svc.ServiceType) != svcType {
				cfg := svcToConfig(svc)
				remaining = append(remaining, cfg)
			}
		}
	}

	svcs := nullable.NewNullableWithValue(remaining)
	body := api.UpdateDatabaseJSONRequestBody{
		Services: svcs,
	}

	var priorTaskID string
	if svcWait {
		priorTaskID, err = newestSubjectTaskID(context.Background(), client, dbID)
		if err != nil {
			return err
		}
	}

	resp, err := client.UpdateDatabaseWithResponse(context.Background(), id, body)
	if err != nil {
		return fmt.Errorf("remove service: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Service type %q removal requested for database %s.\n", svcType, dbID)
	return trackServiceMutation(cmd, client, dbID, priorTaskID)
}

// --- shared helpers ---

// fetchDatabaseWith retrieves a Database using an existing API client.
func fetchDatabaseWith(client *api.ClientWithResponses, idStr string) (*api.Database, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, &ExitError{
			msg:  fmt.Sprintf("invalid database ID %q: %v", idStr, err),
			code: ExitGeneral,
		}
	}

	resp, err := client.GetDatabaseWithResponse(
		context.Background(), id, &api.GetDatabaseParams{},
	)
	if err != nil {
		return nil, fmt.Errorf("get database: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &ExitError{
			msg:  fmt.Sprintf("database %q not found", idStr),
			code: ExitNotFound,
		}
	}

	return resp.JSON200, nil
}

// fetchDatabase creates an API client and retrieves a Database by ID string.
func fetchDatabase(idStr string) (*api.ClientWithResponses, *api.Database, error) {
	client, err := newAPIClient()
	if err != nil {
		return nil, nil, err
	}
	db, err := fetchDatabaseWith(client, idStr)
	if err != nil {
		return nil, nil, err
	}
	return client, db, nil
}

// buildServiceList returns a services slice that preserves all existing
// services whose type differs from newSvc.ServiceType, then appends newSvc.
// This implements the read-modify-write pattern for service updates.
func buildServiceList(db *api.Database, newSvc api.ServiceConfig) []api.ServiceConfig {
	result := make([]api.ServiceConfig, 0)
	if db.Services != nil {
		for _, svc := range *db.Services {
			if string(svc.ServiceType) != string(newSvc.ServiceType) {
				result = append(result, svcToConfig(svc))
			}
		}
	}
	result = append(result, newSvc)
	return result
}

func svcToConfig(svc api.Service) api.ServiceConfig {
	cfg := api.ServiceConfig{
		ServiceId:   &svc.ServiceId,
		ServiceType: api.ServiceConfigServiceType(svc.ServiceType),
		McpConfig:   svc.McpConfig,
		RagConfig:   svc.RagConfig,
		TargetNodes: svc.TargetNodes,
	}
	if hostID, err := svc.HostId.Get(); err == nil && hostID != "" {
		cfg.HostIds = &[]string{hostID}
	}
	return cfg
}

// --- row adapter ---

type svcRowData struct {
	id, typ, state, port, domain string
}

func (r svcRowData) Columns() []string {
	return []string{r.id, r.typ, output.ColorStatus(r.state), r.port, r.domain}
}

func serviceRow(svc api.Service) svcRowData {
	state := ""
	if v, err := svc.State.Get(); err == nil {
		state = v
	}

	port := ""
	if v, err := svc.Port.Get(); err == nil {
		port = fmt.Sprintf("%d", v)
	}

	domain := ""
	if v, err := svc.PublicDomain.Get(); err == nil && v != "" {
		domain = v
	} else if v, err := svc.PrivateDomain.Get(); err == nil && v != "" {
		domain = v
	}

	return svcRowData{
		id:     svc.ServiceId,
		typ:    string(svc.ServiceType),
		state:  state,
		port:   port,
		domain: domain,
	}
}
