package outfmt

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
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
// Priority for writers:
//  1. If WithWriter/WithErrWriter options provided → use them
//  2. Else if GetIO(ctx) returns non-nil → use it
//  3. Else → fall back to os.Stdout/os.Stderr
func FromContext(ctx context.Context, opts ...OutputOption) *Formatter {
	// Start with context IO if available, otherwise use defaults
	io := getIO(ctx)
	f := &Formatter{
		ctx:    ctx,
		out:    io.Out,
		errOut: io.ErrOut,
	}
	f.tabWriter = tabwriter.NewWriter(f.out, 0, 4, 2, ' ', 0)

	// Options override context IO
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

// OutputList outputs a slice of items as either JSON or a text table.
// In JSON mode, items are output directly. In text mode, headers define
// columns and rowFn extracts column values from each item.
func (f *Formatter) OutputList(items any, headers []string, rowFn func(item any) []string) error {
	// If JSON mode, output items directly
	if IsJSON(f.ctx) {
		return f.Output(items)
	}

	// Text mode: use table methods
	val := reflect.ValueOf(items)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return fmt.Errorf("items must be a slice or array, got %T", items)
	}

	if !f.StartTable(headers) {
		return nil
	}

	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		f.Row(rowFn(item)...)
	}

	return f.EndTable()
}

// getIO is a local helper to avoid import cycle with cmd package.
// It extracts the IO from context using the same key pattern as cmd.GetIO.
type ioKey struct{}

// ioStreams is an interface matching the cmd.IO struct to avoid import cycles.
type ioStreams interface {
	// Out, ErrOut, In are expected to be public fields on the struct
	// We use reflection-like approach via type assertion
}

func getIO(ctx context.Context) struct {
	Out    io.Writer
	ErrOut io.Writer
} {
	// Use the same key type as cmd package to access IO
	if val := ctx.Value(ioKey{}); val != nil {
		// Try to extract IO using reflection to avoid import cycle
		// The cmd.IO type has Out, ErrOut, In fields
		// We use a type switch to handle the struct
		type cmdIO struct {
			Out    io.Writer
			ErrOut io.Writer
			In     io.Reader
		}
		if streams, ok := val.(*cmdIO); ok {
			return struct {
				Out    io.Writer
				ErrOut io.Writer
			}{Out: streams.Out, ErrOut: streams.ErrOut}
		}
		// Also try to handle the actual cmd.IO struct through its fields
		// Use reflection to access fields if needed
		v := reflect.ValueOf(val)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() == reflect.Struct {
			outField := v.FieldByName("Out")
			errOutField := v.FieldByName("ErrOut")
			if outField.IsValid() && errOutField.IsValid() {
				if out, ok := outField.Interface().(io.Writer); ok {
					if errOut, ok := errOutField.Interface().(io.Writer); ok {
						return struct {
							Out    io.Writer
							ErrOut io.Writer
						}{Out: out, ErrOut: errOut}
					}
				}
			}
		}
	}
	return struct {
		Out    io.Writer
		ErrOut io.Writer
	}{Out: os.Stdout, ErrOut: os.Stderr}
}
