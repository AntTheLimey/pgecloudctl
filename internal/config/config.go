package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config holds the OAuth client credentials and optional API URL.
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	APIURL       string `json:"api_url,omitempty"`
}

// Token holds a cached JWT access token and its expiry time.
type Token struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Store manages the ~/.pgecloudctl directory and its JSON files.
type Store struct {
	Dir string
}

// DefaultStore returns a Store rooted at ~/.pgecloudctl.
func DefaultStore() *Store {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fall back to current directory if home is unavailable.
		home = "."
	}
	return &Store{Dir: filepath.Join(home, ".pgecloudctl")}
}

// configPath returns the full path to config.json.
func (s *Store) configPath() string {
	return filepath.Join(s.Dir, "config.json")
}

// tokenPath returns the full path to token.json.
func (s *Store) tokenPath() string {
	return filepath.Join(s.Dir, "token.json")
}

// ensureDir creates the store directory with 0700 permissions if it
// does not already exist.
func (s *Store) ensureDir() error {
	return os.MkdirAll(s.Dir, 0o700)
}

// Save marshals cfg to JSON and writes it to config.json with 0600
// permissions.
func (s *Store) Save(cfg *Config) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("config: create dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	if err := os.WriteFile(s.configPath(), data, 0o600); err != nil {
		return fmt.Errorf("config: write config.json: %w", err)
	}
	return nil
}

// Load reads config.json and returns the parsed Config.
func (s *Store) Load() (*Config, error) {
	data, err := os.ReadFile(s.configPath())
	if err != nil {
		return nil, fmt.Errorf("config: read config.json: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse config.json: %w", err)
	}
	return &cfg, nil
}

// SaveToken marshals tok to JSON and writes it to token.json with
// 0600 permissions.
func (s *Store) SaveToken(tok *Token) error {
	if err := s.ensureDir(); err != nil {
		return fmt.Errorf("config: create dir: %w", err)
	}
	data, err := json.MarshalIndent(tok, "", "    ")
	if err != nil {
		return fmt.Errorf("config: marshal token: %w", err)
	}
	if err := os.WriteFile(s.tokenPath(), data, 0o600); err != nil {
		return fmt.Errorf("config: write token.json: %w", err)
	}
	return nil
}

// LoadToken reads token.json and returns the parsed Token.
func (s *Store) LoadToken() (*Token, error) {
	data, err := os.ReadFile(s.tokenPath())
	if err != nil {
		return nil, fmt.Errorf("config: read token.json: %w", err)
	}
	var tok Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, fmt.Errorf("config: parse token.json: %w", err)
	}
	return &tok, nil
}

// Clear removes config.json and token.json from the store directory.
// Missing files are not treated as errors.
func (s *Store) Clear() error {
	for _, path := range []string{s.configPath(), s.tokenPath()} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("config: remove %s: %w", path, err)
		}
	}
	return nil
}
