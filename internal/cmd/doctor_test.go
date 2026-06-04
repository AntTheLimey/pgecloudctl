package cmd

import (
	"strings"
	"testing"
)

func TestDoctorHelp(t *testing.T) {
	out, err := executeCommand("doctor", "--help")
	if err != nil {
		t.Fatalf("doctor help: %v", err)
	}
	if !strings.Contains(out, "doctor") {
		t.Errorf("help missing 'doctor': %q", out)
	}
}

func TestDoctorRuns(t *testing.T) {
	out, err := executeCommand("doctor")
	if err != nil {
		t.Fatalf("doctor: %v", err)
	}
	if !strings.Contains(out, "Version") {
		t.Errorf("output missing 'Version': %q", out)
	}
	if !strings.Contains(out, "Config") {
		t.Errorf("output missing 'Config': %q", out)
	}
}

func TestDoctorJSON(t *testing.T) {
	flagOutput = "json"
	defer func() { flagOutput = "table" }()

	out, err := executeCommand("doctor")
	if err != nil {
		t.Fatalf("doctor json: %v", err)
	}
	if !strings.Contains(out, `"version"`) {
		t.Errorf("JSON output missing 'version' key: %q", out)
	}
	if !strings.Contains(out, `"config"`) {
		t.Errorf("JSON output missing 'config' key: %q", out)
	}
}
