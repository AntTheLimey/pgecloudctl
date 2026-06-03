package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/output"
)

type testRow struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (r testRow) Columns() []string {
	return []string{r.ID, r.Name, r.Status}
}

func TestPrintJSON(t *testing.T) {
	rows := []testRow{
		{ID: "abc123", Name: "my-db", Status: "active"},
		{ID: "def456", Name: "other-db", Status: "stopped"},
	}

	var buf bytes.Buffer
	err := output.Print(&buf, "json", rows, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []testRow
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}
	if got[0].ID != "abc123" || got[0].Name != "my-db" || got[0].Status != "active" {
		t.Errorf("row 0 mismatch: %+v", got[0])
	}
	if got[1].ID != "def456" || got[1].Name != "other-db" || got[1].Status != "stopped" {
		t.Errorf("row 1 mismatch: %+v", got[1])
	}
}

func TestPrintTable(t *testing.T) {
	rows := []output.Row{
		testRow{ID: "abc123", Name: "my-db", Status: "active"},
		testRow{ID: "def456", Name: "other-db", Status: "stopped"},
	}

	headers := []string{"ID", "NAME", "STATUS"}

	var buf bytes.Buffer
	err := output.Print(&buf, "table", rows, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()

	for _, h := range headers {
		if !strings.Contains(out, h) {
			t.Errorf("output missing header %q", h)
		}
	}

	values := []string{"abc123", "my-db", "active", "def456", "other-db", "stopped"}
	for _, v := range values {
		if !strings.Contains(out, v) {
			t.Errorf("output missing value %q", v)
		}
	}
}

func TestPrintJSON_SingleObject(t *testing.T) {
	row := testRow{ID: "xyz789", Name: "solo-db", Status: "creating"}

	var buf bytes.Buffer
	err := output.Print(&buf, "json", row, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got testRow
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if got.ID != "xyz789" || got.Name != "solo-db" || got.Status != "creating" {
		t.Errorf("single object mismatch: %+v", got)
	}
}

func TestPrintYAML(t *testing.T) {
	rows := []testRow{
		{ID: "abc123", Name: "my-db", Status: "active"},
	}

	var buf bytes.Buffer
	err := output.Print(&buf, "yaml", rows, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "id: abc123") {
		t.Errorf("output missing 'id: abc123': %q", out)
	}
	if !strings.Contains(out, "name: my-db") {
		t.Errorf("output missing 'name: my-db': %q", out)
	}
}

func TestPrintUnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	err := output.Print(&buf, "xml", []testRow{}, nil)
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Errorf("unexpected error message: %v", err)
	}
}
