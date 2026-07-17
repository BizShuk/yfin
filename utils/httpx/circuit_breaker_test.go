// circuit_breaker_test.go — Tests `CircuitBreaker` count/rate thresholds, rolling-window expiry, and single-probe state transitions. Capacity: 9 test functions.
package httpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setCircuitClock(cb *CircuitBreaker, current *time.Time) {
	cb.now = func() time.Time { return *current }
}

func TestCircuitBreaker(t *testing.T) {
	// Create circuit breaker with small window for testing
	cb := NewCircuitBreaker(5*time.Second, 3, 100*time.Millisecond)

	// Initially should be closed
	if cb.State() != StateClosed {
		t.Errorf("Expected initial state to be closed, got %v", cb.State())
	}

	// Should allow requests initially
	if !cb.Allow() {
		t.Error("Expected circuit breaker to allow requests initially")
	}

	// Record some failures to trigger opening
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	// Should be open after 3 failures (threshold is 3)
	if cb.State() != StateOpen {
		t.Errorf("Expected state to be open after 3 failures, got %v", cb.State())
	}

	// Should now be open
	if cb.State() != StateOpen {
		t.Errorf("Expected state to be open, got %v", cb.State())
	}

	// Should not allow requests when open
	if cb.Allow() {
		t.Error("Expected circuit breaker to not allow requests when open")
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(5*time.Second, 3, 50*time.Millisecond)

	// Open the circuit
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if cb.State() != StateOpen {
		t.Errorf("Expected state to be open, got %v", cb.State())
	}

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Should transition to half-open (but Allow() needs to be called to trigger transition)
	cb.Allow() // This triggers the transition
	if cb.State() != StateHalfOpen {
		t.Errorf("Expected state to be half-open, got %v", cb.State())
	}

	// Only the transition call receives the half-open probe.
	if cb.Allow() {
		t.Error("Expected circuit breaker to reject a second half-open probe")
	}

	// Record success
	cb.RecordSuccess()

	// Should now be closed after one success
	if cb.State() != StateClosed {
		t.Errorf("Expected state to be closed after success, got %v", cb.State())
	}
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(5*time.Second, 3, 50*time.Millisecond)

	// Open the circuit
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// Wait for reset timeout
	time.Sleep(100 * time.Millisecond)

	// Should be half-open (but Allow() needs to be called to trigger transition)
	cb.Allow() // This triggers the transition
	if cb.State() != StateHalfOpen {
		t.Errorf("Expected state to be half-open, got %v", cb.State())
	}

	// Record a failure in half-open state
	cb.RecordFailure()

	// Should immediately go back to open
	if cb.State() != StateOpen {
		t.Errorf("Expected state to be open after failure in half-open, got %v", cb.State())
	}
}

func TestCircuitBreakerRollingWindow(t *testing.T) {
	cb := NewCircuitBreaker(3*time.Second, 2, 100*time.Millisecond)

	// Record 1 failure (should not open yet)
	cb.RecordFailure()

	if cb.State() != StateClosed {
		t.Errorf("Expected state to be closed with 1 failure, got %v", cb.State())
	}

	// Record another failure (should open: threshold is 2)
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("Expected state to be open after 2 failures, got %v", cb.State())
	}
}

func TestCircuitBreakerRateModeRequiresMinimumSamples(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewFailureRateCircuitBreaker(time.Minute, 0.30, 10, 30*time.Second)
	setCircuitClock(cb, &now)

	for range 9 {
		cb.RecordFailure()
	}
	assert.Equal(t, StateClosed, cb.State())

	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreakerRateModeUsesSuccessDenominator(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewFailureRateCircuitBreaker(time.Minute, 0.30, 10, 30*time.Second)
	setCircuitClock(cb, &now)

	for range 7 {
		cb.RecordSuccess()
	}
	for range 2 {
		cb.RecordFailure()
	}
	assert.Equal(t, StateClosed, cb.State())

	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreakerExpiresOutcomesOutsideWindow(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewFailureRateCircuitBreaker(time.Minute, 0.50, 3, 30*time.Second)
	setCircuitClock(cb, &now)

	cb.RecordFailure()
	now = now.Add(time.Minute + time.Second)
	cb.RecordSuccess()
	cb.RecordSuccess()

	assert.Equal(t, 2, cb.Samples())
	assert.Equal(t, 0, cb.Failures())
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreakerHalfOpenAllowsSingleProbe(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewCircuitBreaker(time.Minute, 1, 30*time.Second)
	setCircuitClock(cb, &now)

	cb.RecordFailure()
	now = now.Add(31 * time.Second)
	assert.True(t, cb.Allow())
	assert.False(t, cb.Allow())

	cb.RecordSuccess()
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreakerNeutralProbeReopens(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cb := NewCircuitBreaker(time.Minute, 1, 30*time.Second)
	setCircuitClock(cb, &now)

	cb.RecordFailure()
	now = now.Add(31 * time.Second)
	require.True(t, cb.Allow())
	cb.record(circuitOutcomeNeutral)

	assert.Equal(t, StateOpen, cb.State())
	assert.False(t, cb.Allow())
}
