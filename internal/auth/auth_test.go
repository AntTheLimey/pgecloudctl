package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/config"
)

// newStore returns a config.Store backed by a temp directory.
func newStore(t *testing.T) *config.Store {
	t.Helper()
	return &config.Store{Dir: t.TempDir()}
}

// TestResolveCredentials_EnvVars verifies that PGEDGE_CLIENT_ID and
// PGEDGE_CLIENT_SECRET env vars take priority and source is "environment".
func TestResolveCredentials_EnvVars(t *testing.T) {
	t.Setenv("PGEDGE_CLIENT_ID", "env-id")
	t.Setenv("PGEDGE_CLIENT_SECRET", "env-secret")

	a := &Auth{Store: newStore(t), APIURL: "https://example.com"}
	creds, source, err := a.ResolveCredentials("", "")
	if err != nil {
		t.Fatalf("ResolveCredentials() error = %v", err)
	}
	if source != "environment" {
		t.Errorf("source = %q, want %q", source, "environment")
	}
	if creds.ClientID != "env-id" {
		t.Errorf("ClientID = %q, want %q", creds.ClientID, "env-id")
	}
	if creds.ClientSecret != "env-secret" {
		t.Errorf("ClientSecret = %q, want %q", creds.ClientSecret, "env-secret")
	}
}

// TestResolveCredentials_Flags verifies that flag values are used when no
// env vars are set, with source "flags".
func TestResolveCredentials_Flags(t *testing.T) {
	t.Setenv("PGEDGE_CLIENT_ID", "")
	t.Setenv("PGEDGE_CLIENT_SECRET", "")

	a := &Auth{Store: newStore(t), APIURL: "https://example.com"}
	creds, source, err := a.ResolveCredentials("flag-id", "flag-secret")
	if err != nil {
		t.Fatalf("ResolveCredentials() error = %v", err)
	}
	if source != "flags" {
		t.Errorf("source = %q, want %q", source, "flags")
	}
	if creds.ClientID != "flag-id" {
		t.Errorf("ClientID = %q, want %q", creds.ClientID, "flag-id")
	}
	if creds.ClientSecret != "flag-secret" {
		t.Errorf("ClientSecret = %q, want %q", creds.ClientSecret, "flag-secret")
	}
}

// TestResolveCredentials_ConfigFile verifies that stored config is used when
// no env vars or flags are present, with source "config".
func TestResolveCredentials_ConfigFile(t *testing.T) {
	t.Setenv("PGEDGE_CLIENT_ID", "")
	t.Setenv("PGEDGE_CLIENT_SECRET", "")

	store := newStore(t)
	cfg := &config.Config{ClientID: "cfg-id", ClientSecret: "cfg-secret"}
	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	a := &Auth{Store: store, APIURL: "https://example.com"}
	creds, source, err := a.ResolveCredentials("", "")
	if err != nil {
		t.Fatalf("ResolveCredentials() error = %v", err)
	}
	if source != "config" {
		t.Errorf("source = %q, want %q", source, "config")
	}
	if creds.ClientID != "cfg-id" {
		t.Errorf("ClientID = %q, want %q", creds.ClientID, "cfg-id")
	}
	if creds.ClientSecret != "cfg-secret" {
		t.Errorf("ClientSecret = %q, want %q", creds.ClientSecret, "cfg-secret")
	}
}

// TestResolveCredentials_NoneFound verifies an error is returned when no
// credentials are available from any source.
func TestResolveCredentials_NoneFound(t *testing.T) {
	t.Setenv("PGEDGE_CLIENT_ID", "")
	t.Setenv("PGEDGE_CLIENT_SECRET", "")

	a := &Auth{Store: newStore(t), APIURL: "https://example.com"}
	_, _, err := a.ResolveCredentials("", "")
	if err == nil {
		t.Error("ResolveCredentials() expected error when no credentials found, got nil")
	}
}

// TestFetchToken verifies that FetchToken POSTs to /oauth/token and returns
// the JWT with an expiry approximately 24 hours from now.
func TestFetchToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/oauth/token" {
			t.Errorf("path = %q, want /oauth/token", r.URL.Path)
		}

		var body struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if body.ClientID != "test-id" || body.ClientSecret != "test-secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-jwt-token",
			"token_type":   "Bearer",
			"expires_in":   86400,
		})
	}))
	defer srv.Close()

	store := newStore(t)
	a := &Auth{Store: store, APIURL: srv.URL}

	tok, err := a.FetchToken("test-id", "test-secret")
	if err != nil {
		t.Fatalf("FetchToken() error = %v", err)
	}
	if tok.AccessToken != "test-jwt-token" {
		t.Errorf("AccessToken = %q, want %q", tok.AccessToken, "test-jwt-token")
	}

	// Expiry should be approximately 24 hours from now.
	want := time.Now().Add(24 * time.Hour)
	diff := tok.ExpiresAt.Sub(want)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("ExpiresAt = %v, want ~%v", tok.ExpiresAt, want)
	}
}

// TestGetToken_CachesAndRefreshes verifies that:
//   - The first call fetches a fresh token and caches it.
//   - A second call uses the cached token (no extra HTTP requests).
//   - After the cached token expires, a fresh token is fetched.
func TestGetToken_CachesAndRefreshes(t *testing.T) {
	fetchCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "fresh-token",
			"token_type":   "Bearer",
			"expires_in":   86400,
		})
	}))
	defer srv.Close()

	store := newStore(t)
	a := &Auth{Store: store, APIURL: srv.URL}

	// First call — should fetch.
	tok1, err := a.GetToken("id", "secret")
	if err != nil {
		t.Fatalf("GetToken() first call error = %v", err)
	}
	if fetchCount != 1 {
		t.Errorf("fetchCount after first call = %d, want 1", fetchCount)
	}

	// Second call — should hit cache.
	tok2, err := a.GetToken("id", "secret")
	if err != nil {
		t.Fatalf("GetToken() second call error = %v", err)
	}
	if fetchCount != 1 {
		t.Errorf("fetchCount after second call = %d, want 1 (cached)", fetchCount)
	}
	if tok1 != tok2 {
		t.Errorf("tok1 = %q, tok2 = %q; expected same cached token", tok1, tok2)
	}

	// Expire the cached token by overwriting it with an expired one.
	expired := &config.Token{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-time.Hour),
	}
	if err := store.SaveToken(expired); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	// Third call — should fetch again.
	tok3, err := a.GetToken("id", "secret")
	if err != nil {
		t.Fatalf("GetToken() third call error = %v", err)
	}
	if fetchCount != 2 {
		t.Errorf("fetchCount after expired token = %d, want 2", fetchCount)
	}
	if tok3 != "fresh-token" {
		t.Errorf("tok3 = %q, want %q", tok3, "fresh-token")
	}
}

// TestGetToken_RefreshesWhenNearExpiry verifies that a token expiring within
// the 5-minute refresh window triggers a new fetch.
func TestGetToken_RefreshesWhenNearExpiry(t *testing.T) {
	fetchCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "refreshed-token",
			"token_type":   "Bearer",
			"expires_in":   86400,
		})
	}))
	defer srv.Close()

	store := newStore(t)

	// Store a token that expires in 3 minutes — inside the 5-minute window.
	nearExpiry := &config.Token{
		AccessToken: "near-expiry-token",
		ExpiresAt:   time.Now().Add(3 * time.Minute),
	}
	if err := store.SaveToken(nearExpiry); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	a := &Auth{Store: store, APIURL: srv.URL}
	tok, err := a.GetToken("id", "secret")
	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}
	if fetchCount != 1 {
		t.Errorf("fetchCount = %d, want 1 (should have refreshed)", fetchCount)
	}
	if tok != "refreshed-token" {
		t.Errorf("tok = %q, want %q", tok, "refreshed-token")
	}
}
