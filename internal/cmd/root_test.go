package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/spf13/cobra"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	// Reset the help flag on every command so that a previous --help call
	// does not cause cobra to print help instead of running the command.
	resetHelpFlags(rootCmd)
	err := rootCmd.Execute()
	return buf.String(), err
}

// resetHelpFlags sets the "help" flag back to false on cmd and all its
// subcommands, preventing flag state from leaking between test calls.
func resetHelpFlags(cmd *cobra.Command) {
	if f := cmd.Flags().Lookup("help"); f != nil {
		_ = f.Value.Set("false")
		f.Changed = false
	}
	for _, sub := range cmd.Commands() {
		resetHelpFlags(sub)
	}
}

func TestVersionCommand(t *testing.T) {
	Version = "v0.1.0-test"
	out, err := executeCommand("version")
	if err != nil {
		t.Fatalf("version command: %v", err)
	}
	if !strings.Contains(out, "v0.1.0-test") {
		t.Errorf("output = %q, want version string", out)
	}
}

func TestNoColorFlag(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	_, err := executeCommand("version")
	if err != nil {
		t.Fatalf("version with --no-color: %v", err)
	}

	if output.ColorEnabled {
		t.Error("ColorEnabled should be false when --no-color is set")
	}
}

func TestNoColorEnvVar(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	flagNoColor = false

	_, err := executeCommand("version")
	if err != nil {
		t.Fatalf("version with NO_COLOR: %v", err)
	}

	if output.ColorEnabled {
		t.Error("ColorEnabled should be false when NO_COLOR env is set")
	}
}

func TestRootHelp(t *testing.T) {
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(out, "pgecloudctl") {
		t.Errorf("help missing 'pgecloudctl': %q", out)
	}
}
