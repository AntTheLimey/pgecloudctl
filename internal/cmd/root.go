package cmd

import (
	"os"

	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	Version = "dev"

	flagAPIURL   string
	flagOutput   string
	flagNoColor  bool
	flagVerbose  bool
	flagClientID string
	flagSecret   string
)

var rootCmd = &cobra.Command{
	Use:   "pgecloudctl",
	Short: "CLI for managing pgEdge Cloud resources",
	Long: `pgecloudctl manages pgEdge Cloud clusters, databases, and services via the REST API.

AI agents: do not improvise from --help output alone. Run ` +
		"`pgecloudctl llms`" + ` for the complete command and workflow
reference (llms-full.txt), and run ` + "`pgecloudctl skill install`" + `
to install the bundled Claude Code skill (~/.claude/skills).`,
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		output.ColorEnabled = !flagNoColor &&
			os.Getenv("NO_COLOR") == "" &&
			term.IsTerminal(int(os.Stdout.Fd()))
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagAPIURL, "api-url",
		"https://api.pgedge.com", "API base URL")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o",
		"table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color",
		false, "Disable color output")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v",
		false, "Show HTTP request/response details")
	rootCmd.PersistentFlags().StringVar(&flagClientID, "client-id",
		"", "API client ID (overrides config file)")
	rootCmd.PersistentFlags().StringVar(&flagSecret, "client-secret",
		"", "API client secret (overrides config file)")

	if v := os.Getenv("PGEDGE_API_URL"); v != "" {
		flagAPIURL = v
	}
}

func Execute() error {
	return rootCmd.Execute()
}
