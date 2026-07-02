package cmd

import (
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
)

func TestBackupStoreWarning(t *testing.T) {
	if w := backupStoreWarning([]string{"bs-abc"}); w != "" {
		t.Errorf("with a store, want empty warning, got %q", w)
	}
	if w := backupStoreWarning(nil); w == "" {
		t.Error("without a store, want a non-empty warning, got empty")
	}
}

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
		r, err := parseFirewallRule("name=ssh,port=22,sources=10.0.0.0/8,sources=192.168.0.0/16")
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
		{"non-int port", "name=https,port=https"},
		{"unknown key", "name=https,port=443,colour=red"},
		{"not key=value", "name=https,port=443,bogus"},
		{"missing name", "port=443,sources=0.0.0.0/0"},
		{"invalid name", "name=pg,port=5432,sources=0.0.0.0/0"},
	}
	for _, tc := range errCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseFirewallRule(tc.in); err == nil {
				t.Errorf("parseFirewallRule(%q) = nil error, want error", tc.in)
			}
		})
	}
}

func TestBuildClusterUpdate(t *testing.T) {
	existingRule := api.ClusterFirewallRuleSettings{Port: 22}
	c := &api.Cluster{
		Regions:        []string{"us-east-1"},
		FirewallRules:  &[]api.ClusterFirewallRuleSettings{existingRule},
		BackupStoreIds: &[]string{"bs-existing"},
	}
	newRule := api.ClusterFirewallRuleSettings{Port: 443}

	t.Run("appends rules and stores, keeps regions", func(t *testing.T) {
		in := buildClusterUpdate(c,
			[]api.ClusterFirewallRuleSettings{newRule},
			[]string{"bs-new"}, nil)
		if in.FirewallRules == nil || len(*in.FirewallRules) != 2 {
			t.Fatalf("firewall rules = %v, want 2", in.FirewallRules)
		}
		if in.BackupStoreIds == nil || len(*in.BackupStoreIds) != 2 {
			t.Fatalf("backup stores = %v, want 2", in.BackupStoreIds)
		}
		if len(in.Regions) != 1 || in.Regions[0] != "us-east-1" {
			t.Errorf("regions = %v, want [us-east-1]", in.Regions)
		}
	})

	t.Run("regions override when supplied", func(t *testing.T) {
		in := buildClusterUpdate(c, nil, nil, []string{"eu-west-1"})
		if len(in.Regions) != 1 || in.Regions[0] != "eu-west-1" {
			t.Errorf("regions = %v, want [eu-west-1]", in.Regions)
		}
	})
}

func TestClusterCreateFirewallParsing(t *testing.T) {
	// Mirrors the loop in runClustersCreate.
	raw := []string{"name=https,port=443,sources=0.0.0.0/0"}
	var rules []api.ClusterFirewallRuleSettings
	for _, s := range raw {
		r, err := parseFirewallRule(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rules = append(rules, r)
	}
	if len(rules) != 1 || rules[0].Port != 443 {
		t.Fatalf("rules = %v, want one rule on port 443", rules)
	}
}
