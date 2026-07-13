// adapters.go — mechanical accessors on `*Config` that hand out
// flat / nested views of the root config tree:
//
//   - GetHTTPConfig      — flatten Yahoo + RateLimit + Retry +
//     CircuitBreaker into the httpx-shaped `HTTPConfig` consumed by
//     `cmd/client.go`.
//   - GetFXConfig        — `&c.FX` (fx sub-tree as-is).
//   - GetScrapeConfig    — `&c.Scrape` (scrape sub-tree as-is).
//   - ValidateInterval   — markets.allowed_intervals membership.
//   - ValidateAdjustmentPolicy — markets.default_adjustment_policy enum.
//
// Capacity: 4 accessor functions + 2 validators.
package types

import (
	"fmt"
	"time"
)

// GetHTTPConfig converts the configuration to httpx.Config
func (c *Config) GetHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		BaseURL:          c.Yahoo.BaseURL,
		Timeout:          time.Duration(c.Yahoo.TimeoutMs) * time.Millisecond,
		IdleTimeout:      time.Duration(c.Yahoo.IdleTimeoutMs) * time.Millisecond,
		MaxConnsPerHost:  c.Yahoo.MaxConnsPerHost,
		UserAgent:        c.Yahoo.UserAgent,
		MaxAttempts:      c.Retry.Attempts,
		BackoffBaseMs:    c.Retry.BaseMs,
		BackoffJitterMs:  c.Retry.BaseMs / 2, // Default jitter
		MaxDelayMs:       c.Retry.MaxDelayMs,
		QPS:              c.RateLimit.PerHostQPS,
		Burst:            c.RateLimit.PerHostBurst,
		CircuitWindow:    time.Duration(c.CircuitBreaker.Window) * time.Second,
		FailureThreshold: c.CircuitBreaker.FailureThreshold,
		ResetTimeout:     time.Duration(c.CircuitBreaker.ResetTimeoutMs) * time.Millisecond,
	}
}

// GetFXConfig returns the FX sub-config tree (verbatim pointer).
func (c *Config) GetFXConfig() *FXConfig {
	return &c.FX
}

// GetScrapeConfig returns the scrape sub-config tree (verbatim pointer).
func (c *Config) GetScrapeConfig() *ScrapeConfig {
	return &c.Scrape
}

// ValidateInterval validates that the interval is allowed
func (c *Config) ValidateInterval(interval string) error {
	for _, allowed := range c.Markets.AllowedIntervals {
		if interval == allowed {
			return nil
		}
	}
	return fmt.Errorf("interval '%s' is not allowed. Allowed intervals: %v", interval, c.Markets.AllowedIntervals)
}

// ValidateAdjustmentPolicy validates that the adjustment policy is allowed
func (c *Config) ValidateAdjustmentPolicy(policy string) error {
	if policy == "raw" || policy == "split_dividend" {
		return nil
	}
	return fmt.Errorf("adjustment policy '%s' is not allowed. Allowed policies: raw, split_dividend", policy)
}
