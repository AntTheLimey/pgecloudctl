package cmd

import (
	"fmt"

	pgecloudctl "github.com/AntTheLimey/pgecloudctl"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(llmsCmd)
}

var llmsCmd = &cobra.Command{
	Use:   "llms",
	Short: "Print the complete AI-agent reference (llms-full.txt)",
	Long: `llms prints the complete machine-readable command and workflow
reference (llms-full.txt) that ships inside the binary. AI agents should
read this before composing commands — it covers every command, flag,
exit code, async/--wait semantics, and end-to-end workflows.`,
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Fprint(cmd.OutOrStdout(), string(pgecloudctl.LLMSFull))
	},
}
