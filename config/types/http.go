// http.go — flat `HTTPConfig` output that `Config.GetHTTPConfig` produces
// for `utils/httpx.NewClient`. Holds only primitive (non-YAML) fields —
// it is the post-load, fully-typed view of the Yahoo + rate-limit +
// retry + circuit-breaker stack that `cmd/client.go` consumes. No
// factory here; see adapters.go for `(*Config).GetHTTPConfig`.
// Capacity: 1 struct (`HTTPConfig`).
package types

import "time"

// HTTPConfig represents HTTP client configuration (compatible with httpx.Config)
type HTTPConfig struct {
	BaseURL          string
	Timeout          time.Duration
	IdleTimeout      time.Duration
	MaxConnsPerHost  int
	UserAgent        string
	MaxAttempts      int
	BackoffBaseMs    int
	BackoffJitterMs  int
	MaxDelayMs       int
	QPS              float64
	Burst            int
	CircuitWindow    time.Duration
	FailureThreshold float64
	ResetTimeout     time.Duration
}
