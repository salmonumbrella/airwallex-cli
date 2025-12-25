package cmd

import (
	"context"
	"io"
	"os"
)

// IO holds the input/output streams for commands.
type IO struct {
	Out    io.Writer // stdout
	ErrOut io.Writer // stderr
	In     io.Reader // stdin (for future use)
}

// DefaultIO returns the standard IO streams.
func DefaultIO() *IO {
	return &IO{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		In:     os.Stdin,
	}
}

// Context key for IO
type ioKey struct{}

// WithIO adds IO streams to a context.
func WithIO(ctx context.Context, io *IO) context.Context {
	return context.WithValue(ctx, ioKey{}, io)
}

// GetIO retrieves IO streams from context, defaulting to standard streams.
func GetIO(ctx context.Context) *IO {
	if io, ok := ctx.Value(ioKey{}).(*IO); ok && io != nil {
		return io
	}
	return DefaultIO()
}
