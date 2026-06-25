package cmd

import (
	"context"
	"fmt"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// SSH key create flags.
var (
	sshKeyCreateName      string
	sshKeyCreatePublicKey string
)

// SSH key delete flags.
var sshKeyDeleteYes bool

func init() {
	rootCmd.AddCommand(sshKeysCmd)
	sshKeysCmd.AddCommand(sshKeysListCmd)
	sshKeysCmd.AddCommand(sshKeysGetCmd)
	sshKeysCmd.AddCommand(sshKeysCreateCmd)
	sshKeysCmd.AddCommand(sshKeysDeleteCmd)

	// create flags
	sshKeysCreateCmd.Flags().StringVar(&sshKeyCreateName, "name", "",
		"SSH key name")
	sshKeysCreateCmd.Flags().StringVar(&sshKeyCreatePublicKey, "public-key", "",
		"SSH public key value")
	_ = sshKeysCreateCmd.MarkFlagRequired("name")
	_ = sshKeysCreateCmd.MarkFlagRequired("public-key")

	// delete flags
	sshKeysDeleteCmd.Flags().BoolVarP(&sshKeyDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
}

var sshKeysCmd = &cobra.Command{
	Use:   "ssh-keys",
	Short: "Manage pgEdge Cloud SSH keys",
}

// --- list ---

var sshKeysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SSH keys",
	RunE:  runSSHKeysList,
}

func runSSHKeysList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.ListSshKeysWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("list ssh keys: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	keys := resp.JSON200
	if keys == nil || len(*keys) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No SSH keys found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*keys))
	for _, k := range *keys {
		rows = append(rows, sshKeyRow{
			id:      k.Id,
			name:    k.Name,
			created: formatTime(k.CreatedAt),
		})
	}

	headers := []string{"ID", "NAME", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var sshKeysGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get SSH key details",
	Args:  cobra.ExactArgs(1),
	RunE:  runSSHKeysGet,
}

func runSSHKeysGet(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid SSH key ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetSshKeyWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get ssh key: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	k := resp.JSON200
	if k == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No SSH key data returned.")
		return nil
	}

	rows := []output.Row{
		sshKeyRow{
			id:      k.Id,
			name:    k.Name,
			created: formatTime(k.CreatedAt),
		},
	}

	headers := []string{"ID", "NAME", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var sshKeysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new SSH key",
	RunE:  runSSHKeysCreate,
}

func runSSHKeysCreate(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	body := api.CreateSshKeyJSONRequestBody{
		Name:      sshKeyCreateName,
		PublicKey: sshKeyCreatePublicKey,
	}

	resp, err := client.CreateSshKeyWithResponse(context.Background(), body)
	if err != nil {
		return fmt.Errorf("create ssh key: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	k := resp.JSON200
	if k == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "SSH key created (no details returned).")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "SSH key %q created (id: %s).\n",
		k.Name, k.Id)
	return nil
}

// --- delete ---

var sshKeysDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an SSH key",
	Args:  cobra.ExactArgs(1),
	RunE:  runSSHKeysDelete,
}

func runSSHKeysDelete(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid SSH key ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	ok, err := confirmDestructive(cmd, sshKeyDeleteYes,
		fmt.Sprintf("Delete SSH key %s? This cannot be undone.", args[0]))
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

	resp, err := client.DeleteSshKeyWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete ssh key: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "SSH key %s deleted.\n", args[0])
	return nil
}

// --- row adapter ---

type sshKeyRow struct {
	id, name, created string
}

func (r sshKeyRow) Columns() []string {
	return []string{r.id, r.name, r.created}
}
