package cmd

import (
	"testing"
)

func TestBuildDatabaseBackups(t *testing.T) {
	b := buildDatabaseBackups("bs-abc123")
	if b == nil {
		t.Fatal("buildDatabaseBackups returned nil")
	}
	if b.Config == nil || len(*b.Config) != 1 {
		t.Fatalf("config = %v, want 1 entry", b.Config)
	}
	repos := (*b.Config)[0].Repositories
	if repos == nil || len(*repos) != 1 {
		t.Fatalf("repositories = %v, want 1 entry", repos)
	}
	got := (*repos)[0].BackupStoreId
	if got == nil || *got != "bs-abc123" {
		t.Fatalf("backup_store_id = %v, want bs-abc123", got)
	}
}
