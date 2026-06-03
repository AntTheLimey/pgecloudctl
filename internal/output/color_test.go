package output_test

import (
	"testing"

	"github.com/AntTheLimey/pgecloudctl/internal/output"
)

func TestBold(t *testing.T) {
	output.ColorEnabled = true
	got := output.Bold("HEADER")
	want := "\033[1mHEADER\033[0m"
	if got != want {
		t.Errorf("Bold() = %q, want %q", got, want)
	}
}

func TestBold_NoColor(t *testing.T) {
	output.ColorEnabled = false
	got := output.Bold("HEADER")
	if got != "HEADER" {
		t.Errorf("Bold() with color disabled = %q, want %q", got, "HEADER")
	}
}

func TestColorStatus(t *testing.T) {
	output.ColorEnabled = true

	tests := []struct {
		status string
		color  string
	}{
		{"active", "\033[32m"},
		{"running", "\033[32m"},
		{"completed", "\033[32m"},
		{"failed", "\033[31m"},
		{"error", "\033[31m"},
		{"creating", "\033[33m"},
		{"pending", "\033[33m"},
		{"deleting", "\033[33m"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := output.ColorStatus(tt.status)
			if tt.color == "" {
				if got != tt.status {
					t.Errorf("ColorStatus(%q) = %q, want %q", tt.status, got, tt.status)
				}
			} else {
				want := tt.color + tt.status + "\033[0m"
				if got != want {
					t.Errorf("ColorStatus(%q) = %q, want %q", tt.status, got, want)
				}
			}
		})
	}
}

func TestColorStatus_NoColor(t *testing.T) {
	output.ColorEnabled = false
	got := output.ColorStatus("active")
	if got != "active" {
		t.Errorf("ColorStatus() with color disabled = %q, want %q", got, "active")
	}
}
