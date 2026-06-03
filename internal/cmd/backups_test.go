package cmd

import (
	"strings"
	"testing"
)

func TestBackupsCreateRequiresFlags(t *testing.T) {
	_, err := executeCommand("backups", "create")
	if err == nil {
		t.Fatal("expected error when required flags missing")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error = %q, want 'required' flag error", err.Error())
	}
}

func TestBackupsGetRequiresArg(t *testing.T) {
	_, err := executeCommand("backups", "get")
	if err == nil {
		t.Fatal("expected error when no arg provided")
	}
}

func TestBackupStoresCreateRequiresFlags(t *testing.T) {
	_, err := executeCommand("backup-stores", "create")
	if err == nil {
		t.Fatal("expected error when required flags missing")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error = %q, want 'required' flag error", err.Error())
	}
}
