// yahoo.go — Yahoo Finance HTTP endpoint settings (base URL, timeouts,
// connection pool, user agent). See concurrency.go / rate_limit.go /
// sessions.go / retry.go for the rate-limit + retry tuning that drives
// the same client. Capacity: 1 struct (`YahooConfig`).
package types

// YahooConfig represents Yahoo Finance API configuration
type YahooConfig struct {
	BaseURL         string `yaml:"base_url"`
	TimeoutMs       int    `yaml:"timeout_ms"`
	IdleTimeoutMs   int    `yaml:"idle_timeout_ms"`
	MaxConnsPerHost int    `yaml:"max_conns_per_host"`
	UserAgent       string `yaml:"user_agent"`
}
