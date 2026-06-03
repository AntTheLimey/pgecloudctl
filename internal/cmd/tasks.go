package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/spf13/cobra"
)

// tasks list flags.
var (
	taskListLimit       int
	taskListOffset      int
	taskListSubjectID   string
	taskListSubjectKind string
	taskListStatus      string
)

// tasks wait flags.
var (
	taskWaitTimeout  int
	taskWaitInterval int
)

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksGetCmd)
	tasksCmd.AddCommand(tasksWaitCmd)

	// list flags
	tasksListCmd.Flags().IntVar(&taskListLimit, "limit", 0,
		"Maximum number of results to return")
	tasksListCmd.Flags().IntVar(&taskListOffset, "offset", 0,
		"Offset into the results for pagination")
	tasksListCmd.Flags().StringVar(&taskListSubjectID, "subject-id", "",
		"Filter by subject ID")
	tasksListCmd.Flags().StringVar(&taskListSubjectKind, "subject-kind", "",
		"Filter by subject kind")
	tasksListCmd.Flags().StringVar(&taskListStatus, "status", "",
		"Filter by status (queued, running, succeeded, failed)")

	// wait flags
	tasksWaitCmd.Flags().IntVar(&taskWaitTimeout, "timeout", 300,
		"Maximum seconds to wait for task completion")
	tasksWaitCmd.Flags().IntVar(&taskWaitInterval, "interval", 5,
		"Polling interval in seconds")
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage pgEdge Cloud tasks",
}

// --- list ---

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE:  runTasksList,
}

func runTasksList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}

	params := &api.ListTasksParams{}
	if taskListLimit > 0 {
		params.Limit = &taskListLimit
	}
	if taskListOffset > 0 {
		params.Offset = &taskListOffset
	}
	if taskListSubjectID != "" {
		params.SubjectId = &taskListSubjectID
	}
	if taskListSubjectKind != "" {
		params.SubjectKind = &taskListSubjectKind
	}
	if taskListStatus != "" {
		s := api.ListTasksParamsStatus(taskListStatus)
		params.Status = &s
	}

	resp, err := client.ListTasksWithResponse(context.Background(), params)
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, resp.JSON200, nil)
	}

	tasks := resp.JSON200
	if tasks == nil || len(*tasks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tasks found.")
		return nil
	}

	rows := make([]output.Row, 0, len(*tasks))
	for _, t := range *tasks {
		rows = append(rows, taskRow(t))
	}

	headers := []string{"ID", "NAME", "STATUS", "SUBJECT", "CREATED"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- get ---

var tasksGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get task details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTasksGet,
}

func runTasksGet(cmd *cobra.Command, args []string) error {
	taskID := args[0]

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	params := &api.ListTasksParams{Id: &taskID}
	resp, err := client.ListTasksWithResponse(context.Background(), params)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
		return err
	}

	tasks := resp.JSON200
	if tasks == nil || len(*tasks) == 0 {
		return &ExitError{
			msg:  fmt.Sprintf("task %q not found", taskID),
			code: ExitNotFound,
		}
	}

	t := (*tasks)[0]

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, &t, nil)
	}

	rows := []output.Row{taskRow(t)}
	headers := []string{"ID", "NAME", "STATUS", "SUBJECT", "CREATED"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

// --- wait ---

var tasksWaitCmd = &cobra.Command{
	Use:   "wait <id>",
	Short: "Wait for a task to complete",
	Long: `Poll a task until it reaches a terminal state (succeeded or failed).

Exit codes:
  0  task succeeded
  1  task failed or other error
  3  timeout exceeded`,
	Args: cobra.ExactArgs(1),
	RunE: runTasksWait,
}

func runTasksWait(cmd *cobra.Command, args []string) error {
	taskID := args[0]

	client, err := newAPIClient()
	if err != nil {
		return err
	}

	deadline := time.Now().Add(time.Duration(taskWaitTimeout) * time.Second)
	interval := time.Duration(taskWaitInterval) * time.Second

	for {
		params := &api.ListTasksParams{Id: &taskID}
		resp, err := client.ListTasksWithResponse(context.Background(), params)
		if err != nil {
			return fmt.Errorf("poll task: %w", err)
		}

		if err := checkResponse(resp.StatusCode(), string(resp.Body)); err != nil {
			return err
		}

		tasks := resp.JSON200
		if tasks == nil || len(*tasks) == 0 {
			return &ExitError{
				msg:  fmt.Sprintf("task %q not found", taskID),
				code: ExitNotFound,
			}
		}

		t := (*tasks)[0]

		switch t.Status {
		case "succeeded":
			if flagOutput != "table" {
				return output.Print(cmd.OutOrStdout(), flagOutput, &t, nil)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Task %s: %s\n",
				t.Id, t.Status)
			return nil

		case "failed":
			if flagOutput != "table" {
				_ = output.Print(cmd.OutOrStdout(), flagOutput, &t, nil)
			}
			msg := fmt.Sprintf("task %s failed", t.Id)
			if t.Error != nil && *t.Error != "" {
				msg = fmt.Sprintf("task %s failed: %s",
					t.Id, *t.Error)
			}
			return &ExitError{msg: msg, code: ExitGeneral}

		default:
			// queued or running — show progress in table mode
			if flagOutput == "table" {
				fmt.Fprintf(cmd.OutOrStdout(), "Task %s: %s...\n",
					t.Id, t.Status)
			}
		}

		if time.Now().Add(interval).After(deadline) {
			return &ExitError{
				msg: fmt.Sprintf(
					"timed out after %ds waiting for task %s (last status: %s)",
					taskWaitTimeout, taskID, t.Status),
				code: ExitTimeout,
			}
		}

		time.Sleep(interval)
	}
}

// --- row adapter ---

type taskRowData struct {
	id, name, status, subject, created string
}

func (r taskRowData) Columns() []string {
	return []string{r.id, r.name, output.ColorStatus(r.status), r.subject, r.created}
}

// taskRow converts an api.Task into a taskRowData for table output.
func taskRow(t api.Task) taskRowData {
	subject := fmt.Sprintf("%s/%s", t.SubjectKind, t.SubjectId)
	return taskRowData{
		id:      t.Id,
		name:    t.Name,
		status:  t.Status,
		subject: subject,
		created: formatTime(t.CreatedAt),
	}
}
