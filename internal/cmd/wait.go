package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/spf13/cobra"
)

// Service-mutation wait flags. These are shared across all service-mutating
// commands (mcp/rag deploy+update, services remove); only one such command
// runs per process invocation, so a single set of package-level vars is safe.
var (
	svcWait         bool
	svcWaitTimeout  int
	svcWaitInterval int
)

// addServiceWaitFlags registers --wait/--timeout/--interval on a
// service-mutating command. Service mutations are asynchronous: the API
// accepts the request and spawns a task, so without --wait the command exits
// as soon as the request is accepted, not when the work completes.
func addServiceWaitFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&svcWait, "wait", false,
		"Wait for the operation's task to reach a terminal state")
	cmd.Flags().IntVar(&svcWaitTimeout, "timeout", 300,
		"Max seconds to wait when --wait is set")
	cmd.Flags().IntVar(&svcWaitInterval, "interval", 5,
		"Polling interval in seconds when --wait is set")
}

// newestSubjectTaskID returns the id of the most recent task for subjectID, or
// "" if the subject has no tasks. Tasks are ranked by created_at, which is an
// RFC3339 timestamp and therefore lexicographically sortable.
func newestSubjectTaskID(
	client *api.ClientWithResponses, subjectID string,
) (string, error) {
	sid := subjectID
	resp, err := client.ListTasksWithResponse(
		context.Background(), &api.ListTasksParams{SubjectId: &sid},
	)
	if err != nil {
		return "", fmt.Errorf("list tasks: %w", err)
	}
	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return "", err
	}

	tasks := resp.JSON200
	if tasks == nil || len(*tasks) == 0 {
		return "", nil
	}

	newest := (*tasks)[0]
	for _, t := range *tasks {
		if t.CreatedAt > newest.CreatedAt {
			newest = t
		}
	}
	return newest.Id, nil
}

// getTaskByID fetches a single task by id.
func getTaskByID(
	client *api.ClientWithResponses, taskID string,
) (*api.Task, error) {
	resp, err := client.ListTasksWithResponse(
		context.Background(), &api.ListTasksParams{Id: &taskID},
	)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return nil, err
	}

	tasks := resp.JSON200
	if tasks == nil || len(*tasks) == 0 {
		return nil, &ExitError{
			msg:  fmt.Sprintf("task %q not found", taskID),
			code: ExitNotFound,
		}
	}
	t := (*tasks)[0]
	return &t, nil
}

// waitForSubjectTask discovers the task spawned by a service mutation on
// subjectID and polls it until it reaches a terminal state.
//
// priorTaskID is the newest task for the subject captured *before* the
// mutation; the newly-created task is the first task whose id differs from it.
// The mutation request returns no task id and the task takes a moment to
// appear, so discovery and polling share a single deadline.
//
// Returns nil when the task succeeds, an ExitError with ExitGeneral when it
// fails, and an ExitError with ExitTimeout if the deadline passes (whether or
// not a task was ever discovered).
func waitForSubjectTask(
	cmd *cobra.Command,
	client *api.ClientWithResponses,
	subjectID, priorTaskID string,
	timeout, interval int,
) error {
	out := cmd.OutOrStdout()
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	iv := time.Duration(interval) * time.Second

	taskID := ""
	for {
		if taskID == "" {
			newest, err := newestSubjectTaskID(client, subjectID)
			if err != nil {
				return err
			}
			if newest != "" && newest != priorTaskID {
				taskID = newest
				fmt.Fprintf(out, "Tracking task %s...\n", taskID)
			}
		} else {
			t, err := getTaskByID(client, taskID)
			if err != nil {
				return err
			}
			switch t.Status {
			case "succeeded":
				fmt.Fprintf(out, "Task %s: succeeded.\n", t.Id)
				return nil
			case "failed":
				msg := fmt.Sprintf("task %s failed", t.Id)
				if t.Error != nil && *t.Error != "" {
					msg = fmt.Sprintf("task %s failed: %s", t.Id, *t.Error)
				}
				return &ExitError{msg: msg, code: ExitGeneral}
			default:
				fmt.Fprintf(out, "Task %s: %s...\n", t.Id, t.Status)
			}
		}

		if time.Now().After(deadline) {
			if taskID == "" {
				return &ExitError{
					msg: fmt.Sprintf(
						"timed out after %ds waiting for a task on %s",
						timeout, subjectID),
					code: ExitTimeout,
				}
			}
			return &ExitError{
				msg: fmt.Sprintf(
					"timed out after %ds waiting for task %s",
					timeout, taskID),
				code: ExitTimeout,
			}
		}

		time.Sleep(iv)
	}
}

// trackServiceMutation reports how to monitor an accepted service mutation on
// subjectID, and — when --wait is set — blocks until the spawned task reaches a
// terminal state. priorTaskID must be the newest task for the subject captured
// before the mutation (only meaningful when waiting).
func trackServiceMutation(
	cmd *cobra.Command,
	client *api.ClientWithResponses,
	subjectID, priorTaskID string,
) error {
	if svcWait {
		return waitForSubjectTask(
			cmd, client, subjectID, priorTaskID,
			svcWaitTimeout, svcWaitInterval,
		)
	}
	fmt.Fprintf(cmd.OutOrStdout(),
		"Monitor with: pgecloudctl tasks list --subject-id %s\n", subjectID)
	return nil
}
