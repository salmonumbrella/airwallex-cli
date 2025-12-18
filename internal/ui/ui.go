package ui

import (
	"context"
	"io"
	"os"

	"golang.org/x/term"
)

type contextKey string

const uiKey contextKey = "ui"

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

type UI struct {
	out   io.Writer
	err   io.Writer
	color bool
}

func New(colorMode string) *UI {
	var color bool
	switch colorMode {
	case "never":
		color = false
	case "always":
		color = true
	default: // auto
		color = term.IsTerminal(int(os.Stdout.Fd()))
	}

	if os.Getenv("NO_COLOR") != "" {
		color = false
	}

	return &UI{
		out:   os.Stdout,
		err:   os.Stderr,
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
		msg = colorGreen + msg + colorReset
	}
	_, _ = io.WriteString(u.err, msg+"\n")
}

func (u *UI) Error(msg string) {
	if u.color {
		msg = colorRed + msg + colorReset
	}
	_, _ = io.WriteString(u.err, msg+"\n")
}

func (u *UI) Info(msg string) {
	_, _ = io.WriteString(u.err, msg+"\n")
}
