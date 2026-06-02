package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/config"
)

const refreshWindow = 5 * time.Minute

// Credentials holds a resolved client ID and secret.
type Credentials struct {
	ClientID     string
	ClientSecret string
}

// Auth performs credential resolution and token management.
type Auth struct {
	Store  *config.Store
	APIURL string
}

// ResolveCredentials returns credentials from the highest-priority source:
//
//  1. Environment variables PGEDGE_CLIENT_ID / PGEDGE_CLIENT_SECRET
//  2. flagID / flagSecret arguments
//  3. Saved config file in the Store
//
// The second return value is one of "environment", "flags", or "config".
// An error is returned when no credentials are available from any source.
func (a *Auth) ResolveCredentials(flagID, flagSecret string) (*Credentials, string, error) {
	// 1. Environment variables.
	envID := os.Getenv("PGEDGE_CLIENT_ID")
	envSecret := os.Getenv("PGEDGE_CLIENT_SECRET")
	if envID != "" && envSecret != "" {
		return &Credentials{ClientID: envID, ClientSecret: envSecret}, "environment", nil
	}

	// 2. CLI flags.
	if flagID != "" && flagSecret != "" {
		return &Credentials{ClientID: flagID, ClientSecret: flagSecret}, "flags", nil
	}

	// 3. Config file.
	cfg, err := a.Store.Load()
	if err == nil && cfg.ClientID != "" && cfg.ClientSecret != "" {
		return &Credentials{ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret}, "config", nil
	}

	return nil, "", fmt.Errorf("auth: no credentials found — set PGEDGE_CLIENT_ID/PGEDGE_CLIENT_SECRET, use --client-id/--client-secret flags, or run 'pgecloudctl auth login'")
}

// tokenRequest is the JSON body sent to /oauth/token.
type tokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// tokenResponse is the JSON body returned by /oauth/token.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// FetchToken posts to APIURL/oauth/token, caches the resulting token via
// the Store, and returns it.
func (a *Auth) FetchToken(clientID, clientSecret string) (*config.Token, error) {
	body, err := json.Marshal(tokenRequest{ClientID: clientID, ClientSecret: clientSecret})
	if err != nil {
		return nil, fmt.Errorf("auth: marshal request: %w", err)
	}

	url := a.APIURL + "/oauth/token"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:noctx // token endpoint is a simple POST without request-scoped context
	if err != nil {
		return nil, fmt.Errorf("auth: POST %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth: token endpoint returned %s: %s", resp.Status, string(raw))
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("auth: decode token response: %w", err)
	}

	tok := &config.Token{
		AccessToken: tr.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second),
	}
	if err := a.Store.SaveToken(tok); err != nil {
		return nil, fmt.Errorf("auth: cache token: %w", err)
	}
	return tok, nil
}

// GetToken returns a valid access token string. It loads the cached token
// and reuses it if it is more than refreshWindow away from expiry. Otherwise
// it fetches a fresh token.
func (a *Auth) GetToken(clientID, clientSecret string) (string, error) {
	tok, err := a.Store.LoadToken()
	if err == nil && time.Until(tok.ExpiresAt) > refreshWindow {
		return tok.AccessToken, nil
	}

	fresh, err := a.FetchToken(clientID, clientSecret)
	if err != nil {
		return "", err
	}
	return fresh.AccessToken, nil
}
