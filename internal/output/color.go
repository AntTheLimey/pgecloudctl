package output

import "os"

// ColorEnabled controls whether ANSI color codes are emitted.
// Set to true by root command when stdout is a terminal and NO_COLOR is unset.
var ColorEnabled bool

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		ColorEnabled = false
	}
}

// Bold wraps s in ANSI bold escape codes when color is enabled.
func Bold(s string) string {
	if !ColorEnabled {
		return s
	}
	return bold + s + reset
}

// ColorStatus wraps s in an ANSI color appropriate for its status value.
// Green for success states, red for failure states, yellow for transitional
// states. Returns s unchanged for unrecognised statuses or when color is off.
func ColorStatus(s string) string {
	if !ColorEnabled {
		return s
	}
	switch s {
	case "active", "running", "completed", "succeeded", "available":
		return green + s + reset
	case "failed", "error":
		return red + s + reset
	case "creating", "pending", "deleting", "queued":
		return yellow + s + reset
	default:
		return s
	}
}
