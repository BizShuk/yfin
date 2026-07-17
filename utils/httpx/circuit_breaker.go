// circuit_breaker.go — identity-local rolling-window circuit breaker state and transitions. Capacity: 2 modes, 3 states, 2 constructors.
package httpx

import (
	"sync"
	"time"
)

type circuitOutcomeKind uint8

const (
	circuitOutcomeNeutral circuitOutcomeKind = iota
	circuitOutcomeSuccess
	circuitOutcomeFailure
)

type circuitOutcome struct {
	at     time.Time
	failed bool
}

// CircuitBreaker implements count-based and failure-rate circuit breaking.
// Outcomes are retained only for the configured rolling time window.
type CircuitBreaker struct {
	window               time.Duration
	failureThreshold     int
	failureRateThreshold float64
	minimumRequests      int
	resetTimeout         time.Duration

	state         CircuitState
	outcomes      []circuitOutcome
	openedAt      time.Time
	probeInFlight bool
	now           func() time.Time
	mu            sync.Mutex
}

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a legacy count-mode circuit breaker.
func NewCircuitBreaker(window time.Duration, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		window:           window,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            StateClosed,
		now:              time.Now,
	}
}

// NewFailureRateCircuitBreaker creates a rolling failure-rate circuit breaker.
func NewFailureRateCircuitBreaker(
	window time.Duration,
	failureRateThreshold float64,
	minimumRequests int,
	resetTimeout time.Duration,
) *CircuitBreaker {
	return &CircuitBreaker{
		window:               window,
		failureRateThreshold: failureRateThreshold,
		minimumRequests:      minimumRequests,
		resetTimeout:         resetTimeout,
		state:                StateClosed,
		now:                  time.Now,
	}
}

// Allow reports whether the breaker permits a request. After resetTimeout an
// open breaker grants exactly one half-open probe.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := cb.now()
	switch cb.state {
	case StateClosed:
		cb.pruneLocked(now)
		return true
	case StateOpen:
		if now.Sub(cb.openedAt) < cb.resetTimeout {
			return false
		}
		cb.state = StateHalfOpen
		cb.probeInFlight = true
		return true
	case StateHalfOpen:
		if cb.probeInFlight {
			return false
		}
		cb.probeInFlight = true
		return true
	default:
		return false
	}
}

// RecordSuccess records one logical request as an availability success.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.record(circuitOutcomeSuccess)
}

// RecordFailure records one logical request as an availability failure.
func (cb *CircuitBreaker) RecordFailure() {
	cb.record(circuitOutcomeFailure)
}

func (cb *CircuitBreaker) record(kind circuitOutcomeKind) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := cb.now()
	cb.pruneLocked(now)

	switch cb.state {
	case StateHalfOpen:
		cb.probeInFlight = false
		if kind == circuitOutcomeSuccess {
			cb.state = StateClosed
			cb.outcomes = nil
			cb.openedAt = time.Time{}
			return
		}
		if kind == circuitOutcomeFailure {
			cb.outcomes = append(cb.outcomes, circuitOutcome{at: now, failed: true})
		}
		cb.state = StateOpen
		cb.openedAt = now
		return
	case StateOpen:
		return
	case StateClosed:
		if kind == circuitOutcomeNeutral {
			return
		}
		cb.outcomes = append(cb.outcomes, circuitOutcome{
			at:     now,
			failed: kind == circuitOutcomeFailure,
		})
		if cb.shouldOpenLocked() {
			cb.state = StateOpen
			cb.openedAt = now
		}
	}
}

func (cb *CircuitBreaker) pruneLocked(now time.Time) {
	if cb.window <= 0 || len(cb.outcomes) == 0 {
		return
	}
	cutoff := now.Add(-cb.window)
	firstActive := 0
	for firstActive < len(cb.outcomes) && cb.outcomes[firstActive].at.Before(cutoff) {
		firstActive++
	}
	if firstActive == 0 {
		return
	}
	copy(cb.outcomes, cb.outcomes[firstActive:])
	cb.outcomes = cb.outcomes[:len(cb.outcomes)-firstActive]
}

func (cb *CircuitBreaker) shouldOpenLocked() bool {
	failures := cb.failuresLocked()
	if cb.failureThreshold > 0 {
		return failures >= cb.failureThreshold
	}
	if cb.minimumRequests <= 0 || len(cb.outcomes) < cb.minimumRequests {
		return false
	}
	return float64(failures)/float64(len(cb.outcomes)) >= cb.failureRateThreshold
}

func (cb *CircuitBreaker) failuresLocked() int {
	failures := 0
	for _, outcome := range cb.outcomes {
		if outcome.failed {
			failures++
		}
	}
	return failures
}

// State returns the current breaker state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Failures returns the active-window failure count.
func (cb *CircuitBreaker) Failures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.pruneLocked(cb.now())
	return cb.failuresLocked()
}

// Samples returns the active-window outcome count.
func (cb *CircuitBreaker) Samples() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.pruneLocked(cb.now())
	return len(cb.outcomes)
}
