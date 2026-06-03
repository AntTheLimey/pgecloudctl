package cmd

import (
	"strings"
	"testing"
)

func TestSSHKeysCreateRequiresFlags(t *testing.T) {
	_, err := executeCommand("ssh-keys", "create")
	if err == nil {
		t.Fatal("expected error when required flags missing")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error = %q, want 'required' flag error", err.Error())
	}
}

func TestSSHKeysGetRequiresArg(t *testing.T) {
	_, err := executeCommand("ssh-keys", "get")
	if err == nil {
		t.Fatal("expected error when no arg provided")
	}
}

func TestSSHKeysDeleteRequiresArg(t *testing.T) {
	_, err := executeCommand("ssh-keys", "delete")
	if err == nil {
		t.Fatal("expected error when no arg provided")
	}
}

func TestSSHKeysHelp(t *testing.T) {
	out, err := executeCommand("ssh-keys", "--help")
	if err != nil {
		t.Fatalf("ssh-keys help: %v", err)
	}
	if !strings.Contains(out, "ssh-keys") {
		t.Errorf("help missing 'ssh-keys': %q", out)
	}
}
