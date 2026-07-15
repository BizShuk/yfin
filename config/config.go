// config.go — root `Config` struct aggregating every sub-domain config
// (app / yahoo / concurrency / rate_limit / retry / circuit_breaker /
// markets / fx / scrape / observability / secrets) into one tree that
// maps 1:1 onto the on-disk YAML. Adapter methods live in adapters.go.
// The `HTTP` and `Scrape.HTTP` fields hold the post-load assembled
// `*httpx.Config` (the canonical HTTP type), populated once in
// `Loader.Load`. Capacity: 1 struct (`Config`).
package config

import "github.com/bizshuk/yfin/utils/httpx"

// Config represents the complete configuration for yfinance-go
type Config struct {
	App            AppConfig            `yaml:"app"`
	Yahoo          YahooConfig          `yaml:"yahoo"`
	Concurrency    ConcurrencyConfig    `yaml:"concurrency"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	Retry          RetryConfig          `yaml:"retry"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Markets        MarketsConfig        `yaml:"markets"`
	FX             FXConfig             `yaml:"fx"`
	Scrape         ScrapeConfig         `yaml:"scrape"`
	Observability  ObservabilityConfig  `yaml:"observability"`
	Secrets        []SecretConfig       `yaml:"secrets"`

	// HTTP is the assembled HTTP-layer config used by every caller
	// (cmd/client.go, cmd/scrape, etc.). Nil until Load() succeeds.
	HTTP *httpx.Config `yaml:"-"`
}
