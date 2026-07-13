// config.go — root `Config` struct aggregating every sub-domain config
// (app / yahoo / concurrency / rate_limit / sessions / retry /
// circuit_breaker / markets / fx / bus / scrape / observability /
// secrets) into one tree that maps 1:1 onto the on-disk YAML. Adapter
// methods live in adapters.go. Capacity: 1 struct (`Config`).
package types

// Config represents the complete configuration for yfinance-go
type Config struct {
	App            AppConfig            `yaml:"app"`
	Yahoo          YahooConfig          `yaml:"yahoo"`
	Concurrency    ConcurrencyConfig    `yaml:"concurrency"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	Sessions       SessionsConfig       `yaml:"sessions"`
	Retry          RetryConfig          `yaml:"retry"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Markets        MarketsConfig        `yaml:"markets"`
	FX             FXConfig             `yaml:"fx"`
	Bus            BusConfig            `yaml:"bus"`
	Scrape         ScrapeConfig         `yaml:"scrape"`
	Observability  ObservabilityConfig  `yaml:"observability"`
	Secrets        []SecretConfig       `yaml:"secrets"`
}
