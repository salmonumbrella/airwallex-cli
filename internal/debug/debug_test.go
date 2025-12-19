package debug

import (
	"context"
	"testing"
)

func TestWithDebug(t *testing.T) {
	ctx := context.Background()

	// Default should be false
	if IsEnabled(ctx) {
		t.Error("expected debug disabled by default")
	}

	// Enable debug
	ctx = WithDebug(ctx, true)
	if !IsEnabled(ctx) {
		t.Error("expected debug enabled after WithDebug(true)")
	}

	// Disable debug
	ctx = WithDebug(ctx, false)
	if IsEnabled(ctx) {
		t.Error("expected debug disabled after WithDebug(false)")
	}
}
