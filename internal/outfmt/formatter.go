package outfmt

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// Formatter handles output formatting for commands.
type Formatter struct {
	ctx       context.Context
	out       io.Writer
	errOut    io.Writer
	tabWriter *tabwriter.Writer
}

// OutputOption configures a Formatter.
type OutputOption func(*Formatter)

// WithWriter sets the output writer (default: os.Stdout).
func WithWriter(w io.Writer) OutputOption {
	return func(f *Formatter) {
		f.out = w
		f.tabWriter = tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	}
}

// WithErrWriter sets the error writer (default: os.Stderr).
func WithErrWriter(w io.Writer) OutputOption {
	return func(f *Formatter) {
		f.errOut = w
	}
}

// FromContext creates a Formatter from context with optional configuration.
func FromContext(ctx context.Context, opts ...OutputOption) *Formatter {
	f := &Formatter{
		ctx:    ctx,
		out:    os.Stdout,
		errOut: os.Stderr,
	}
	f.tabWriter = tabwriter.NewWriter(f.out, 0, 4, 2, ' ', 0)

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Output writes data as JSON or text based on context format.
// For JSON mode, applies JQ filtering if a query is present.
func (f *Formatter) Output(data any) error {
	if IsJSON(f.ctx) {
		return WriteJSONFiltered(f.out, data, GetQuery(f.ctx))
	}
	return nil
}

// StartTable writes table headers and returns true if in text mode.
// Returns false if in JSON mode (caller should skip row writing).
func (f *Formatter) StartTable(headers []string) bool {
	if IsJSON(f.ctx) {
		return false
	}

	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(f.tabWriter, "\t")
		}
		fmt.Fprint(f.tabWriter, h)
	}
	fmt.Fprintln(f.tabWriter)
	return true
}

// Row writes a single row to the table.
func (f *Formatter) Row(columns ...string) {
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(f.tabWriter, "\t")
		}
		fmt.Fprint(f.tabWriter, col)
	}
	fmt.Fprintln(f.tabWriter)
}

// EndTable flushes the table output.
func (f *Formatter) EndTable() error {
	return f.tabWriter.Flush()
}

// Empty writes a message to stderr indicating no results were found.
func (f *Formatter) Empty(message string) {
	fmt.Fprintln(f.errOut, message)
}
