package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/google/uuid"
)

func resolveHostIDs(client *api.ClientWithResponses, clusterID uuid.UUID, nodeNames []string) ([]string, error) {
	resp, err := client.ListClusterNodesWithResponse(
		context.Background(), clusterID, &api.ListClusterNodesParams{})
	if err != nil {
		return nil, fmt.Errorf("list cluster nodes: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return nil, err
	}

	nodes := resp.JSON200
	if nodes == nil || len(*nodes) == 0 {
		return nil, &ExitError{
			msg:  fmt.Sprintf("cluster %s has no nodes", clusterID),
			code: ExitGeneral,
		}
	}

	nodeList := *nodes

	if len(nodeNames) == 0 {
		if len(nodeList) == 1 {
			return []string{nodeList[0].Id}, nil
		}
		names := make([]string, len(nodeList))
		for i, n := range nodeList {
			names[i] = n.Name
		}
		return nil, &ExitError{
			msg: fmt.Sprintf(
				"cluster has %d nodes (%s) — specify --target-nodes",
				len(nodeList), strings.Join(names, ", ")),
			code: ExitGeneral,
		}
	}

	nameToID := make(map[string]string, len(nodeList))
	for _, n := range nodeList {
		nameToID[n.Name] = n.Id
	}

	hostIDs := make([]string, 0, len(nodeNames))
	for _, name := range nodeNames {
		id, ok := nameToID[name]
		if !ok {
			valid := make([]string, 0, len(nameToID))
			for k := range nameToID {
				valid = append(valid, k)
			}
			return nil, &ExitError{
				msg: fmt.Sprintf(
					"node %q not found in cluster — valid names: %s",
					name, strings.Join(valid, ", ")),
				code: ExitGeneral,
			}
		}
		hostIDs = append(hostIDs, id)
	}

	return hostIDs, nil
}
