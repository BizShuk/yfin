// rate_limit.go — QPS + burst limits at the per-host layer (Yahoo
// Finance). per_session_* was removed: session rotation was dropped
// (see CLAUDE.md), and the per-session knobs were vestigial. Capacity:
// 1 struct (`RateLimitConfig`).
package config

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	PerHostQPS   float64 `yaml:"per_host_qps"`
	PerHostBurst int     `yaml:"per_host_burst"`
}
