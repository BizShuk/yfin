// defaults.go — shared HTTP policy presets.
package httpx

import "time"

// DefaultScrapeConfig returns the conservative HTTP policy used for Yahoo
// HTML pages. Config loading and the scrape service both consume this single
// owner so retry, rate-limit, timeout, and body-cap defaults cannot drift.
func DefaultScrapeConfig() *Config {
	return &Config{
		BaseURL:              "https://finance.yahoo.com",
		Timeout:              10 * time.Second,
		IdleTimeout:          90 * time.Second,
		MaxConnsPerHost:      10,
		MaxAttempts:          4,
		BackoffBaseMs:        300,
		BackoffJitterMs:      150,
		MaxDelayMs:           4000,
		QPS:                  0.7,
		Burst:                1,
		CircuitWindow:        60 * time.Second,
		FailureThreshold:     0,
		FailureRateThreshold: 0.30,
		MinimumRequests:      10,
		ResetTimeout:         30 * time.Second,
		UserAgent:            "Mozilla/5.0 (compatible; yfin/1.x)",
		MaxBodyBytes:         8 << 20,
	}
}
