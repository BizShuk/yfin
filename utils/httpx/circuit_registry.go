// circuit_registry.go — lazy per-authority and logical-group ownership of independent circuit breakers. Capacity: 1 registry, 1 config normalizer.
package httpx

import (
	"strings"
	"sync"
)

type circuitBreakerRegistry struct {
	mu       sync.Mutex
	breakers map[circuitBreakerKey]*CircuitBreaker
	new      func() *CircuitBreaker
}

type circuitBreakerKey struct {
	authority string
	group     string
}

func newCircuitBreakerRegistry(config *Config) *circuitBreakerRegistry {
	normalizeCircuitConfig(config)
	window := config.CircuitWindow
	failureThreshold := config.FailureThreshold
	failureRateThreshold := config.FailureRateThreshold
	minimumRequests := config.MinimumRequests
	resetTimeout := config.ResetTimeout

	var factory func() *CircuitBreaker
	if failureThreshold > 0 {
		factory = func() *CircuitBreaker {
			return NewCircuitBreaker(
				window,
				failureThreshold,
				resetTimeout,
			)
		}
	} else {
		factory = func() *CircuitBreaker {
			return NewFailureRateCircuitBreaker(
				window,
				failureRateThreshold,
				minimumRequests,
				resetTimeout,
			)
		}
	}

	return &circuitBreakerRegistry{
		breakers: make(map[circuitBreakerKey]*CircuitBreaker),
		new:      factory,
	}
}

func normalizeCircuitConfig(config *Config) {
	defaults := DefaultConfig()
	if config.CircuitWindow <= 0 {
		config.CircuitWindow = defaults.CircuitWindow
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = defaults.ResetTimeout
	}
	if config.FailureThreshold <= 0 && config.FailureRateThreshold <= 0 {
		config.FailureRateThreshold = defaults.FailureRateThreshold
	}
	if config.FailureThreshold <= 0 && config.MinimumRequests <= 0 {
		config.MinimumRequests = defaults.MinimumRequests
	}
}

func (r *circuitBreakerRegistry) forRequest(authority, group string) *CircuitBreaker {
	key := circuitBreakerKey{
		authority: strings.ToLower(strings.TrimSpace(authority)),
		group:     normalizeCircuitGroup(group),
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	breaker, ok := r.breakers[key]
	if !ok {
		breaker = r.new()
		r.breakers[key] = breaker
	}
	return breaker
}

func (r *circuitBreakerRegistry) forHost(host string) *CircuitBreaker {
	return r.forRequest(host, "")
}
