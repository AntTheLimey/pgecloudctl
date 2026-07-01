package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/google/uuid"
)

// resolveIDPrefix returns the single ID in ids equal to or prefixed by
// input. Errors if input matches zero or more than one ID. kind names
// the resource for error messages (e.g. "cluster").
func resolveIDPrefix(input string, ids []string, kind string) (string, error) {
	// Match case-insensitively: UUIDs are canonical lowercase but the
	// full-UUID path (uuid.Parse) accepts any case, so prefix matching
	// must too, or "A1B2" would fail to match "a1b2..." inconsistently.
	lower := strings.ToLower(input)
	var matches []string
	for _, id := range ids {
		if strings.HasPrefix(strings.ToLower(id), lower) {
			matches = append(matches, id)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", &ExitError{
			msg:  fmt.Sprintf("no %s matches ID prefix %q", kind, input),
			code: ExitNotFound,
		}
	default:
		return "", &ExitError{
			msg: fmt.Sprintf("ambiguous %s ID prefix %q matches: %s",
				kind, input, strings.Join(matches, ", ")),
			code: ExitGeneral,
		}
	}
}

// resolveClusterID returns the cluster UUID for a full UUID or a unique
// ID prefix. A full UUID is returned without any API call.
func resolveClusterID(ctx context.Context, client *api.ClientWithResponses,
	input string) (uuid.UUID, error) {
	if id, err := uuid.Parse(input); err == nil {
		return id, nil
	}
	resp, err := client.ListClustersWithResponse(ctx, &api.ListClustersParams{})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("list clusters: %w", err)
	}
	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return uuid.UUID{}, err
	}
	var ids []string
	if resp.JSON200 != nil {
		for _, c := range *resp.JSON200 {
			ids = append(ids, c.Id)
		}
	}
	matched, err := resolveIDPrefix(input, ids, "cluster")
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(matched)
}

// resolveDatabaseID returns the database UUID for a full UUID or a
// unique ID prefix. A full UUID is returned without any API call.
func resolveDatabaseID(ctx context.Context, client *api.ClientWithResponses,
	input string) (uuid.UUID, error) {
	if id, err := uuid.Parse(input); err == nil {
		return id, nil
	}
	resp, err := client.ListDatabasesWithResponse(ctx, &api.ListDatabasesParams{})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("list databases: %w", err)
	}
	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return uuid.UUID{}, err
	}
	var ids []string
	if resp.JSON200 != nil {
		for _, d := range *resp.JSON200 {
			ids = append(ids, d.Id)
		}
	}
	matched, err := resolveIDPrefix(input, ids, "database")
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(matched)
}
