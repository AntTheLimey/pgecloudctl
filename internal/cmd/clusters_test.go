package cmd

import (
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
)

func strSlice(p *[]string) []string {
	if p == nil {
		return nil
	}
	return *p
}

func TestParseFirewallRule(t *testing.T) {
	name := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}

	t.Run("happy path with one source", func(t *testing.T) {
		r, err := parseFirewallRule("name=https,port=443,sources=0.0.0.0/0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name(r.Name) != "https" {
			t.Errorf("name = %q, want https", name(r.Name))
		}
		if r.Port != 443 {
			t.Errorf("port = %d, want 443", r.Port)
		}
		if got := strSlice(r.Sources); len(got) != 1 || got[0] != "0.0.0.0/0" {
			t.Errorf("sources = %v, want [0.0.0.0/0]", got)
		}
	})

	t.Run("repeated key accumulates sources", func(t *testing.T) {
		r, err := parseFirewallRule("port=22,sources=10.0.0.0/8,sources=192.168.0.0/16")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := strSlice(r.Sources); len(got) != 2 {
			t.Errorf("sources = %v, want 2 elements", got)
		}
	})

	errCases := []struct {
		name string
		in   string
	}{
		{"missing port", "name=https,sources=0.0.0.0/0"},
		{"non-int port", "port=https"},
		{"unknown key", "port=443,colour=red"},
		{"not key=value", "port=443,bogus"},
	}
	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseFirewallRule(tc.in); err == nil {
				t.Errorf("parseFirewallRule(%q) = nil error, want error", tc.in)
			}
		})
	}
}

var _ = api.ClusterFirewallRuleSettings{} // keep api import if unused above
