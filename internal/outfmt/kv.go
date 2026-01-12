package outfmt

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// KV represents a key/value row for text output.
type KV struct {
	Key   string
	Value string
}

// WriteKV writes key/value rows in a tab-aligned format.
func WriteKV(w io.Writer, rows []KV) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for _, row := range rows {
		if row.Key == "" {
			continue
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", row.Key, row.Value)
	}
	return tw.Flush()
}
