package wait

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWait_SuccessOnFirstPoll(t *testing.T) {
	calls := 0
	cfg := Config{
		Timeout:       5 * time.Second,
		PollInterval:  100 * time.Millisecond,
		SuccessStates: []string{"COMPLETED"},
		FailureStates: []string{"FAILED"},
	}

	result, err := For(context.Background(), cfg, func() (string, error) {
		calls++
		return "COMPLETED", nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "COMPLETED" {
		t.Errorf("got %q, want COMPLETED", result)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWait_SuccessAfterPolling(t *testing.T) {
	calls := 0
	cfg := Config{
		Timeout:       5 * time.Second,
		PollInterval:  50 * time.Millisecond,
		SuccessStates: []string{"COMPLETED"},
		FailureStates: []string{"FAILED"},
	}

	result, err := For(context.Background(), cfg, func() (string, error) {
		calls++
		if calls < 3 {
			return "PENDING", nil
		}
		return "COMPLETED", nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "COMPLETED" {
		t.Errorf("got %q, want COMPLETED", result)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWait_FailureState(t *testing.T) {
	cfg := Config{
		Timeout:       5 * time.Second,
		PollInterval:  100 * time.Millisecond,
		SuccessStates: []string{"COMPLETED"},
		FailureStates: []string{"FAILED"},
	}

	result, err := For(context.Background(), cfg, func() (string, error) {
		return "FAILED", nil
	})

	if err == nil {
		t.Error("expected error for failure state")
	}
	if result != "FAILED" {
		t.Errorf("got %q, want FAILED", result)
	}

	var stateErr *StateError
	if !errors.As(err, &stateErr) {
		t.Errorf("expected StateError, got %T", err)
	}
}

func TestWait_Timeout(t *testing.T) {
	cfg := Config{
		Timeout:       100 * time.Millisecond,
		PollInterval:  30 * time.Millisecond,
		SuccessStates: []string{"COMPLETED"},
	}

	_, err := For(context.Background(), cfg, func() (string, error) {
		return "PENDING", nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestWait_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := Config{
		Timeout:       5 * time.Second,
		PollInterval:  50 * time.Millisecond,
		SuccessStates: []string{"COMPLETED"},
	}

	go func() {
		time.Sleep(80 * time.Millisecond)
		cancel()
	}()

	_, err := For(ctx, cfg, func() (string, error) {
		return "PENDING", nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected Canceled, got %v", err)
	}
}

func TestWait_PollError(t *testing.T) {
	cfg := Config{
		Timeout:       5 * time.Second,
		PollInterval:  100 * time.Millisecond,
		SuccessStates: []string{"COMPLETED"},
	}

	expectedErr := errors.New("poll failed")
	_, err := For(context.Background(), cfg, func() (string, error) {
		return "", expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected poll error, got %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Timeout != 5*time.Minute {
		t.Errorf("expected 5m timeout, got %v", cfg.Timeout)
	}
	if cfg.PollInterval != 2*time.Second {
		t.Errorf("expected 2s poll interval, got %v", cfg.PollInterval)
	}
}

func TestStateError(t *testing.T) {
	err := &StateError{State: "FAILED"}
	expected := "reached failure state: FAILED"
	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}
}
