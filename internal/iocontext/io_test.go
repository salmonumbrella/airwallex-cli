package iocontext

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestIO_DefaultIO(t *testing.T) {
	io := DefaultIO()
	if io == nil {
		t.Fatal("DefaultIO() returned nil")
	}
	if io.Out == nil || io.ErrOut == nil || io.In == nil {
		t.Error("DefaultIO() should have non-nil streams")
	}
}

func TestIO_WithIO(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	customIO := &IO{
		Out:    &outBuf,
		ErrOut: &errBuf,
		In:     strings.NewReader("test input"),
	}

	ctx := WithIO(context.Background(), customIO)
	retrieved := GetIO(ctx)

	if retrieved != customIO {
		t.Error("GetIO() did not return the IO that was set with WithIO()")
	}
}

func TestIO_GetIO_DefaultsWhenNotSet(t *testing.T) {
	ctx := context.Background()
	io := GetIO(ctx)

	if io == nil {
		t.Fatal("GetIO() should never return nil")
	}
	// When IO is not in context, GetIO should return default streams
	// We can't check exact equality with os.Stdout since DefaultIO creates a new struct
	if io.Out == nil || io.ErrOut == nil || io.In == nil {
		t.Error("GetIO() without context should return default streams")
	}
}

func TestIO_HasIO(t *testing.T) {
	// Test with no IO in context
	ctx := context.Background()
	if HasIO(ctx) {
		t.Error("HasIO() should return false when IO is not in context")
	}

	// Test with IO in context
	customIO := &IO{
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
		In:     strings.NewReader(""),
	}
	ctx = WithIO(ctx, customIO)
	if !HasIO(ctx) {
		t.Error("HasIO() should return true when IO is in context")
	}
}
