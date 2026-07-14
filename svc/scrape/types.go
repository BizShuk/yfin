// configuration (`Config`, `RetryConfig`, `EndpointConfig`, `RobotsPolicy`,
// `RobotsEnforce/Warn/Ignore`, `IsValidRobotsPolicy`, `DefaultConfig`).
// All scrape data structs live in model/ (model.FetchMeta, model.ScrapeNewsItem,
// model.NewsStats, model.RobotsRule, etc.); this file only retains
// scrape-package-internal configuration types and constructors.

package scrape

import "time"


// model.ScrapeNewsItem is the raw scrape-shape news article (carries ImageURL +
// RelatedTickers); distinct from model.NewsItem which is the cleaned SDK
// scrape-internal code.


// Config represents the scraping configuration
type Config struct {
	Enabled      bool           `yaml:"enabled"`
	UserAgent    string         `yaml:"user_agent"`
	TimeoutMs    int            `yaml:"timeout_ms"`
	QPS          float64        `yaml:"qps"`
	Burst        int            `yaml:"burst"`
	Retry        RetryConfig    `yaml:"retry"`
	RobotsPolicy string         `yaml:"robots_policy"`
	CacheTTLMs   int            `yaml:"cache_ttl_ms"`
	Endpoints    EndpointConfig `yaml:"endpoints"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	Attempts   int `yaml:"attempts"`
	BaseMs     int `yaml:"base_ms"`
	MaxDelayMs int `yaml:"max_delay_ms"`
}

// EndpointConfig represents endpoint-specific configuration
type EndpointConfig struct {
	KeyStatistics bool `yaml:"key_statistics"`
	Financials    bool `yaml:"financials"`
	Analysis      bool `yaml:"analysis"`
	Profile       bool `yaml:"profile"`
	News          bool `yaml:"news"`
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:   true,
		UserAgent: "Mozilla/5.0 (Ampy yfinance-go scraper)",
		TimeoutMs: 10000,
		QPS:       0.7,
		Burst:     1,
		Retry: RetryConfig{
			Attempts:   4,
			BaseMs:     300,
			MaxDelayMs: 4000,
		},
		RobotsPolicy: "enforce",
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

// model.RobotsRule, model.RobotsCache, model.BackoffPolicyConfig, model.RateLimitConfig are
// from svc/scrape directly.

// (time import retained for DefaultConfig potential future timestamp fields.)
var _ = time.Now
