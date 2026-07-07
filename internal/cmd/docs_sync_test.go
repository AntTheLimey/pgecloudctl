package cmd

import (
	"strings"
	"testing"

	pgecloudctl "github.com/AntTheLimey/pgecloudctl"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TestLLMSFullCoversCommandTree walks the cobra command tree and fails
// when a command or one of its local flags is missing from the embedded
// llms-full.txt. This keeps the AI-facing reference authoritative:
// adding a command or flag without documenting it breaks the build.
func TestLLMSFullCoversCommandTree(t *testing.T) {
	doc := string(pgecloudctl.LLMSFull)

	var walk func(cmd *cobra.Command, path string)
	walk = func(cmd *cobra.Command, path string) {
		if cmd.Hidden || cmd.Name() == "help" ||
			cmd.Name() == "completion" {
			return
		}

		if path != "" && !strings.Contains(doc, path) {
			t.Errorf("llms-full.txt does not mention command %q — "+
				"document it (and its flags) before shipping", path)
		}

		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return
			}
			if !strings.Contains(doc, "--"+f.Name) {
				t.Errorf("llms-full.txt does not mention flag --%s "+
					"of %q — document it before shipping", f.Name, path)
			}
		})

		for _, sub := range cmd.Commands() {
			subPath := strings.TrimSpace(path + " " + sub.Name())
			walk(sub, subPath)
		}
	}

	walk(rootCmd, "")
}
