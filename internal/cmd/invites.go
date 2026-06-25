package cmd

import (
	"context"
	"fmt"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Invite create flags.
var (
	inviteCreateEmail      string
	inviteCreateExpiration int
)

// Invite accept flags.
var inviteAcceptToken string

// Invite delete flags.
var inviteDeleteYes bool

func init() {
	rootCmd.AddCommand(invitesCmd)
	invitesCmd.AddCommand(invitesListCmd)
	invitesCmd.AddCommand(invitesGetCmd)
	invitesCmd.AddCommand(invitesCreateCmd)
	invitesCmd.AddCommand(invitesDeleteCmd)
	invitesCmd.AddCommand(invitesAcceptCmd)

	// create flags
	invitesCreateCmd.Flags().StringVar(&inviteCreateEmail, "email", "",
		"Email address to invite")
	invitesCreateCmd.Flags().IntVar(&inviteCreateExpiration, "expiration", 0,
		"Invite expiration in hours (optional)")
	_ = invitesCreateCmd.MarkFlagRequired("email")

	// accept flags
	invitesAcceptCmd.Flags().StringVar(&inviteAcceptToken, "token", "",
		"Invite token")
	_ = invitesAcceptCmd.MarkFlagRequired("token")

	// delete flags
	invitesDeleteCmd.Flags().BoolVarP(&inviteDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
}

var invitesCmd = &cobra.Command{
	Use:   "invites",
	Short: "Manage pgEdge Cloud invites",
}

// --- list ---

var invitesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List invites",
	RunE:  runInvitesList,
}

func runInvitesList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.ListInvitesWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("list invites: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	invites := resp.JSON200
	if invites == nil || len(*invites) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No invites found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*invites))
	for _, inv := range *invites {
		rows = append(rows, inviteRow{
			id:        inv.Id,
			email:     inv.Email,
			invitedBy: derefString(inv.InvitedBy),
			team:      inv.TeamName,
			expiresAt: formatTime(inv.ExpiresAt),
			createdAt: formatTime(inv.CreatedAt),
		})
	}

	headers := []string{"ID", "EMAIL", "INVITED BY", "TEAM", "EXPIRES AT", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var invitesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get invite details",
	Args:  cobra.ExactArgs(1),
	RunE:  runInvitesGet,
}

func runInvitesGet(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid invite ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.GetInviteWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get invite: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	inv := resp.JSON200
	if inv == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No invite data returned.")
		return nil
	}

	rows := []output.Row{
		inviteRow{
			id:        inv.Id,
			email:     inv.Email,
			invitedBy: derefString(inv.InvitedBy),
			team:      inv.TeamName,
			expiresAt: formatTime(inv.ExpiresAt),
			createdAt: formatTime(inv.CreatedAt),
		},
	}

	headers := []string{"ID", "EMAIL", "INVITED BY", "TEAM", "EXPIRES AT", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- create ---

var invitesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new invite",
	RunE:  runInvitesCreate,
}

func runInvitesCreate(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	if inviteCreateExpiration < 0 {
		return &ExitError{
			msg:  "--expiration must be a positive number of hours",
			code: ExitGeneral,
		}
	}

	body := api.CreateInviteJSONRequestBody{
		Email: inviteCreateEmail,
	}
	if inviteCreateExpiration > 0 {
		body.Expiration = &inviteCreateExpiration
	}

	resp, err := client.CreateInviteWithResponse(context.Background(), body)
	if err != nil {
		return fmt.Errorf("create invite: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	inv := resp.JSON200
	if inv == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "Invite created (no details returned).")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Invite for %q created (id: %s).\n",
		inv.Email, inv.Id)
	return nil
}

// --- delete ---

var invitesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an invite",
	Args:  cobra.ExactArgs(1),
	RunE:  runInvitesDelete,
}

func runInvitesDelete(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid invite ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	ok, err := confirmDestructive(cmd, inviteDeleteYes,
		fmt.Sprintf("Delete invite %s? This cannot be undone.", args[0]))
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

	resp, err := client.DeleteInviteWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete invite: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Invite %s deleted.\n", args[0])
	return nil
}

// --- accept ---

var invitesAcceptCmd = &cobra.Command{
	Use:   "accept <id>",
	Short: "Accept an invite",
	Args:  cobra.ExactArgs(1),
	RunE:  runInvitesAccept,
}

func runInvitesAccept(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid invite ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.AcceptInviteWithResponse(
		context.Background(), id,
		&api.AcceptInviteParams{InviteToken: inviteAcceptToken},
	)
	if err != nil {
		return fmt.Errorf("accept invite: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Invite accepted.")
	return nil
}

// --- row adapter ---

type inviteRow struct {
	id, email, invitedBy, team, expiresAt, createdAt string
}

func (r inviteRow) Columns() []string {
	return []string{r.id, r.email, r.invitedBy, r.team, r.expiresAt, r.createdAt}
}
