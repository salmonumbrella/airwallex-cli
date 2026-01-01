// Package wait provides a unified polling pattern for waiting on resource state changes.
package wait

import (
	"context"
	"fmt"
	"slices"
	"time"
)

// Config configures wait behavior.
type Config struct {
	Timeout       time.Duration // Max time to wait
	PollInterval  time.Duration // Time between polls
	SuccessStates []string      // Terminal success states
	FailureStates []string      // Terminal failure states
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Timeout:      5 * time.Minute,
		PollInterval: 2 * time.Second,
	}
}

// StateError indicates the resource reached a failure state.
type StateError struct {
	State string
}

func (e *StateError) Error() string {
	return fmt.Sprintf("reached failure state: %s", e.State)
}

// For polls until a terminal state is reached or timeout.
// pollFn should return the current state.
func For(ctx context.Context, cfg Config, pollFn func() (string, error)) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	// Poll immediately first
	state, err := pollFn()
	if err != nil {
		return "", err
	}
	if isTerminal(state, cfg) {
		if slices.Contains(cfg.FailureStates, state) {
			return state, &StateError{State: state}
		}
		return state, nil
	}

	for {
		select {
		case <-ctx.Done():
			return state, ctx.Err()
		case <-ticker.C:
			state, err = pollFn()
			if err != nil {
				return "", err
			}
			if isTerminal(state, cfg) {
				if slices.Contains(cfg.FailureStates, state) {
					return state, &StateError{State: state}
				}
				return state, nil
			}
		}
	}
}

func isTerminal(state string, cfg Config) bool {
	return slices.Contains(cfg.SuccessStates, state) ||
		slices.Contains(cfg.FailureStates, state)
}
