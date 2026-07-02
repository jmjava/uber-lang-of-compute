package wheel

import (
	"fmt"
	"strings"
	"time"
)

// State tracks the current position on the Compute Wheel.
type State struct {
	CurrentTimeSlice   time.Time
	ActiveContextIndex int
	RotationCount      int
}

// AdvanceResult describes the outcome of advancing the wheel after a slot completes.
type AdvanceResult struct {
	State         State
	AdvancedSlice bool
	Done          bool // true when max rotations reached
}

// ParseInterval parses duration strings including day suffix (e.g. 1d, 24h, 0s).
func ParseInterval(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty interval")
	}
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		d, err := time.ParseDuration(days + "h")
		if err != nil {
			return 0, fmt.Errorf("parse day interval %q: %w", s, err)
		}
		return d * 24, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("parse interval %q: %w", s, err)
	}
	return d, nil
}

// InitialState returns the starting wheel position.
func InitialState(startTimeSlice string, interval time.Duration, now time.Time) (State, error) {
	var start time.Time
	var err error

	if startTimeSlice != "" {
		start, err = time.Parse(time.RFC3339, startTimeSlice)
		if err != nil {
			return State{}, fmt.Errorf("parse startTimeSlice: %w", err)
		}
	} else {
		start = now.UTC()
		if interval > 0 {
			start = start.Truncate(interval)
		}
	}

	return State{
		CurrentTimeSlice:   start.UTC(),
		ActiveContextIndex: 0,
		RotationCount:      0,
	}, nil
}

// AdvanceAfterCompletion moves to the next context on the wheel, or the next time slice
// when all contexts have processed the current slice.
func AdvanceAfterCompletion(state State, contextCount int, interval time.Duration, maxRotations int) AdvanceResult {
	if contextCount <= 0 {
		return AdvanceResult{State: state, Done: true}
	}

	next := state
	next.ActiveContextIndex++

	if next.ActiveContextIndex >= contextCount {
		next.ActiveContextIndex = 0
		next.CurrentTimeSlice = next.CurrentTimeSlice.Add(interval)
		next.RotationCount++

		if maxRotations > 0 && next.RotationCount >= maxRotations {
			return AdvanceResult{State: next, AdvancedSlice: true, Done: true}
		}
		return AdvanceResult{State: next, AdvancedSlice: true}
	}

	return AdvanceResult{State: next, AdvancedSlice: false}
}

// ActiveContextName returns the context name at the current index.
func ActiveContextName(contexts []string, index int) (string, error) {
	if index < 0 || index >= len(contexts) {
		return "", fmt.Errorf("context index %d out of range (count=%d)", index, len(contexts))
	}
	return contexts[index], nil
}

// FormatTimeSlice formats a time slice for use in resource names and status.
func FormatTimeSlice(t time.Time) string {
	return t.UTC().Format("20060102t150405z")
}

// ParseTimeSlice parses a formatted time slice string.
func ParseTimeSlice(s string) (time.Time, error) {
	return time.Parse("20060102t150405z", s)
}

// WorkflowName builds a deterministic Workflow name for a wheel slot.
func WorkflowName(wheelName, contextName, timeSlice string) string {
	base := fmt.Sprintf("%s-%s-%s", wheelName, contextName, timeSlice)
	base = strings.ToLower(base)
	base = strings.ReplaceAll(base, "_", "-")
	if len(base) <= 63 {
		return base
	}
	// Truncate preserving suffix uniqueness via hash of full name
	hash := fmt.Sprintf("%x", len(base))
	return base[:63-len(hash)] + hash
}

// RequeueDelay returns how long to wait before checking an in-flight workflow again.
func RequeueDelay(interval time.Duration) time.Duration {
	if interval <= 0 {
		return time.Second
	}
	if interval < 30*time.Second {
		return interval
	}
	return 10 * time.Second
}
