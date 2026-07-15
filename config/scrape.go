// scrape.go — web-scraping engine settings (HTML pages on
// finance.yahoo.com that the JSON API doesn't return). The nested
// `ScrapeEndpointConfig` toggles which endpoints the scraper is allowed
// to hit (key_statistics / financials / analysis / profile / news).
//
// HTTP-layer knobs (timeout, qps, retry, user_agent, body cap) are NOT
// declared here — they live in `utils/httpx.Config` and the loader
// populates `HTTP *httpx.Config` from `svc/scrape.DefaultConfig().HTTP`.
// To customise, callers pass a tuned `*httpx.Config` via Go code
// (e.g., `facade.NewScrapeClientWithHTTP` or by mutating
// `cfg.Scrape.HTTP` before `scrape.NewClient`).
//
// Capacity: 2 structs (`ScrapeConfig`, `ScrapeEndpointConfig`).
package config

import "github.com/bizshuk/yfin/utils/httpx"

// ScrapeConfig represents scraping configuration
type ScrapeConfig struct {
	// HTTP is the assembled HTTP-layer config for scrape traffic. The
	// loader populates this from `svc/scrape.DefaultConfig().HTTP` after
	// Load(), so yaml users get sensible scrape-tuned defaults (QPS 0.7,
	// 4 retries with 300–4000 ms backoff, 8 MiB body cap, robots.txt
	// enforced). Override via Go code if needed.
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