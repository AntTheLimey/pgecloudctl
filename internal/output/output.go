package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Row is implemented by any type that can provide its column values as strings.
type Row interface {
	Columns() []string
}

// Print writes data to w in the requested format.
//
// format "json"  — pretty-printed JSON via json.Encoder.
// format "table" — tab-aligned table via text/tabwriter; headers printed first.
// Any other format returns an error.
//
// For table mode, data must be []Row. For JSON mode, data is encoded directly.
func Print(w io.Writer, format string, data interface{}, headers []string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)

	case "table":
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

		if len(headers) > 0 {
			fmt.Fprintln(tw, strings.Join(headers, "\t"))
		}

		if rows, ok := data.([]Row); ok {
			for _, r := range rows {
				fmt.Fprintln(tw, strings.Join(r.Columns(), "\t"))
			}
		}

		return tw.Flush()

	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}
