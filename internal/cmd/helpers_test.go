package cmd

import (
	"testing"
)

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantCode int
		wantNil  bool
	}{
		{"200 OK", 200, "", 0, true},
		{"201 Created", 201, "", 0, true},
		{"204 No Content", 204, "", 0, true},
		{"401 Unauthorized", 401, "unauthorized", ExitAuth, false},
		{"403 Forbidden", 403, "forbidden", ExitAuth, false},
		{"404 Not Found", 404, "not found", ExitNotFound, false},
		{"408 Timeout", 408, "timeout", ExitTimeout, false},
		{"500 Internal", 500, "internal error", ExitGeneral, false},
		{"504 Gateway Timeout", 504, "gateway timeout", ExitTimeout, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkResponse(tt.status, tt.body)
			if tt.wantNil {
				if err != nil {
					t.Errorf("checkResponse(%d) = %v, want nil", tt.status, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("checkResponse(%d) = nil, want error", tt.status)
			}
			ee, ok := err.(*ExitError)
			if !ok {
				t.Fatalf("checkResponse(%d) returned %T, want *ExitError", tt.status, err)
			}
			if ee.Code() != tt.wantCode {
				t.Errorf("exit code = %d, want %d", ee.Code(), tt.wantCode)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-03-15T10:30:00Z", "2024-03-15"},
		{"2024-03-15", "2024-03-15"},
		{"short", "short"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := formatTime(tt.input)
			if got != tt.want {
				t.Errorf("formatTime(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{"empty", nil, ""},
		{"single", []string{"us-east-1"}, "us-east-1"},
		{"multiple", []string{"us-east-1", "eu-west-1"}, "us-east-1, eu-west-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinStrings(tt.input)
			if got != tt.want {
				t.Errorf("joinStrings(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
