package ui

import (
	"context"
	"io"
	"os"

	"github.com/muesli/termenv"
)

type contextKey string

const uiKey contextKey = "ui"

type UI struct {
	out   *termenv.Output
	err   *termenv.Output
	color bool
}

func New(colorMode string) *UI {
	out := termenv.NewOutput(os.Stdout)
	errOut := termenv.NewOutput(os.Stderr)

	var color bool
	switch colorMode {
	case "never":
		color = false
	case "always":
		color = true
	default: // auto
		color = out.ColorProfile() != termenv.Ascii
	}

	if os.Getenv("NO_COLOR") != "" {
		color = false
	}

	return &UI{
		out:   out,
		err:   errOut,
		color: color,
	}
}

func WithUI(ctx context.Context, u *UI) context.Context {
	return context.WithValue(ctx, uiKey, u)
}

func FromContext(ctx context.Context) *UI {
	if u, ok := ctx.Value(uiKey).(*UI); ok {
		return u
	}
	return New("auto")
}

func (u *UI) Out() io.Writer {
	return u.out
}

func (u *UI) Err() io.Writer {
	return u.err
}

func (u *UI) Success(msg string) {
	if u.color {
		msg = termenv.String(msg).Foreground(termenv.ANSIGreen).String()
	}
	_, _ = u.err.WriteString(msg + "\n")
}

func (u *UI) Error(msg string) {
	if u.color {
		msg = termenv.String(msg).Foreground(termenv.ANSIRed).String()
	}
	_, _ = u.err.WriteString(msg + "\n")
}

func (u *UI) Info(msg string) {
	_, _ = u.err.WriteString(msg + "\n")
}
