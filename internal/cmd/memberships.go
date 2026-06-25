package cmd

import (
	"context"
	"fmt"

	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Membership delete flags.
var membershipDeleteYes bool

func init() {
	rootCmd.AddCommand(membershipsCmd)
	membershipsCmd.AddCommand(membershipsListCmd)
	membershipsCmd.AddCommand(membershipsDeleteCmd)

	// delete flags
	membershipsDeleteCmd.Flags().BoolVarP(&membershipDeleteYes, "yes", "y",
		false, "Skip confirmation prompt")
}

var membershipsCmd = &cobra.Command{
	Use:   "memberships",
	Short: "Manage pgEdge Cloud team memberships",
}

// --- list ---

var membershipsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List team memberships",
	RunE:  runMembershipsList,
}

func runMembershipsList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	resp, err := client.ListMembershipsWithResponse(context.Background())
	if err != nil {
		return fmt.Errorf("list memberships: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	members := resp.JSON200
	if members == nil || len(*members) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No memberships found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*members))
	for _, m := range *members {
		rows = append(rows, membershipRow{
			id:        m.Id,
			userName:  m.UserName,
			userEmail: m.UserEmail,
			createdAt: formatTime(m.CreatedAt),
		})
	}

	headers := []string{"ID", "USER NAME", "USER EMAIL", "CREATED AT"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- delete ---

var membershipsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove a team member",
	Args:  cobra.ExactArgs(1),
	RunE:  runMembershipsDelete,
}

func runMembershipsDelete(cmd *cobra.Command, args []string) error {
	id, err := uuid.Parse(args[0])
	if err != nil {
		return &ExitError{
			msg:  fmt.Sprintf("invalid membership ID %q: %v", args[0], err),
			code: ExitGeneral,
		}
	}

	ok, err := confirmDestructive(cmd, membershipDeleteYes,
		fmt.Sprintf("Remove team member %s?", args[0]))
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

	resp, err := client.DeleteMembershipWithResponse(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete membership: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Membership %s deleted.\n", args[0])
	return nil
}

// --- row adapter ---

type membershipRow struct {
	id, userName, userEmail, createdAt string
}

func (r membershipRow) Columns() []string {
	return []string{r.id, r.userName, r.userEmail, r.createdAt}
}
