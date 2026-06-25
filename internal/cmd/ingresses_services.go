package cmd

import (
	"context"
	"fmt"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/spf13/cobra"
)

// Ingress service register flags.
var (
	ingressSvcRegisterDatabaseID string
	ingressSvcRegisterServiceID  string
)

// Ingress service deregister flags.
var ingressSvcDeregisterYes bool

func init() {
	ingressesCmd.AddCommand(ingressServicesCmd)
	ingressServicesCmd.AddCommand(ingressServicesListCmd)
	ingressServicesCmd.AddCommand(ingressServicesRegisterCmd)
	ingressServicesCmd.AddCommand(ingressServicesDeregisterCmd)

	// register flags
	ingressServicesRegisterCmd.Flags().StringVar(&ingressSvcRegisterDatabaseID,
		"database-id", "", "Database ID to register")
	ingressServicesRegisterCmd.Flags().StringVar(&ingressSvcRegisterServiceID,
		"service-id", "", "Service ID to expose")
	_ = ingressServicesRegisterCmd.MarkFlagRequired("database-id")
	_ = ingressServicesRegisterCmd.MarkFlagRequired("service-id")

	ingressServicesDeregisterCmd.Flags().BoolVarP(&ingressSvcDeregisterYes,
		"yes", "y", false, "Skip confirmation prompt")
}

var ingressServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage services registered on an ingress",
}

// --- list ---

var ingressServicesListCmd = &cobra.Command{
	Use:   "list <ingress-id>",
	Short: "List services registered on an ingress",
	Args:  cobra.ExactArgs(1),
	RunE:  runIngressServicesList,
}

func runIngressServicesList(cmd *cobra.Command, args []string) error {
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

	resp, err := client.ListIngressServicesWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("list ingress services: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	svcs := resp.JSON200
	if svcs == nil || len(*svcs) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No services found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*svcs))
	for _, svc := range *svcs {
		rows = append(rows, ingressServiceRow(svc))
	}

	headers := []string{"SERVICE ID", "DATABASE ID", "URL"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- register ---

var ingressServicesRegisterCmd = &cobra.Command{
	Use:   "register <ingress-id>",
	Short: "Register a service on an ingress",
	Args:  cobra.ExactArgs(1),
	RunE:  runIngressServicesRegister,
}

func runIngressServicesRegister(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid ingress ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	dbID, err := uuid.Parse(ingressSvcRegisterDatabaseID)
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid database ID %q: %v", ingressSvcRegisterDatabaseID, err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.CreateIngressServiceJSONRequestBody{
		DatabaseId: openapi_types.UUID(dbID),
		ServiceId:  ingressSvcRegisterServiceID,
	}

	resp, err := client.CreateIngressServiceWithResponse(context.Background(), id, body)
	if err != nil {
		return fmt.Errorf("register ingress service: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	svc := resp.JSON200
	if svc == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "Service registered (no details returned).")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Service %q registered on ingress %s (url: %s).\n",
		svc.ServiceId, args[0], svc.Url)
	return nil
}

// --- deregister ---

var ingressServicesDeregisterCmd = &cobra.Command{
	Use:   "deregister <ingress-id> <service-id>",
	Short: "Deregister a service from an ingress",
	Args:  cobra.ExactArgs(2),
	RunE:  runIngressServicesDeregister,
}

func runIngressServicesDeregister(cmd *cobra.Command, args []string) error {
	ok, err := confirmDestructive(cmd, ingressSvcDeregisterYes, fmt.Sprintf(
		"Deregister service %s from ingress %s?", args[1], args[0]))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	ingressID, err := uuid.Parse(args[0])
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

	resp, err := client.DeleteIngressServiceWithResponse(
		context.Background(), ingressID, args[1],
	)
	if err != nil {
		return fmt.Errorf("deregister ingress service: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Service %s deregistered from ingress %s.\n", args[1], args[0])
	return nil
}

// --- row adapter ---

type ingressServiceRowData struct {
	serviceID, databaseID, url string
}

func (r ingressServiceRowData) Columns() []string {
	return []string{r.serviceID, r.databaseID, r.url}
}

func ingressServiceRow(svc api.ServiceRegistration) ingressServiceRowData {
	return ingressServiceRowData{
		serviceID:  svc.ServiceId,
		databaseID: svc.DatabaseId.String(),
		url:        svc.Url,
	}
}
