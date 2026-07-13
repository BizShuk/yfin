// retry.go — shared retry + circuit-breaker config used by both the
// Yahoo HTTP client (driven from `Config.GetHTTPConfig`) and the bus
// publisher (driven from `Config.GetBusConfig`). Capacity: 2 structs
// (`RetryConfig`, `CircuitBreakerConfig`).
package config

// RetryConfig represents retry configuration
type RetryConfig struct {
	Attempts   int `yaml:"attempts"`
	BaseMs     int `yaml:"base_ms"`
	MaxDelayMs int `yaml:"max_delay_ms"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	Window           int     `yaml:"window"`
	FailureThreshold float64 `yaml:"failure_threshold"`
	ResetTimeoutMs   int     `yaml:"reset_timeout_ms"`
	HalfOpenProbes   int     `yaml:"half_open_probes"`
}
