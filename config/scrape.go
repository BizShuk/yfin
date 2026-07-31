// scrape.go — web-scraping engine settings (HTML pages on
// finance.yahoo.com that the JSON API doesn't return). The nested
// `ScrapeEndpointConfig` toggles which endpoints the scraper is allowed
// to hit (key_statistics / financials / analysis / profile / news).
//
// HTTP-layer knobs (timeout, qps, retry, user_agent, body cap) are NOT
// declared here — they live in `utils/httpx.Config` and the loader
// populates `HTTP *httpx.Config` from `httpx.DefaultScrapeConfig()`.
// To customise, callers mutate `cfg.Scrape.HTTP` after loading and apply
// the resulting config through `(*facade.Client).ConfigureScrape`.
package config

import "github.com/bizshuk/yfin/utils/httpx"

// ScrapeConfig represents scraping configuration
type ScrapeConfig struct {
	// HTTP is the assembled HTTP-layer config for scrape traffic. The
	// loader populates this from `httpx.DefaultScrapeConfig()` after Load().
	// Override the returned config through Go code when needed.
	HTTP *httpx.Config `yaml:"-"`

	// Scrape-only knobs (no HTTP-layer concerns).
	Enabled      bool                 `yaml:"enabled"`
	RobotsPolicy string               `yaml:"robots_policy"`
	CacheTTLMs   int                  `yaml:"cache_ttl_ms"`
	Endpoints    ScrapeEndpointConfig `yaml:"endpoints"`
}

// ScrapeEndpointConfig represents endpoint-specific scraping configuration
type ScrapeEndpointConfig struct {
	KeyStatistics bool `yaml:"key_statistics"`
	Financials    bool `yaml:"financials"`
	Analysis      bool `yaml:"analysis"`
	Profile       bool `yaml:"profile"`
	News          bool `yaml:"news"`
}
