// scrape.go — facade-level wrapper over `svc/scrape` so the yfin CLI can
// run scrape preview modes (`--check`, `--preview-json`, `--preview-news`,
// `--preview-proto`) without importing svc/scrape directly.
//
// All exported entry points take a *Client receiver except the package-level
// parser wrappers, which are pure (no state) and so don't need one. Each
// parser is a thin pass-through to svc/scrape — the point isn't to add
// logic, it's to keep cmd → svc at one remove via facade.
//
// The return types still reference svc/scrape DTOs (ComprehensiveFinancialsDTO,
// etc.) because cmd/scrape/format*.go printers take them. Type imports from
// svc/scrape are allowed (these are pure data declarations, not runtime
// calls); cmd cannot directly call svc/scrape.Fetch / Parse* / etc.
package facade

import (
	"context"
	"fmt"
	"time"

	"github.com/bizshuk/yfin/config/types"
	"github.com/bizshuk/yfin/svc/scrape"
)

// ScrapeFetch issues a single page fetch against the Yahoo Finance scrape
// client. Wraps svc/scrape.Client.Fetch so cmd never touches it directly.
// Returns the raw response body plus the fetch metadata (host, status,
// bytes, redirects, etc.).
func (c *Client) ScrapeFetch(ctx context.Context, ticker, endpoint string) ([]byte, *scrape.FetchMeta, error) {
	scraper, err := c.scrapeClientForCmd()
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

// ScrapeFetchWithClient is a variant of ScrapeFetch that accepts an explicit
// scrape.Client (used by cmd/scrape preview modes that wire a custom client
// via `cmd.CreateScrapeClient` today; after Phase C, callers can use either
// the facade-bound client or a locally-built one).
func ScrapeFetchWithClient(ctx context.Context, scraper scrape.Client, ticker, endpoint string) ([]byte, *scrape.FetchMeta, error) {
	url := BuildScrapeURL(ticker, endpoint)
	return scraper.Fetch(ctx, url)
}

// scrapeClientForCmd returns the scrape.Client owned by this facade.Client.
// It exists as a thin accessor so callers needing raw fetch access (rare;
// preview modes mostly use ScrapeFetch above) can drive it without re-
// importing svc/scrape.
func (c *Client) scrapeClientForCmd() (scrape.Client, error) {
	if c.scrapeClient == nil {
		return nil, fmt.Errorf("scrape client not initialised on facade.Client")
	}
	return c.scrapeClient, nil
}

// NewScrapeClientFromConfig builds a svc/scrape.Client from ampy-config's
// flat ScrapeConfig struct. Wraps the same field-by-field mapping that
// cmd/scrape/ previously did inline.
func NewScrapeClientFromConfig(cfg *types.ScrapeConfig) (scrape.Client, error) {
	return scrape.NewClient(&scrape.Config{
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
	}, nil)
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
func ParseComprehensiveFinancials(body []byte, symbol, mic string) (*scrape.ComprehensiveFinancialsDTO, error) {
	return scrape.ParseComprehensiveFinancials(body, symbol, mic)
}

// ParseComprehensiveFinancialsWithCurrency parses financials using the
// currency code resolved from a second financials page body.
func ParseComprehensiveFinancialsWithCurrency(body, financialsBody []byte, symbol, mic string) (*scrape.ComprehensiveFinancialsDTO, error) {
	return scrape.ParseComprehensiveFinancialsWithCurrency(body, financialsBody, symbol, mic)
}

// ParseComprehensiveKeyStatistics parses the key-statistics page body.
func ParseComprehensiveKeyStatistics(body []byte, symbol, mic string) (*scrape.ComprehensiveKeyStatisticsDTO, error) {
	return scrape.ParseComprehensiveKeyStatistics(body, symbol, mic)
}

// ParseComprehensiveProfile parses the profile page body.
func ParseComprehensiveProfile(body []byte, symbol, mic string) (*scrape.ComprehensiveProfileDTO, error) {
	return scrape.ParseComprehensiveProfile(body, symbol, mic)
}

// ParseAnalysis parses the analysis page body.
func ParseAnalysis(body []byte, symbol, mic string) (*scrape.ComprehensiveAnalysisDTO, error) {
	return scrape.ParseAnalysis(body, symbol, mic)
}

// ParseAnalystInsights parses the analyst-insights page body.
func ParseAnalystInsights(body []byte, symbol, mic string) (*scrape.AnalystInsightsDTO, error) {
	return scrape.ParseAnalystInsights(body, symbol, mic)
}

// ParseNews parses the news page body into a slice of articles + stats.
func ParseNews(body []byte, baseURL string, now time.Time) ([]scrape.NewsItem, *scrape.NewsStats, error) {
	return scrape.ParseNews(body, baseURL, now)
}
