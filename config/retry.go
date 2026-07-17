// retry.go — shared retry + circuit-breaker config used by the
// Yahoo HTTP client (driven from `Config.HTTP`, the assembled
// `*httpx.Config` post-load). Capacity: 2 structs (`RetryConfig`,
// `CircuitBreakerConfig`).
package config

const defaultCircuitMinimumRequests = 10

// RetryConfig represents retry configuration
type RetryConfig struct {
	Attempts   int `yaml:"attempts"`
	BaseMs     int `yaml:"base_ms"`
	MaxDelayMs int `yaml:"max_delay_ms"`
}

// CircuitBreakerConfig represents circuit breaker configuration.
// `HalfOpenProbes` was removed: utils/httpx uses a single-probe
// half-open transition and has no field for N probes. To customise
// the half-open behaviour, supply a tuned *httpx.Config via Go code
// (facade.NewClientWithConfig).
type CircuitBreakerConfig struct {
	Window           int     `yaml:"window"`
	FailureThreshold float64 `yaml:"failure_threshold"`
	MinimumRequests  int     `yaml:"minimum_requests"`
	ResetTimeoutMs   int     `yaml:"reset_timeout_ms"`
}
