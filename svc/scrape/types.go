// types.go — scrape-package configuration (`Config`, `EndpointConfig`,
// `RobotsPolicy`, `RobotsEnforce/Warn/Ignore`, `IsValidRobotsPolicy`,
// `DefaultConfig`). All scrape data structs live in model/
// (model.FetchMeta, model.ScrapeNewsItem, model.NewsStats,
// model.RobotsRule, etc.); this file only retains scrape-package
// configuration types and constructors.
//
// `Config` holds scrape-specific knobs (RobotsPolicy, CacheTTLMs,
// Endpoints, Enabled) plus `HTTP *httpx.Config` — the canonical
// HTTP-layer config. The flat HTTP fields that used to live here
// (UserAgent, TimeoutMs, QPS, Burst, Retry{}) have moved out: the
// config loader assembles them into `*httpx.Config` and callers read
// `cfg.HTTP.QPS`, `cfg.HTTP.MaxAttempts`, etc. directly. Capacity: 1
// Config + 1 EndpointConfig + 3 RobotsPolicy constants + 2 helpers.
package scrape

import (
	"time"

	"github.com/bizshuk/yfin/utils/httpx"
)

// Config represents the scraping configuration.
type Config struct {
	// HTTP is the canonical HTTP-layer config (Timeout, QPS, Burst,
	// Retry, UserAgent, MaxBodyBytes, etc.). Caller (facade.NewScrape-
	// ClientFromConfig or tests) populates this from the assembled
	// `config.Scrape.HTTP` rather than re-mapping scrape-side fields
	// by hand.
	HTTP *httpx.Config

	// Scrape-only knobs that don't belong on the HTTP layer.
	Enabled      bool
	RobotsPolicy string
	CacheTTLMs   int
	Endpoints    EndpointConfig
}

// EndpointConfig represents endpoint-specific configuration
type EndpointConfig struct {
	KeyStatistics bool
	Financials    bool
	Analysis      bool
	Profile       bool
	News          bool
}

// DefaultConfig returns a sensible default configuration with scrape-
// tuned HTTP defaults: 0.7 QPS, 4 retries with 300–4000 ms backoff,
// 8 MiB body cap, robots.txt enforced. The HTTP defaults match the
// pre-consolidation hardcoded values in `NewClient` so existing
// callers see identical behavior when they pass nil to NewScrapeClient.
func DefaultConfig() *Config {
	return &Config{
		HTTP: &httpx.Config{
			BaseURL:          "https://finance.yahoo.com",
			Timeout:          10 * time.Second,
			IdleTimeout:      90 * time.Second,
			MaxConnsPerHost:  10,
			MaxAttempts:      4,
			BackoffBaseMs:    300,
			BackoffJitterMs:  150,
			MaxDelayMs:       4000,
			QPS:              0.7,
			Burst:            1,
			CircuitWindow:    60 * time.Second,
			FailureThreshold: 5,
			ResetTimeout:     30 * time.Second,
			UserAgent:        "Mozilla/5.0 (Ampy yfinance-go scraper)",
			MaxBodyBytes:     8 << 20, // 8 MiB — scrape body cap.
		},
		Enabled:      true,
		RobotsPolicy: string(RobotsEnforce),
		CacheTTLMs:   60000,
		Endpoints: EndpointConfig{
			KeyStatistics: true,
			Financials:    true,
			Analysis:      true,
			Profile:       true,
			News:          true,
		},
	}
}

// RobotsPolicy represents the robots.txt policy
type RobotsPolicy string

const (
	RobotsEnforce RobotsPolicy = "enforce"
	RobotsWarn    RobotsPolicy = "warn"
	RobotsIgnore  RobotsPolicy = "ignore"
)

// IsValidRobotsPolicy checks if a robots policy is valid
func IsValidRobotsPolicy(policy string) bool {
	return policy == string(RobotsEnforce) ||
		policy == string(RobotsWarn) ||
		policy == string(RobotsIgnore)
}