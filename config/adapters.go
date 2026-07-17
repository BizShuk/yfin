// adapters.go — mechanical accessors on `*Config` that hand out
// flat / nested views of the root config tree:
//
//   - assembleHTTPConfig        — flatten Yahoo + RateLimit + Retry +
//     CircuitBreaker into `*httpx.Config`, the canonical HTTP-layer
//     config used by `cmd/client.go`.
//   - assembleScrapeHTTPConfig  — populate `c.Scrape.HTTP` from
//     scrape-tuned Go defaults. The scrape HTTP knobs are no longer
//     in yaml (they live in `svc/scrape.DefaultConfig().HTTP`); this
//     adapter populates a parallel default. To customise, callers
//     mutate `cfg.Scrape.HTTP` post-load or pass a tuned `*httpx.Config`
//     directly to `facade.NewScrapeClientWithHTTP`.
//   - GetHTTPConfig      — returns c.HTTP (assembled at Load() time).
//   - GetFXConfig        — `&c.FX` (fx sub-tree as-is).
//   - GetScrapeConfig    — `&c.Scrape` (scrape sub-tree as-is).
//   - ValidateInterval   — markets.allowed_intervals membership.
//   - ValidateAdjustmentPolicy — markets.default_adjustment_policy enum.
//
// Capacity: 4 accessor functions + 2 validators + 2 assemblers.
package config

import (
	"fmt"
	"time"

	"github.com/bizshuk/yfin/utils/httpx"
)

// assembleHTTPConfig flattens the yaml-side sub-sections (Yahoo +
// RateLimit + Retry + CircuitBreaker) into a single *httpx.Config.
// Called once by Loader.Load() so callers can read everything via
// `c.HTTP.Timeout`, `c.HTTP.QPS`, etc. without per-field copy.
func (c *Config) assembleHTTPConfig() *httpx.Config {
	jitter := 0
	if c.Retry.BaseMs > 0 {
		jitter = c.Retry.BaseMs / 2
	}
	return &httpx.Config{
		BaseURL:              c.Yahoo.BaseURL,
		Timeout:              time.Duration(c.Yahoo.TimeoutMs) * time.Millisecond,
		IdleTimeout:          time.Duration(c.Yahoo.IdleTimeoutMs) * time.Millisecond,
		MaxConnsPerHost:      c.Yahoo.MaxConnsPerHost,
		UserAgent:            c.Yahoo.UserAgent,
		MaxAttempts:          c.Retry.Attempts,
		BackoffBaseMs:        c.Retry.BaseMs,
		BackoffJitterMs:      jitter,
		MaxDelayMs:           c.Retry.MaxDelayMs,
		QPS:                  c.RateLimit.PerHostQPS,
		Burst:                c.RateLimit.PerHostBurst,
		CircuitWindow:        time.Duration(c.CircuitBreaker.Window) * time.Second,
		FailureThreshold:     0,
		FailureRateThreshold: c.CircuitBreaker.FailureThreshold,
		MinimumRequests:      c.CircuitBreaker.MinimumRequests,
		ResetTimeout:         time.Duration(c.CircuitBreaker.ResetTimeoutMs) * time.Millisecond,
		// MaxBodyBytes defaults to 0 (unlimited) — matches the previous
		// cmd/client.go behaviour before the facade indirection.
	}
}

// scrapeHTTPDefaults mirrors svc/scrape.DefaultConfig().HTTP. Kept
// here (duplicated) because the `config` package is a leaf and cannot
// import `svc/scrape`. If you change these values, update both files.
// IMPORTANT: this is the single place where MaxBodyBytes=8 MiB is set
// for scrape traffic — preserves the hardcoded default that used to
// live in `svc/scrape/client.go:NewClient`.
func scrapeHTTPDefaults() *httpx.Config {
	return &httpx.Config{
		BaseURL:              "https://finance.yahoo.com",
		Timeout:              10 * time.Second,
		IdleTimeout:          90 * time.Second,
		MaxConnsPerHost:      10,
		UserAgent:            "Mozilla/5.0 (Ampy yfinance-go scraper)",
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
		MaxBodyBytes:         8 << 20, // 8 MiB — scrape body cap.
	}
}

// assembleScrapeHTTPConfig populates `c.Scrape.HTTP` from the
// scrape-tuned Go defaults. Scrape HTTP knobs (timeout / qps / retry /
// user_agent) are no longer in yaml — defaults come from
// `svc/scrape.DefaultConfig().HTTP` (mirrored here as scrapeHTTPDefaults).
// To customise, callers mutate `cfg.Scrape.HTTP` post-load or pass a
// tuned `*httpx.Config` directly via `facade.NewScrapeClientWithHTTP`.
func (c *Config) assembleScrapeHTTPConfig() *httpx.Config {
	return scrapeHTTPDefaults()
}

// GetHTTPConfig returns the post-load assembled HTTP config. Returns
// nil if Load() was not called or failed.
func (c *Config) GetHTTPConfig() *HTTPConfig {
	return c.HTTP
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

// ValidateAdjustmentPolicy validates that the adjustment policy is valid
func (c *Config) ValidateAdjustmentPolicy(policy string) error {
	if policy == "raw" || policy == "split_dividend" {
		return nil
	}
	return fmt.Errorf("adjustment policy '%s' is not valid. Allowed policies: 'raw', 'split_dividend'", policy)
}
