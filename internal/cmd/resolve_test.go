package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
)

func TestResolveIDPrefix(t *testing.T) {
	ids := []string{
		"a1b2c3d4-0000-0000-0000-000000000001",
		"a1ffffff-0000-0000-0000-000000000002",
		"b9000000-0000-0000-0000-000000000003",
	}
	t.Run("unique prefix", func(t *testing.T) {
		got, err := resolveIDPrefix("b9", ids, "cluster")
		if err != nil || got != ids[2] {
			t.Fatalf("got (%q, %v), want (%q, nil)", got, err, ids[2])
		}
	})
	t.Run("ambiguous prefix", func(t *testing.T) {
		if _, err := resolveIDPrefix("a1", ids, "cluster"); err == nil {
			t.Errorf("expected ambiguous error")
		}
	})
	t.Run("no match", func(t *testing.T) {
		if _, err := resolveIDPrefix("zz", ids, "cluster"); err == nil {
			t.Errorf("expected not-found error")
		}
	})
}

func TestResolveClusterID(t *testing.T) {
	full := "a1b2c3d4-1111-2222-3333-444455556666"
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.Cluster{{Id: full}})
	}

	t.Run("full uuid skips API", func(t *testing.T) {
		// nil client proves no list call is made for a full UUID.
		got, err := resolveClusterID(context.Background(), nil, full)
		if err != nil || got.String() != full {
			t.Fatalf("got (%q, %v), want (%q, nil)", got, err, full)
		}
	})

	t.Run("prefix resolves via list", func(t *testing.T) {
		client := newTestClient(t, handler)
		got, err := resolveClusterID(context.Background(), client, "a1b2")
		if err != nil || got.String() != full {
			t.Fatalf("got (%q, %v), want (%q, nil)", got, err, full)
		}
	})
}

func TestResolveDatabaseID(t *testing.T) {
	full := "b2c3d4e5-1111-2222-3333-444455556666"
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]api.Database{{Id: full}})
	}

	t.Run("full uuid skips API", func(t *testing.T) {
		// nil client proves no list call is made for a full UUID.
		got, err := resolveDatabaseID(context.Background(), nil, full)
		if err != nil || got.String() != full {
			t.Fatalf("got (%q, %v), want (%q, nil)", got, err, full)
		}
	})

	t.Run("prefix resolves via list", func(t *testing.T) {
		client := newTestClient(t, handler)
		got, err := resolveDatabaseID(context.Background(), client, "b2c3")
		if err != nil || got.String() != full {
			t.Fatalf("got (%q, %v), want (%q, nil)", got, err, full)
		}
	})
}
