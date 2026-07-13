// scrape.go — web-scraping engine settings (HTML pages on
// finance.yahoo.com that the JSON API doesn't return). The nested
// `ScrapeEndpointConfig` toggles which endpoints the scraper is allowed
// to hit (key_statistics / financials / analysis / profile / news).
// Capacity: 3 structs (`ScrapeConfig`, `ScrapeRetryConfig`,
// `ScrapeEndpointConfig`).
package types

// ScrapeConfig represents scraping configuration
type ScrapeConfig struct {
	Enabled      bool                 `yaml:"enabled"`
	UserAgent    string               `yaml:"user_agent"`
	TimeoutMs    int                  `yaml:"timeout_ms"`
	QPS          float64              `yaml:"qps"`
	Burst        int                  `yaml:"burst"`
	Retry        ScrapeRetryConfig    `yaml:"retry"`
	RobotsPolicy string               `yaml:"robots_policy"`
	CacheTTLMs   int                  `yaml:"cache_ttl_ms"`
	Endpoints    ScrapeEndpointConfig `yaml:"endpoints"`
}

// ScrapeRetryConfig represents scraping retry configuration
type ScrapeRetryConfig struct {
	Attempts   int `yaml:"attempts"`
	BaseMs     int `yaml:"base_ms"`
	MaxDelayMs int `yaml:"max_delay_ms"`
}

// ScrapeEndpointConfig represents endpoint-specific scraping configuration
type ScrapeEndpointConfig struct {
	KeyStatistics bool `yaml:"key_statistics"`
	Financials    bool `yaml:"financials"`
	Analysis      bool `yaml:"analysis"`
	Profile       bool `yaml:"profile"`
	News          bool `yaml:"news"`
}
