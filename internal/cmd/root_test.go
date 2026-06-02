package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
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

func TestRootHelp(t *testing.T) {
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(out, "pgecloudctl") {
		t.Errorf("help missing 'pgecloudctl': %q", out)
	}
}
