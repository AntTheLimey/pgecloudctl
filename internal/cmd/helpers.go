package cmd

import "strings"

// truncateID returns the first 8 characters of id, or the full string if
// shorter. Useful for displaying UUIDs in table output.
func truncateID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// joinStrings joins a slice of strings with ", " as the separator.
func joinStrings(ss []string) string {
	return strings.Join(ss, ", ")
}

// formatTime returns just the date portion (first 10 chars) of an ISO-8601
// timestamp string, e.g. "2024-03-15" from "2024-03-15T10:30:00Z".
func formatTime(t string) string {
	if len(t) <= 10 {
		return t
	}
	return t[:10]
}
