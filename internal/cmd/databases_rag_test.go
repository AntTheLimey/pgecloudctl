package cmd

import "testing"

func TestParsePipelineConfig(t *testing.T) {
	bareArray := `[
	  {"name": "p1", "tables": [{"table": "t", "text_column": "c", "vector_column": "v"}]}
	]`
	wrapped := `{"pipelines": [
	  {"name": "p1", "tables": [{"table": "t", "text_column": "c", "vector_column": "v"}]}
	]}`

	tests := []struct {
		name      string
		data      string
		wantLen   int
		wantName  string
		wantError bool
	}{
		{name: "bare array", data: bareArray, wantLen: 1, wantName: "p1"},
		{name: "wrapped object", data: wrapped, wantLen: 1, wantName: "p1"},
		{name: "empty array", data: `[]`, wantLen: 0},
		{name: "invalid json", data: `{not json`, wantError: true},
		{name: "wrong shape", data: `"a string"`, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePipelineConfig([]byte(tt.data))
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tt.wantLen)
			}
			if tt.wantName != "" && got[0].Name != tt.wantName {
				t.Errorf("name = %q, want %q", got[0].Name, tt.wantName)
			}
		})
	}
}
