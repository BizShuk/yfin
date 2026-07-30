// scrape.go — facade-level wrapper over `svc/scrape` so the yfin CLI can
// run scrape preview modes (`--check`, `--preview-json`, `--preview-news`)
// without importing svc/scrape directly.
//
// All exported entry points take a *Client receiver except the package-level
// parser wrappers, which are pure (no state) and so don't need one. Each
// parser is a thin pass-through to svc/scrape — the point isn't to add
// logic, it's to keep cmd → svc at one remove via facade.
//
// Return types use model.* DTOs because cmd/scrape formatters take them.
// facade owns the public ScrapeClient contract; cmd cannot directly call
// svc/scrape.Fetch / Parse* / etc.
package facade

import (
	"context"
	"fmt"
	"time"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/svc/scrape"
)

// ScrapeClient is the facade-owned transport contract for raw scrape fetches.
// svc/scrape.Client satisfies it without leaking that service type through
// facade signatures.
type ScrapeClient interface {
	Fetch(ctx context.Context, url string) ([]byte, *model.FetchMeta, error)
}

// ScrapeFetch issues a single page fetch against the Yahoo Finance scrape
// client. Wraps svc/scrape.Client.Fetch so cmd never touches it directly.
// Returns the raw response body plus the fetch metadata (host, status,
// bytes, redirects, etc.).
func (c *Client) ScrapeFetch(ctx context.Context, ticker, endpoint string) ([]byte, *model.FetchMeta, error) {
	scraper, err := c.configuredScrapeClient()
	if err != nil {
		return nil, nil, err
	}
	url := BuildScrapeURL(ticker, endpoint)
	body, meta, err := scraper.Fetch(ctx, url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	return body, meta, nil
}

// ScrapeFetchWithClient is a variant for callers that provide an explicit
// facade-owned ScrapeClient.
func ScrapeFetchWithClient(ctx context.Context, scraper ScrapeClient, ticker, endpoint string) ([]byte, *model.FetchMeta, error) {
	url := BuildScrapeURL(ticker, endpoint)
	return scraper.Fetch(ctx, url)
}

// configuredScrapeClient returns the ScrapeClient owned by this facade.Client.
func (c *Client) configuredScrapeClient() (ScrapeClient, error) {
	if c.scrapeClient == nil {
		return nil, fmt.Errorf("scrape client not initialised on facade.Client")
	}
	return c.scrapeClient, nil
}

// NewScrapeClientFromConfig builds a facade-owned ScrapeClient from the YAML
// ScrapeConfig. ConfigureScrape is preferred when an existing facade Client
// should retain one shared connection pool.
func NewScrapeClientFromConfig(cfg *config.ScrapeConfig) (ScrapeClient, error) {
	return scrape.NewClient(scrapeConfigFromConfig(cfg), nil)
}

// ConfigureScrape applies the loaded scrape policy to a facade Client while
// preserving its shared HTTP connection pool.
func (c *Client) ConfigureScrape(cfg *config.ScrapeConfig) error {
	if c == nil || c.httpClient == nil {
		return fmt.Errorf("facade client HTTP transport is not initialised")
	}

	scrapeConfig := scrapeConfigFromConfig(cfg)
	c.httpClient.UseProfile(yahooWebHostProfile(scrapeConfig.HTTP))
	client, err := scrape.NewClient(scrapeConfig, c.httpClient)
	if err != nil {
		return fmt.Errorf("configure scrape client: %w", err)
	}
	c.scrapeClient = client
	return nil
}

func scrapeConfigFromConfig(cfg *config.ScrapeConfig) *scrape.Config {
	if cfg == nil {
		return scrape.DefaultConfig()
	}

	httpConfig := cfg.HTTP
	if httpConfig == nil {
		httpConfig = scrape.DefaultConfig().HTTP
	}

	return &scrape.Config{
		HTTP:         httpConfig,
		Enabled:      cfg.Enabled,
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
}

// BuildScrapeURL builds the canonical Yahoo Finance scrape URL for a given
// ticker + endpoint. Endpoint keys are the same as the --endpoints flag
// accepts: profile, key-statistics, financials, balance-sheet, cash-flow,
// analysis, analyst-insights, news.
func BuildScrapeURL(ticker, endpoint string) string {
	const baseURL = "https://finance.yahoo.com"
	switch endpoint {
	case "profile":
		return baseURL + "/quote/" + ticker + "/profile"
	case "key-statistics":
		return baseURL + "/quote/" + ticker + "/key-statistics"
	case "financials":
		return baseURL + "/quote/" + ticker + "/financials"
	case "balance-sheet":
		return baseURL + "/quote/" + ticker + "/balance-sheet"
	case "cash-flow":
		return baseURL + "/quote/" + ticker + "/cash-flow"
	case "analysis":
		return baseURL + "/quote/" + ticker + "/analysis"
	case "analyst-insights":
		return baseURL + "/quote/" + ticker + "/analyst-insights"
	case "news":
		return baseURL + "/quote/" + ticker + "/news"
	default:
		return baseURL + "/quote/" + ticker
	}
}

// Parsers — thin pass-throughs to svc/scrape.Parse* so cmd can route
// preview-json logic through facade without holding an svc/scrape import for
// runtime calls (type imports for the returned DTO are fine).

// ParseComprehensiveFinancials parses the financials page body.
func ParseComprehensiveFinancials(body []byte, symbol, mic string) (*model.ComprehensiveFinancialsDTO, error) {
	return scrape.ParseComprehensiveFinancials(body, symbol, mic)
}

// ParseComprehensiveFinancialsWithCurrency parses financials using the
// currency code resolved from a second financials page body.
func ParseComprehensiveFinancialsWithCurrency(body, financialsBody []byte, symbol, mic string) (*model.ComprehensiveFinancialsDTO, error) {
	return scrape.ParseComprehensiveFinancialsWithCurrency(body, financialsBody, symbol, mic)
}

// ParseComprehensiveKeyStatistics parses the key-statistics page body.
func ParseComprehensiveKeyStatistics(body []byte, symbol, mic string) (*model.ComprehensiveKeyStatisticsDTO, error) {
	return scrape.ParseComprehensiveKeyStatistics(body, symbol, mic)
}

// ParseComprehensiveProfile parses the profile page body.
func ParseComprehensiveProfile(body []byte, symbol, mic string) (*model.ComprehensiveProfileDTO, error) {
	return scrape.ParseComprehensiveProfile(body, symbol, mic)
}

// ParseAnalysis parses the analysis page body.
func ParseAnalysis(body []byte, symbol, mic string) (*model.ComprehensiveAnalysisDTO, error) {
	return scrape.ParseAnalysis(body, symbol, mic)
}

// ParseAnalystInsights parses the analyst-insights page body.
func ParseAnalystInsights(body []byte, symbol, mic string) (*model.AnalystInsightsDTO, error) {
	return scrape.ParseAnalystInsights(body, symbol, mic)
}

// ParseNews parses the news page body into a slice of articles + stats.
func ParseNews(body []byte, baseURL string, now time.Time) ([]model.ScrapeNewsItem, *model.NewsStats, error) {
	return scrape.ParseNews(body, baseURL, now)
}
