package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/google/uuid"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *api.ClientWithResponses {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := api.NewClientWithResponses(srv.URL)
	if err != nil {
		t.Fatalf("create test client: %v", err)
	}
	return client
}

func nodesHandler(nodes []api.ClusterNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nodes)
	}
}

func TestResolveHostIDs_SingleNodeAutoSelect(t *testing.T) {
	client := newTestClient(t, nodesHandler([]api.ClusterNode{
		{Id: "aaa-111", Name: "n1"},
	}))

	ids, err := resolveHostIDs(client, uuid.New(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 1 || ids[0] != "aaa-111" {
		t.Errorf("got %v, want [aaa-111]", ids)
	}
}

func TestResolveHostIDs_MultiNodeRequiresFlag(t *testing.T) {
	client := newTestClient(t, nodesHandler([]api.ClusterNode{
		{Id: "aaa-111", Name: "n1"},
		{Id: "bbb-222", Name: "n2"},
	}))

	_, err := resolveHostIDs(client, uuid.New(), nil)
	if err == nil {
		t.Fatal("expected error for multi-node without --target-nodes")
	}
	ee, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T", err)
	}
	if ee.Code() != ExitGeneral {
		t.Errorf("exit code = %d, want %d", ee.Code(), ExitGeneral)
	}
}

func TestResolveHostIDs_NameResolution(t *testing.T) {
	client := newTestClient(t, nodesHandler([]api.ClusterNode{
		{Id: "aaa-111", Name: "n1"},
		{Id: "bbb-222", Name: "n2"},
		{Id: "ccc-333", Name: "n3"},
	}))

	ids, err := resolveHostIDs(client, uuid.New(), []string{"n2", "n1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != "bbb-222" || ids[1] != "aaa-111" {
		t.Errorf("got %v, want [bbb-222 aaa-111]", ids)
	}
}

func TestResolveHostIDs_InvalidName(t *testing.T) {
	client := newTestClient(t, nodesHandler([]api.ClusterNode{
		{Id: "aaa-111", Name: "n1"},
	}))

	_, err := resolveHostIDs(client, uuid.New(), []string{"n99"})
	if err == nil {
		t.Fatal("expected error for invalid node name")
	}
	ee, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T", err)
	}
	if ee.Code() != ExitGeneral {
		t.Errorf("exit code = %d, want %d", ee.Code(), ExitGeneral)
	}
}

// Ensure context is used properly (compile-time check).
var _ = context.Background
