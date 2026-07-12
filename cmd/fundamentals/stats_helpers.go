// stats_helpers.go — minimal scrape.Client constructor + URL builder shared
// by fundamentals/stats.go and fundamentals/profile.go. The scrape sub-package
// already defines buildScrapeURL/createScrapeClient — duplicating them here
// avoids a fundamentals → scrape cross-package dep. Capacity: 2 helpers.
package fundamentals

import (
	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/svc/scrape"
)

// scrapeNewClient creates a scrape.Client from ampy-config. Mirrors
// scrape/format.go's createScrapeClient — duplicated to keep this sub-package
// independent of scrape (no cross-package dep).
func scrapeNewClient(cfg *config.ScrapeConfig) (scrape.Client, error) {
	scrapeCfg := &scrape.Config{
		Enabled:   cfg.Enabled,
		UserAgent: cfg.UserAgent,
		TimeoutMs: cfg.TimeoutMs,
		QPS:       cfg.QPS,
		Burst:     cfg.Burst,
		Retry: scrape.RetryConfig{
			Attempts:   cfg.Retry.Attempts,
			BaseMs:     cfg.Retry.BaseMs,
			MaxDelayMs: cfg.Retry.MaxDelayMs,
		},
		RobotsPolicy: cfg.RobotsPolicy,
		CacheTTLMs:   cfg.CacheTTLMs,
		Endpoints: scrape.EndpointConfig{
			KeyStatistics: cfg.Endpoints.KeyStatistics,
			Financials:    cfg.Endpoints.Financials,
			Analysis:      cfg.Endpoints.Analysis,
			Profile:       cfg.Endpoints.Profile,
			News:          cfg.Endpoints.News,
		},
	}
	return scrape.NewClient(scrapeCfg, nil)
}

// buildScrapeURL mirrors scrape/scrape_run.go's buildScrapeURL — same Yahoo
// Finance URL shape for both key-statistics and profile endpoints.
func buildScrapeURL(ticker, endpoint string) string {
	const baseURL = "https://finance.yahoo.com"
	switch endpoint {
	case "profile":
		return baseURL + "/quote/" + ticker + "/profile"
	case "key-statistics":
		return baseURL + "/quote/" + ticker + "/key-statistics"
	default:
		return baseURL + "/quote/" + ticker
	}
}
