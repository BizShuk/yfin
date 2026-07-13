// rate_limit.go — QPS + burst limits at the per-host and per-session
// layer (Yahoo Finance). Capacity: 1 struct (`RateLimitConfig`).
package types

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	PerHostQPS      float64 `yaml:"per_host_qps"`
	PerHostBurst    int     `yaml:"per_host_burst"`
	PerSessionQPS   float64 `yaml:"per_session_qps"`
	PerSessionBurst int     `yaml:"per_session_burst"`
}
