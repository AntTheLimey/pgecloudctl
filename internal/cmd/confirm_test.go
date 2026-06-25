package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func confirmTestCmd(stdin string) (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	c := &cobra.Command{}
	c.SetOut(buf)
	c.SetIn(strings.NewReader(stdin))
	return c, buf
}

// withTerminal temporarily overrides stdinIsTerminal for a test.
func withTerminal(t *testing.T, isTTY bool) {
	t.Helper()
	orig := stdinIsTerminal
	stdinIsTerminal = func() bool { return isTTY }
	t.Cleanup(func() { stdinIsTerminal = orig })
}

func TestConfirmDestructive_SkipProceeds(t *testing.T) {
	// --yes set: proceed without consulting the terminal at all.
	withTerminal(t, false)
	cmd, _ := confirmTestCmd("")

	ok, err := confirmDestructive(cmd, true, "Delete it?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected proceed=true when skip is set")
	}
}

func TestConfirmDestructive_NonInteractiveFailsLoud(t *testing.T) {
	withTerminal(t, false)
	cmd, _ := confirmTestCmd("")

	ok, err := confirmDestructive(cmd, false, "Delete it?")
	if ok {
		t.Error("expected proceed=false in non-interactive context")
	}
	if err == nil {
		t.Fatal("expected error in non-interactive context without --yes")
	}
	ee, isExit := err.(*ExitError)
	if !isExit {
		t.Fatalf("expected *ExitError, got %T", err)
	}
	if ee.Code() != ExitGeneral {
		t.Errorf("exit code = %d, want %d", ee.Code(), ExitGeneral)
	}
}

func TestConfirmDestructive_TerminalYesProceeds(t *testing.T) {
	withTerminal(t, true)
	cmd, buf := confirmTestCmd("y\n")

	ok, err := confirmDestructive(cmd, false, "Delete it?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected proceed=true after 'y'")
	}
	if !strings.Contains(buf.String(), "Delete it? [y/N]:") {
		t.Errorf("expected prompt in output, got: %q", buf.String())
	}
}

func TestConfirmDestructive_TerminalNoAborts(t *testing.T) {
	withTerminal(t, true)
	cmd, buf := confirmTestCmd("n\n")

	ok, err := confirmDestructive(cmd, false, "Delete it?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected proceed=false after 'n'")
	}
	if !strings.Contains(buf.String(), "Aborted.") {
		t.Errorf("expected 'Aborted.' in output, got: %q", buf.String())
	}
}
