package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Cloud-account create flags.
var (
	caCreateType        string
	caCreateName        string
	caCreateDescription string

	// AWS
	caCreateRoleARN string

	// Azure
	caCreateTenantID       string
	caCreateSubscriptionID string
	caCreateClientID       string
	caCreateClientSecret   string
	caCreateResourceGroup  string

	// GCP
	caCreateProjectID      string
	caCreateServiceAccount string
)

// Cloud-account delete flags.
var caDeleteYes bool

func init() {
	rootCmd.AddCommand(cloudAccountsCmd)
	cloudAccountsCmd.AddCommand(cloudAccountsListCmd)
	cloudAccountsCmd.AddCommand(cloudAccountsGetCmd)
	cloudAccountsCmd.AddCommand(cloudAccountsCreateCmd)
	cloudAccountsCmd.AddCommand(cloudAccountsDeleteCmd)

	// create flags — provider type
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateType, "type", "",
		"Cloud provider type: aws, azure, or gcp")
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateName, "name", "",
		"Display name for the cloud account")
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateDescription, "description", "",
		"Optional description")
	_ = cloudAccountsCreateCmd.MarkFlagRequired("type")

	// AWS flags
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateRoleARN, "role-arn", "",
		"AWS IAM Role ARN (required for --type aws)")

	// Azure flags
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateTenantID, "tenant-id", "",
		"Azure tenant ID (required for --type azure)")
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateSubscriptionID,
		"subscription-id", "",
		"Azure subscription ID (required for --type azure)")
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateClientID,
		"azure-client-id", "",
		"Azure client/application ID (required for --type azure)")
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateClientSecret,
		"azure-client-secret", "",
		"Azure client secret (required for --type azure)")
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateResourceGroup,
		"resource-group", "",
		"Azure resource group (optional for --type azure)")

	// GCP flags
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateProjectID, "project-id", "",
		"GCP project ID (required for --type gcp)")
	cloudAccountsCreateCmd.Flags().StringVar(&caCreateServiceAccount,
		"service-account", "",
		"GCP service account email (required for --type gcp)")

	// delete flags
	cloudAccountsDeleteCmd.Flags().BoolVarP(&caDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")

	cloudAccountsCmd.AddCommand(cloudAccountsCFTemplateCmd)
}

var cloudAccountsCmd = &cobra.Command{
	Use:     "cloud-accounts",
	Aliases: []string{"ca"},
	Short:   "Manage pgEdge Cloud accounts",
}

// --- list ---

var cloudAccountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloud accounts",
	RunE:  runCloudAccountsList,
}

func runCloudAccountsList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.ListCloudAccountsWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("list cloud accounts: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	accounts := resp.JSON200
	if accounts == nil || len(*accounts) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No cloud accounts found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*accounts))
	for _, a := range *accounts {
		rows = append(rows, cloudAccountRow{
			id:      a.Id,
			name:    a.Name,
			typ:     a.Type,
			created: formatTime(a.CreatedAt),
		})
	}

	headers := []string{"ID", "NAME", "TYPE", "CREATED"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var cloudAccountsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get cloud account details",
	Args:  cobra.ExactArgs(1),
	RunE:  runCloudAccountsGet,
}

func runCloudAccountsGet(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid cloud account ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetCloudAccountWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get cloud account: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	a := resp.JSON200
	if a == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No cloud account data returned.")
		return nil
	}

	rows := []output.Row{
		cloudAccountRow{
			id:      a.Id,
			name:    a.Name,
			typ:     a.Type,
			created: formatTime(a.CreatedAt),
		},
	}

	headers := []string{"ID", "NAME", "TYPE", "CREATED"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var cloudAccountsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cloud account",
	RunE:  runCloudAccountsCreate,
}

func runCloudAccountsCreate(cmd *cobra.Command, _ []string) error {
	providerType := strings.ToLower(caCreateType)

	var creds api.CreateCloudAccountInput_Credentials

	switch providerType {
	case "aws":
		if caCreateRoleARN == "" {
			return &ExitError{
				msg:  "--role-arn is required for --type aws",
				code: ExitGeneral,
			}
		}
		if err := creds.FromAwsCredentials(api.AwsCredentials{
			RoleArn: caCreateRoleARN,
		}); err != nil {
			return fmt.Errorf("build AWS credentials: %w", err)
		}

	case "azure":
		missing := []string{}
		if caCreateTenantID == "" {
			missing = append(missing, "--tenant-id")
		}
		if caCreateSubscriptionID == "" {
			missing = append(missing, "--subscription-id")
		}
		if caCreateClientID == "" {
			missing = append(missing, "--azure-client-id")
		}
		if caCreateClientSecret == "" {
			missing = append(missing, "--azure-client-secret")
		}
		if len(missing) > 0 {
			return &ExitError{
				msg: fmt.Sprintf(
					"%s required for --type azure",
					strings.Join(missing, ", "),
				),
				code: ExitGeneral,
			}
		}
		azCreds := api.AzureCredentials{
			TenantId:       caCreateTenantID,
			SubscriptionId: caCreateSubscriptionID,
			ClientId:       caCreateClientID,
			ClientSecret:   caCreateClientSecret,
		}
		if caCreateResourceGroup != "" {
			azCreds.ResourceGroup = &caCreateResourceGroup
		}
		if err := creds.FromAzureCredentials(azCreds); err != nil {
			return fmt.Errorf("build Azure credentials: %w", err)
		}

	case "gcp":
		missing := []string{}
		if caCreateProjectID == "" {
			missing = append(missing, "--project-id")
		}
		if caCreateServiceAccount == "" {
			missing = append(missing, "--service-account")
		}
		if len(missing) > 0 {
			return &ExitError{
				msg: fmt.Sprintf(
					"%s required for --type gcp",
					strings.Join(missing, ", "),
				),
				code: ExitGeneral,
			}
		}
		if err := creds.FromGoogleCredentials(api.GoogleCredentials{
			ProjectId:      caCreateProjectID,
			Provider:       "gcp",
			ServiceAccount: caCreateServiceAccount,
		}); err != nil {
			return fmt.Errorf("build GCP credentials: %w", err)
		}

	default:
		return &ExitError{
			msg:  fmt.Sprintf("unknown provider type %q: must be aws, azure, or gcp", caCreateType),
			code: ExitGeneral,
		}
	}

	body := api.CreateCloudAccountJSONRequestBody{
		Type:        providerType,
		Credentials: creds,
	}
	if caCreateName != "" {
		body.Name = &caCreateName
	}
	if caCreateDescription != "" {
		body.Description = &caCreateDescription
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.CreateCloudAccountWithResponse(context.Background(), body)
	if err != nil {
		return fmt.Errorf("create cloud account: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	a := resp.JSON200
	if a == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "Cloud account created (no details returned).")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(),
		"Cloud account %q created (id: %s, type: %s).\n",
		a.Name, a.Id, a.Type)
	return nil
}

// --- delete ---

var cloudAccountsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a cloud account",
	Args:  cobra.ExactArgs(1),
	RunE:  runCloudAccountsDelete,
}

func runCloudAccountsDelete(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid cloud account ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	ok, err := confirmDestructive(cmd, caDeleteYes,
		fmt.Sprintf("Delete cloud account %s? This cannot be undone.",
			args[0]))
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

	resp, err := client.DeleteCloudAccountWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete cloud account: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Cloud account %s deleted.\n", args[0])
	return nil
}

// --- cloudformation-template ---

var cloudAccountsCFTemplateCmd = &cobra.Command{
	Use:   "cloudformation-template",
	Short: "Get the CloudFormation template for AWS IAM setup",
	RunE:  runCloudAccountsCFTemplate,
}

func runCloudAccountsCFTemplate(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetCloudFormationTemplateWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("get cloudformation template: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	templates := resp.JSON200
	if templates == nil || len(*templates) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No template returned.")
		return nil
	}

	for _, tmpl := range *templates {
		fmt.Fprintln(cmd.OutOrStdout(), tmpl.Url)
	}
	return nil
}

// --- row adapter ---

type cloudAccountRow struct {
	id, name, typ, created string
}

func (r cloudAccountRow) Columns() []string {
	return []string{r.id, r.name, r.typ, r.created}
}
