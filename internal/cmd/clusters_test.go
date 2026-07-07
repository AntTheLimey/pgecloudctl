package cmd

import (
	"encoding/json"
	"reflect"
	"strings"
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

func TestParseClusterNetwork(t *testing.T) {
	str := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}

	tests := []struct {
		name          string
		in            string
		defaultRegion string
		wantErr       bool
		wantRegion    string
		wantCidr      string
		wantPublic    []string
		wantPrivate   []string
	}{
		{
			name:       "public cluster network",
			in:         "region=us-east-1,cidr=10.4.0.0/16,public-subnets=10.4.1.0/24",
			wantRegion: "us-east-1",
			wantCidr:   "10.4.0.0/16",
			wantPublic: []string{"10.4.1.0/24"},
		},
		{
			name: "private cluster network with both subnet kinds",
			in: "region=us-east-1,cidr=10.3.0.0/16," +
				"public-subnets=10.3.1.0/24,private-subnets=10.3.128.0/24",
			wantRegion:  "us-east-1",
			wantCidr:    "10.3.0.0/16",
			wantPublic:  []string{"10.3.1.0/24"},
			wantPrivate: []string{"10.3.128.0/24"},
		},
		{
			name:       "repeated subnets accumulate",
			in:         "region=us-east-1,public-subnets=10.4.1.0/24,public-subnets=10.4.2.0/24",
			wantRegion: "us-east-1",
			wantPublic: []string{"10.4.1.0/24", "10.4.2.0/24"},
		},
		{
			name:          "region defaults on single-region clusters",
			in:            "cidr=10.4.0.0/16",
			defaultRegion: "us-east-1",
			wantRegion:    "us-east-1",
			wantCidr:      "10.4.0.0/16",
		},
		{
			name:    "region required on multi-region clusters",
			in:      "cidr=10.4.0.0/16",
			wantErr: true,
		},
		{
			name:    "unknown key",
			in:      "region=us-east-1,subnet=10.4.1.0/24",
			wantErr: true,
		},
		{
			name:    "not key=value",
			in:      "region=us-east-1,bogus",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := parseClusterNetwork(tt.in, tt.defaultRegion)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseClusterNetwork(%q) = nil error, want error", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n.Region != tt.wantRegion {
				t.Errorf("region = %q, want %q", n.Region, tt.wantRegion)
			}
			if str(n.Cidr) != tt.wantCidr {
				t.Errorf("cidr = %q, want %q", str(n.Cidr), tt.wantCidr)
			}
			if got := strSlice(n.PublicSubnets); !equalStrings(got, tt.wantPublic) {
				t.Errorf("public subnets = %v, want %v", got, tt.wantPublic)
			}
			if got := strSlice(n.PrivateSubnets); !equalStrings(got, tt.wantPrivate) {
				t.Errorf("private subnets = %v, want %v", got, tt.wantPrivate)
			}
		})
	}
}

func TestParseClusterNode(t *testing.T) {
	str := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	num := func(p *int) int {
		if p == nil {
			return 0
		}
		return *p
	}

	tests := []struct {
		name          string
		in            string
		defaultRegion string
		wantErr       string
		wantName      string
		wantRegion    string
		wantInstance  string
		wantSize      int
		wantIops      int
		wantType      string
		wantAZ        string
	}{
		{
			name:         "spec payload node",
			in:           "name=n1,region=us-east-1,instance-type=r7g.medium,volume-size=30",
			wantName:     "n1",
			wantRegion:   "us-east-1",
			wantInstance: "r7g.medium",
			wantSize:     30,
		},
		{
			name: "all keys",
			in: "name=n1,region=us-east-1,instance-type=r7g.large," +
				"volume-size=100,volume-iops=3000,volume-type=gp2," +
				"availability-zone=us-east-1a",
			wantName:     "n1",
			wantRegion:   "us-east-1",
			wantInstance: "r7g.large",
			wantSize:     100,
			wantIops:     3000,
			wantType:     "gp2",
			wantAZ:       "us-east-1a",
		},
		{
			name:          "region defaults on single-region clusters",
			in:            "name=n1,instance-type=r7g.medium",
			defaultRegion: "us-east-1",
			wantName:      "n1",
			wantRegion:    "us-east-1",
			wantInstance:  "r7g.medium",
		},
		{
			name:    "gp3 rejected (CLOUD-480)",
			in:      "name=n1,region=us-east-1,volume-type=gp3",
			wantErr: "gp3",
		},
		{
			name:    "gp3 rejected case-insensitively",
			in:      "name=n1,region=us-east-1,volume-type=GP3",
			wantErr: "gp3",
		},
		{
			name:    "region required on multi-region clusters",
			in:      "name=n1,instance-type=r7g.medium",
			wantErr: "region is required",
		},
		{
			name:    "non-int volume-size",
			in:      "name=n1,region=us-east-1,volume-size=big",
			wantErr: "not an integer",
		},
		{
			name:    "non-int volume-iops",
			in:      "name=n1,region=us-east-1,volume-iops=fast",
			wantErr: "not an integer",
		},
		{
			name:    "unknown key",
			in:      "name=n1,region=us-east-1,colour=red",
			wantErr: "unknown key",
		},
		{
			name:    "not key=value",
			in:      "name=n1,region=us-east-1,bogus",
			wantErr: "not key=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := parseClusterNode(tt.in, tt.defaultRegion)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseClusterNode(%q) = nil error, want error containing %q",
						tt.in, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if str(n.Name) != tt.wantName {
				t.Errorf("name = %q, want %q", str(n.Name), tt.wantName)
			}
			if n.Region != tt.wantRegion {
				t.Errorf("region = %q, want %q", n.Region, tt.wantRegion)
			}
			if str(n.InstanceType) != tt.wantInstance {
				t.Errorf("instance type = %q, want %q", str(n.InstanceType), tt.wantInstance)
			}
			if num(n.VolumeSize) != tt.wantSize {
				t.Errorf("volume size = %d, want %d", num(n.VolumeSize), tt.wantSize)
			}
			if num(n.VolumeIops) != tt.wantIops {
				t.Errorf("volume iops = %d, want %d", num(n.VolumeIops), tt.wantIops)
			}
			if str(n.VolumeType) != tt.wantType {
				t.Errorf("volume type = %q, want %q", str(n.VolumeType), tt.wantType)
			}
			if str(n.AvailabilityZone) != tt.wantAZ {
				t.Errorf("availability zone = %q, want %q", str(n.AvailabilityZone), tt.wantAZ)
			}
		})
	}
}

func TestBuildCreateNodes(t *testing.T) {
	str := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}

	type nodeExp struct {
		name          string
		region        string
		instanceType  string
		volumeSize    int
		volumeSizeNil bool
		volumeTypeNil bool
	}
	tests := []struct {
		name         string
		nodeFlags    []string
		instanceType string
		volumeSize   int
		regions      []string
		wantErr      bool
		wantNil      bool
		want         []nodeExp
	}{
		{
			name:    "nil without node flags or shorthand",
			regions: []string{"us-east-1"},
			wantNil: true,
		},
		{
			name:         "shorthand synthesizes one node per region",
			instanceType: "r7g.medium",
			volumeSize:   30,
			regions:      []string{"us-east-1", "eu-west-1"},
			want: []nodeExp{
				{
					name: "n1", region: "us-east-1",
					instanceType: "r7g.medium", volumeSize: 30,
					volumeTypeNil: true,
				},
				{
					name: "n2", region: "eu-west-1",
					instanceType: "r7g.medium", volumeSize: 30,
					volumeTypeNil: true,
				},
			},
		},
		{
			name:         "instance-type only shorthand",
			instanceType: "r7g.medium",
			regions:      []string{"us-east-1"},
			want: []nodeExp{
				{
					name: "n1", region: "us-east-1",
					instanceType: "r7g.medium", volumeSizeNil: true,
				},
			},
		},
		{
			name: "explicit --node values are parsed",
			nodeFlags: []string{
				"name=n1,instance-type=r7g.medium,volume-size=30",
			},
			regions: []string{"us-east-1"},
			want: []nodeExp{
				{
					name: "n1", region: "us-east-1",
					instanceType: "r7g.medium", volumeSize: 30,
				},
			},
		},
		{
			name:         "--node and shorthand conflict",
			nodeFlags:    []string{"name=n1"},
			instanceType: "r7g.medium",
			regions:      []string{"us-east-1"},
			wantErr:      true,
		},
		{
			name:      "parse errors propagate (gp3)",
			nodeFlags: []string{"volume-type=gp3"},
			regions:   []string{"us-east-1"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := buildCreateNodes(tt.nodeFlags,
				tt.instanceType, tt.volumeSize, tt.regions)
			if tt.wantErr {
				if err == nil {
					t.Fatal("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if nodes != nil {
					t.Errorf("nodes = %v, want nil", nodes)
				}
				return
			}
			if len(nodes) != len(tt.want) {
				t.Fatalf("got %d nodes, want %d", len(nodes), len(tt.want))
			}
			for i, w := range tt.want {
				n := nodes[i]
				if str(n.Name) != w.name {
					t.Errorf("node[%d] name = %q, want %q",
						i, str(n.Name), w.name)
				}
				if n.Region != w.region {
					t.Errorf("node[%d] region = %q, want %q",
						i, n.Region, w.region)
				}
				if w.instanceType != "" &&
					str(n.InstanceType) != w.instanceType {
					t.Errorf("node[%d] instance type = %q, want %q",
						i, str(n.InstanceType), w.instanceType)
				}
				if w.volumeSizeNil {
					if n.VolumeSize != nil {
						t.Errorf("node[%d] volume size = %v, want nil",
							i, *n.VolumeSize)
					}
				} else if n.VolumeSize == nil ||
					*n.VolumeSize != w.volumeSize {
					t.Errorf("node[%d] volume size = %v, want %d",
						i, n.VolumeSize, w.volumeSize)
				}
				if w.volumeTypeNil && n.VolumeType != nil {
					t.Errorf("node[%d] volume type = %q, want unset "+
						"(server defaults to gp2; never gp3 per "+
						"CLOUD-480)", i, str(n.VolumeType))
				}
			}
		})
	}
}

func TestBuildCreateNetworks(t *testing.T) {
	tests := []struct {
		name        string
		flags       []string
		regions     []string
		wantErr     bool
		wantNil     bool
		wantRegions []string
	}{
		{
			name:    "nil without network flags",
			regions: []string{"us-east-1"},
			wantNil: true,
		},
		{
			name: "multiple networks parse in order",
			flags: []string{
				"region=us-east-1,cidr=10.4.0.0/16",
				"region=eu-west-1,cidr=10.5.0.0/16",
			},
			regions:     []string{"us-east-1", "eu-west-1"},
			wantRegions: []string{"us-east-1", "eu-west-1"},
		},
		{
			name:    "parse errors propagate",
			flags:   []string{"cidr=10.4.0.0/16"},
			regions: []string{"us-east-1", "eu-west-1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networks, err := buildCreateNetworks(tt.flags, tt.regions)
			if tt.wantErr {
				if err == nil {
					t.Fatal("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if networks != nil {
					t.Errorf("networks = %v, want nil", networks)
				}
				return
			}
			if len(networks) != len(tt.wantRegions) {
				t.Fatalf("got %d networks, want %d",
					len(networks), len(tt.wantRegions))
			}
			for i, want := range tt.wantRegions {
				if networks[i].Region != want {
					t.Errorf("network[%d] region = %q, want %q",
						i, networks[i].Region, want)
				}
			}
		})
	}
}

// setClusterCreateFlags sets the clusters create flag globals and
// restores the previous values on cleanup.
func setClusterCreateFlags(t *testing.T, name, cloudAccountID string,
	regions []string, nodeLocation string, backupStoreIDs, firewallRules,
	networks, nodes []string) {
	t.Helper()
	prev := []func(){
		func() { clusterCreateName = "" },
		func() { clusterCreateCloudAccountID = "" },
		func() { clusterCreateRegions = nil },
		func() { clusterCreateNodeLocation = "" },
		func() { clusterCreateBackupStoreIDs = nil },
		func() { clusterCreateFirewallRules = nil },
		func() { clusterCreateNetworks = nil },
		func() { clusterCreateNodes = nil },
		func() { clusterCreateInstanceType = "" },
		func() { clusterCreateVolumeSize = 0 },
	}
	t.Cleanup(func() {
		for _, restore := range prev {
			restore()
		}
	})
	clusterCreateName = name
	clusterCreateCloudAccountID = cloudAccountID
	clusterCreateRegions = regions
	clusterCreateNodeLocation = nodeLocation
	clusterCreateBackupStoreIDs = backupStoreIDs
	clusterCreateFirewallRules = firewallRules
	clusterCreateNetworks = networks
	clusterCreateNodes = nodes
}

// TestClusterCreateBodyMatchesServiceSpecs verifies that the CLI can
// reproduce, without raw curl, the reference cluster payloads from the
// pgEdge Cloud test-environment service specs (rulemaster and
// postgrest-cloud-test docs/service-specs.md §1).
func TestClusterCreateBodyMatchesServiceSpecs(t *testing.T) {
	tests := []struct {
		name          string
		clusterName   string
		nodeLocation  string
		firewallRules []string
		networks      []string
		nodes         []string
		wantJSON      string
	}{
		{
			name:         "public cluster (postgrest-pub)",
			clusterName:  "postgrest-pub",
			nodeLocation: "public",
			firewallRules: []string{
				"name=postgres,port=5432,sources=0.0.0.0/0",
				"name=https,port=443,sources=0.0.0.0/0",
			},
			networks: []string{
				"region=us-east-1,cidr=10.4.0.0/16," +
					"public-subnets=10.4.1.0/24",
			},
			nodes: []string{
				"name=n1,region=us-east-1," +
					"instance-type=r7g.medium,volume-size=30",
			},
			wantJSON: `{
				"name": "postgrest-pub",
				"cloud_account_id": "5be8beea-321f-418f-b33f-bea07c89d4ee",
				"backup_store_ids": ["dbd5e3d9-2364-4066-86d5-5b13cc8deaba"],
				"regions": ["us-east-1"],
				"node_location": "public",
				"nodes": [{
					"name": "n1",
					"region": "us-east-1",
					"instance_type": "r7g.medium",
					"volume_size": 30
				}],
				"networks": [{
					"region": "us-east-1",
					"cidr": "10.4.0.0/16",
					"public_subnets": ["10.4.1.0/24"]
				}],
				"firewall_rules": [
					{"name": "postgres", "port": 5432, "sources": ["0.0.0.0/0"]},
					{"name": "https", "port": 443, "sources": ["0.0.0.0/0"]}
				]
			}`,
		},
		{
			name:         "private cluster (postgrest-priv)",
			clusterName:  "postgrest-priv",
			nodeLocation: "private",
			firewallRules: []string{
				"name=postgres,port=5432,sources=0.0.0.0/0",
			},
			networks: []string{
				"region=us-east-1,cidr=10.3.0.0/16," +
					"public-subnets=10.3.1.0/24," +
					"private-subnets=10.3.128.0/24",
			},
			nodes: []string{
				"name=n1,region=us-east-1," +
					"instance-type=r7g.medium,volume-size=30",
			},
			wantJSON: `{
				"name": "postgrest-priv",
				"cloud_account_id": "5be8beea-321f-418f-b33f-bea07c89d4ee",
				"backup_store_ids": ["dbd5e3d9-2364-4066-86d5-5b13cc8deaba"],
				"regions": ["us-east-1"],
				"node_location": "private",
				"nodes": [{
					"name": "n1",
					"region": "us-east-1",
					"instance_type": "r7g.medium",
					"volume_size": 30
				}],
				"networks": [{
					"region": "us-east-1",
					"cidr": "10.3.0.0/16",
					"public_subnets": ["10.3.1.0/24"],
					"private_subnets": ["10.3.128.0/24"]
				}],
				"firewall_rules": [{
					"name": "postgres",
					"port": 5432,
					"sources": ["0.0.0.0/0"]
				}]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setClusterCreateFlags(t, tt.clusterName,
				"5be8beea-321f-418f-b33f-bea07c89d4ee",
				[]string{"us-east-1"}, tt.nodeLocation,
				[]string{"dbd5e3d9-2364-4066-86d5-5b13cc8deaba"},
				tt.firewallRules, tt.networks, tt.nodes)

			body, err := buildClusterCreateBody()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			var gotAny, wantAny any
			if err := json.Unmarshal(got, &gotAny); err != nil {
				t.Fatalf("unmarshal got: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantJSON), &wantAny); err != nil {
				t.Fatalf("unmarshal want: %v", err)
			}
			if !reflect.DeepEqual(gotAny, wantAny) {
				t.Errorf("payload mismatch\ngot:  %s\nwant: %s",
					got, tt.wantJSON)
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
