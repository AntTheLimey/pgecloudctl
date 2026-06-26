package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/spf13/cobra"
)

// taskListLookback bounds how many recent tasks newestSubjectTaskID requests.
// The API returns tasks newest-first, so the newest task for a subject is
// always within the first page; an explicit limit guards against the server's
// default page size silently truncating it.
const taskListLookback = 100

// Wait flags. Shared across every asynchronous command (create/delete of
// task-backed resources, mcp/rag deploy+update, services remove); only one
// such command runs per process invocation, so a single set of package-level
// vars is safe.
var (
	waitFlag         bool
	waitTimeoutFlag  int
	waitIntervalFlag int
)

// addWaitFlags registers --wait/--timeout/--interval on an asynchronous
// command. These operations have the API accept the request and spawn a task,
// so without --wait the command exits as soon as the request is accepted, not
// when the work completes.
func addWaitFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&waitFlag, "wait", false,
		"Wait for the operation's task to reach a terminal state")
	cmd.Flags().IntVar(&waitTimeoutFlag, "timeout", 300,
		"Max seconds to wait when --wait is set")
	cmd.Flags().IntVar(&waitIntervalFlag, "interval", 5,
		"Polling interval in seconds when --wait is set")
}

// newestSubjectTaskID returns the id of the most recent task for subjectID, or
// "" if the subject has no tasks. Tasks are ranked by created_at, which is an
// RFC3339 timestamp and therefore lexicographically sortable.
func newestSubjectTaskID(
	ctx context.Context, client *api.ClientWithResponses, subjectID string,
) (string, error) {
	sid := subjectID
	limit := taskListLookback
	resp, err := client.ListTasksWithResponse(
		ctx, &api.ListTasksParams{SubjectId: &sid, Limit: &limit},
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
	ctx context.Context, client *api.ClientWithResponses, taskID string,
) (*api.Task, error) {
	resp, err := client.ListTasksWithResponse(
		ctx, &api.ListTasksParams{Id: &taskID},
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
// appear, so discovery and polling share a single deadline. Each poll request
// is bounded by that deadline, so a hung request cannot outlive --timeout.
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
	if interval < 1 {
		interval = 1
	}
	iv := time.Duration(interval) * time.Second

	taskID := ""
	for {
		if time.Now().After(deadline) {
			return timeoutError(timeout, subjectID, taskID)
		}

		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		var stepErr error
		if taskID == "" {
			newest, err := newestSubjectTaskID(ctx, client, subjectID)
			stepErr = err
			if err == nil && newest != "" && newest != priorTaskID {
				taskID = newest
				fmt.Fprintf(out, "Tracking task %s...\n", taskID)
			}
		} else {
			t, err := getTaskByID(ctx, client, taskID)
			stepErr = err
			if err == nil {
				switch t.Status {
				case "succeeded":
					cancel()
					fmt.Fprintf(out, "Task %s: succeeded.\n", t.Id)
					return nil
				case "failed":
					cancel()
					msg := fmt.Sprintf("task %s failed", t.Id)
					if t.Error != nil && *t.Error != "" {
						msg = fmt.Sprintf("task %s failed: %s", t.Id, *t.Error)
					}
					return &ExitError{msg: msg, code: ExitGeneral}
				default:
					fmt.Fprintf(out, "Task %s: %s...\n", t.Id, t.Status)
				}
			}
		}
		cancel()

		if stepErr != nil {
			if errors.Is(stepErr, context.DeadlineExceeded) {
				return timeoutError(timeout, subjectID, taskID)
			}
			return stepErr
		}

		time.Sleep(iv)
	}
}

// timeoutError builds the ExitTimeout returned when waiting exceeds --timeout.
// The message names the task if one was discovered, otherwise the subject.
func timeoutError(timeout int, subjectID, taskID string) error {
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
			"timed out after %ds waiting for task %s", timeout, taskID),
		code: ExitTimeout,
	}
}

// trackMutation handles the asynchronous tail of a command. When --wait is
// set it blocks until the spawned task reaches a terminal state. Otherwise, in
// table output, it prints how to monitor the task; in machine output (json or
// yaml) it stays silent so stdout remains parseable.
//
// priorTaskID must be the newest task for the subject captured before the
// mutation (only meaningful when waiting). For a freshly created resource it
// is "", since the new resource has no prior tasks.
func trackMutation(
	cmd *cobra.Command,
	client *api.ClientWithResponses,
	subjectID, priorTaskID string,
) error {
	if waitFlag {
		return waitForSubjectTask(
			cmd, client, subjectID, priorTaskID,
			waitTimeoutFlag, waitIntervalFlag,
		)
	}
	if flagOutput == "table" {
		fmt.Fprintf(cmd.OutOrStdout(),
			"Monitor with: pgecloudctl tasks list --subject-id %s\n", subjectID)
	}
	return nil
}
