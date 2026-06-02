package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/auth"
	"github.com/AntTheLimey/pgecloudctl/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with pgEdge Cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Fprint(cmd.OutOrStdout(), "Client ID: ")
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)

		fmt.Fprint(cmd.OutOrStdout(), "Client Secret: ")
		secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(cmd.OutOrStdout())
		if err != nil {
			return fmt.Errorf("read secret: %w", err)
		}
		clientSecret := strings.TrimSpace(string(secretBytes))

		if clientID == "" || clientSecret == "" {
			return fmt.Errorf("client ID and secret are required")
		}

		store := config.DefaultStore()
		a := &auth.Auth{Store: store, APIURL: flagAPIURL}

		tok, err := a.FetchToken(clientID, clientSecret)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		cfg := &config.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			APIURL:       flagAPIURL,
		}
		if err := store.Save(cfg); err != nil {
			return fmt.Errorf("save credentials: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(),
			"Authenticated. Token expires %s.\nCredentials saved to %s/config.json\n",
			tok.ExpiresAt.Format(time.RFC3339), store.Dir)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication state",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := config.DefaultStore()
		a := &auth.Auth{Store: store, APIURL: flagAPIURL}

		creds, source, err := a.ResolveCredentials(flagClientID, flagSecret)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Not authenticated.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Client ID:    %s\n", creds.ClientID)
		fmt.Fprintf(cmd.OutOrStdout(), "Auth source:  %s\n", source)
		fmt.Fprintf(cmd.OutOrStdout(), "API URL:      %s\n", flagAPIURL)

		tok, err := store.LoadToken()
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Token:        not cached")
		} else if time.Until(tok.ExpiresAt) <= 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Token:        expired")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Token:        valid (expires %s)\n",
				tok.ExpiresAt.Format(time.RFC3339))
		}

		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := config.DefaultStore()
		if err := store.Clear(); err != nil {
			return fmt.Errorf("clear credentials: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Logged out. Credentials removed.")
		return nil
	},
}
