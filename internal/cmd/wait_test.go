package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/api"
	"github.com/spf13/cobra"
)

// tasksMux returns a handler that serves ListTasks requests. Requests with an
// "id" query are answered by byID; all others (subject_id filters) return
// subjectTasks.
func tasksMux(
	subjectTasks []api.Task,
	byID func(id string) (api.Task, bool),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if id := r.URL.Query().Get("id"); id != "" {
			if t, ok := byID(id); ok {
				_ = json.NewEncoder(w).Encode([]api.Task{t})
				return
			}
			_ = json.NewEncoder(w).Encode([]api.Task{})
			return
		}
		_ = json.NewEncoder(w).Encode(subjectTasks)
	}
}

func waitTestCmd() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	c := &cobra.Command{}
	c.SetOut(buf)
	return c, buf
}

func TestNewestSubjectTaskID(t *testing.T) {
	tests := []struct {
		name  string
		tasks []api.Task
		want  string
	}{
		{
			name:  "no tasks",
			tasks: []api.Task{},
			want:  "",
		},
		{
			name: "picks latest by created_at regardless of order",
			tasks: []api.Task{
				{Id: "b", CreatedAt: "2026-06-25T01:00:00Z"},
				{Id: "c", CreatedAt: "2026-06-25T03:00:00Z"},
				{Id: "a", CreatedAt: "2026-06-25T02:00:00Z"},
			},
			want: "c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(t, tasksMux(tt.tasks, nil))
			got, err := newestSubjectTaskID(client, "subj")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWaitForSubjectTask_Succeeds(t *testing.T) {
	subjectTasks := []api.Task{
		{Id: "new-task", CreatedAt: "2026-06-25T02:00:00Z"},
		{Id: "old-task", CreatedAt: "2026-06-25T01:00:00Z"},
	}
	calls := 0
	byID := func(id string) (api.Task, bool) {
		calls++
		status := "running"
		if calls >= 2 {
			status = "succeeded"
		}
		return api.Task{Id: id, Status: status}, true
	}

	client := newTestClient(t, tasksMux(subjectTasks, byID))
	cmd, buf := waitTestCmd()

	err := waitForSubjectTask(cmd, client, "subj", "old-task", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Tracking task new-task") {
		t.Errorf("expected discovery line, got: %q", out)
	}
	if !strings.Contains(out, "Task new-task: succeeded") {
		t.Errorf("expected success line, got: %q", out)
	}
}

func TestWaitForSubjectTask_Fails(t *testing.T) {
	subjectTasks := []api.Task{
		{Id: "new-task", CreatedAt: "2026-06-25T02:00:00Z"},
		{Id: "old-task", CreatedAt: "2026-06-25T01:00:00Z"},
	}
	errMsg := "node unreachable"
	byID := func(id string) (api.Task, bool) {
		return api.Task{Id: id, Status: "failed", Error: &errMsg}, true
	}

	client := newTestClient(t, tasksMux(subjectTasks, byID))
	cmd, _ := waitTestCmd()

	err := waitForSubjectTask(cmd, client, "subj", "old-task", 10, 0)
	if err == nil {
		t.Fatal("expected error for failed task")
	}
	ee, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T", err)
	}
	if ee.Code() != ExitGeneral {
		t.Errorf("exit code = %d, want %d", ee.Code(), ExitGeneral)
	}
	if !strings.Contains(ee.Error(), errMsg) {
		t.Errorf("expected error detail %q in %q", errMsg, ee.Error())
	}
}

func TestWaitForSubjectTask_TimesOutWhenNoNewTask(t *testing.T) {
	// Only the prior task is ever present, so no new task is discovered.
	subjectTasks := []api.Task{
		{Id: "old-task", CreatedAt: "2026-06-25T01:00:00Z"},
	}

	client := newTestClient(t, tasksMux(subjectTasks, nil))
	cmd, _ := waitTestCmd()

	err := waitForSubjectTask(cmd, client, "subj", "old-task", 0, 0)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	ee, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T", err)
	}
	if ee.Code() != ExitTimeout {
		t.Errorf("exit code = %d, want %d", ee.Code(), ExitTimeout)
	}
}
