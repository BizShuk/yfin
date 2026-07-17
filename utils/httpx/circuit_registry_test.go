// circuit_registry_test.go — Tests authority reuse, host isolation, and construction-time config snapshots. Capacity: 3 tests.
package httpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreakerRegistryReusesHostBreaker(t *testing.T) {
	registry := newCircuitBreakerRegistry(&Config{
		CircuitWindow:    time.Minute,
		FailureThreshold: 1,
		ResetTimeout:     time.Minute,
	})

	assert.Same(
		t,
		registry.forHost("query1.finance.yahoo.com"),
		registry.forHost("QUERY1.finance.yahoo.com"),
	)
}

func TestCircuitBreakerRegistryIsolatesHosts(t *testing.T) {
	registry := newCircuitBreakerRegistry(&Config{
		CircuitWindow:    time.Minute,
		FailureThreshold: 1,
		ResetTimeout:     time.Minute,
	})
	a := registry.forHost("query1.finance.yahoo.com")
	b := registry.forHost("finance.yahoo.com")

	require.NotSame(t, a, b)
	a.RecordFailure()
	assert.Equal(t, StateOpen, a.State())
	assert.Equal(t, StateClosed, b.State())
}

func TestCircuitBreakerRegistrySnapshotsConfig(t *testing.T) {
	config := &Config{
		CircuitWindow:    time.Minute,
		FailureThreshold: 1,
		ResetTimeout:     time.Minute,
	}
	registry := newCircuitBreakerRegistry(config)
	config.FailureThreshold = 10

	breaker := registry.forHost("finance.yahoo.com")
	breaker.RecordFailure()
	assert.Equal(t, StateOpen, breaker.State())
}

func TestCircuitBreakerRegistryReusesNormalizedGroup(t *testing.T) {
	registry := newCircuitBreakerRegistry(&Config{
		CircuitWindow:    time.Minute,
		FailureThreshold: 1,
		ResetTimeout:     time.Minute,
	})
	assert.Same(t,
		registry.forRequest("QUERY2.finance.yahoo.com", " Yahoo-Auth "),
		registry.forRequest("query2.finance.yahoo.com", "yahoo-auth"),
	)
}

func TestCircuitBreakerRegistryIsolatesGroupsOnSameHost(t *testing.T) {
	registry := newCircuitBreakerRegistry(&Config{
		CircuitWindow:    time.Minute,
		FailureThreshold: 1,
		ResetTimeout:     time.Minute,
	})
	auth := registry.forRequest("query2.finance.yahoo.com", "yahoo-auth")
	chart := registry.forRequest("query2.finance.yahoo.com", "yahoo-chart")
	ungrouped := registry.forHost("query2.finance.yahoo.com")

	require.NotSame(t, auth, chart)
	require.NotSame(t, auth, ungrouped)
	assert.Same(t, ungrouped, registry.forRequest("query2.finance.yahoo.com", ""))
	auth.RecordFailure()
	assert.Equal(t, StateOpen, auth.State())
	assert.Equal(t, StateClosed, chart.State())
	assert.Equal(t, StateClosed, ungrouped.State())
}
