package config

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func TestSaveAndLoad(t *testing.T) {
    dir := t.TempDir()
    s := &Store{Dir: dir}

    cfg := &Config{
        ClientID:     "test-client-id",
        ClientSecret: "test-client-secret",
        APIURL:       "https://api.example.com",
    }

    if err := s.Save(cfg); err != nil {
        t.Fatalf("Save() error = %v", err)
    }

    got, err := s.Load()
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }

    if got.ClientID != cfg.ClientID {
        t.Errorf("ClientID = %q, want %q", got.ClientID, cfg.ClientID)
    }
    if got.ClientSecret != cfg.ClientSecret {
        t.Errorf("ClientSecret = %q, want %q", got.ClientSecret, cfg.ClientSecret)
    }
    if got.APIURL != cfg.APIURL {
        t.Errorf("APIURL = %q, want %q", got.APIURL, cfg.APIURL)
    }
}

func TestFilePermissions(t *testing.T) {
    dir := t.TempDir()
    s := &Store{Dir: dir}

    cfg := &Config{
        ClientID:     "test-client-id",
        ClientSecret: "test-client-secret",
    }

    if err := s.Save(cfg); err != nil {
        t.Fatalf("Save() error = %v", err)
    }

    info, err := os.Stat(filepath.Join(dir, "config.json"))
    if err != nil {
        t.Fatalf("Stat() error = %v", err)
    }

    got := info.Mode().Perm()
    if got != 0600 {
        t.Errorf("file permissions = %04o, want 0600", got)
    }
}

func TestLoadMissing(t *testing.T) {
    dir := t.TempDir()
    s := &Store{Dir: dir}

    _, err := s.Load()
    if err == nil {
        t.Error("Load() on missing file should return an error, got nil")
    }
}

func TestSaveAndLoadToken(t *testing.T) {
    dir := t.TempDir()
    s := &Store{Dir: dir}

    expires := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
    tok := &Token{
        AccessToken: "my-access-token",
        ExpiresAt:   expires,
    }

    if err := s.SaveToken(tok); err != nil {
        t.Fatalf("SaveToken() error = %v", err)
    }

    got, err := s.LoadToken()
    if err != nil {
        t.Fatalf("LoadToken() error = %v", err)
    }

    if got.AccessToken != tok.AccessToken {
        t.Errorf("AccessToken = %q, want %q", got.AccessToken, tok.AccessToken)
    }
    if !got.ExpiresAt.Equal(tok.ExpiresAt) {
        t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, tok.ExpiresAt)
    }
}

func TestClear(t *testing.T) {
    dir := t.TempDir()
    s := &Store{Dir: dir}

    cfg := &Config{ClientID: "id", ClientSecret: "secret"}
    if err := s.Save(cfg); err != nil {
        t.Fatalf("Save() error = %v", err)
    }

    tok := &Token{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)}
    if err := s.SaveToken(tok); err != nil {
        t.Fatalf("SaveToken() error = %v", err)
    }

    if err := s.Clear(); err != nil {
        t.Fatalf("Clear() error = %v", err)
    }

    if _, err := os.Stat(filepath.Join(dir, "config.json")); !os.IsNotExist(err) {
        t.Error("config.json should not exist after Clear()")
    }
    if _, err := os.Stat(filepath.Join(dir, "token.json")); !os.IsNotExist(err) {
        t.Error("token.json should not exist after Clear()")
    }
}

func TestDefaultDir(t *testing.T) {
    home, err := os.UserHomeDir()
    if err != nil {
        t.Fatalf("UserHomeDir() error = %v", err)
    }

    s := DefaultStore()
    want := filepath.Join(home, ".pgecloudctl")

    if s.Dir != want {
        t.Errorf("DefaultStore().Dir = %q, want %q", s.Dir, want)
    }
}
