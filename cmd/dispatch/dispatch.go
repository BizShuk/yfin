// dispatch.go — `commandRegistry` mapping 30+ Python-style yfinance-style
// command names (`info`/`history`/`actions`/`income`/`news`/`options`/
// `isin`/...) to fetch closures over `FetchContext` (`*facade.Client`).
//
// All svc calls route through facade (`YahooDispatch` / `ScrapeFinancials` /
// etc.); dispatch.go never imports svc/yahoo or svc/scrape directly.
//
// Capacity: 1 `FetchContext` struct + 1 `fetchFunc` type + 30+ registry
// entries.
package dispatch

import (
	"context"

	"github.com/bizshuk/yfin/facade"
)

// FetchContext bundles everything a command needs to fetch its data.
type FetchContext struct {
	Root  *facade.Client // top-level client (provides Fetch*, YahooDispatch, Scrape*)
	RunID string
}

// fetchFunc fetches a single command's data; result must be JSON-marshalable.
type fetchFunc func(ctx context.Context, fc *FetchContext, symbol string) (any, error)

// commandOrder mirrors skills/scripts/config.py and is the canonical batch
// execution order used by parity checks and production dispatch.
var commandOrder = []string{
	"info",
	"history",
	"actions",
	"income",
	"balance",
	"cashflow",
	"major-holders",
	"institutional-holders",
	"mutualfund-holders",
	"insider-transactions",
	"insider-purchases",
	"insider-roster",
	"recommendations",
	"recommendations-summary",
	"upgrades",
	"earnings-dates",
	"earnings-history",
	"eps-trend",
	"eps-revisions",
	"earnings-estimates",
	"revenue-estimates",
	"growth-estimates",
	"price-targets",
	"news",
	"calendar",
	"sec-filings",
	"sustainability",
	"isin",
	"options",
	"metadata",
}

// commandRegistry maps Python-style command names to fetchers. All entries
// route through facade — never svc/yahoo or svc/scrape directly.
var commandRegistry = map[string]fetchFunc{
	// Yahoo chart / quoteSummary-based fetchers (authed yahoo.Client).
	"info": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "info", s)
	},
	"actions": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "actions", s)
	},
	"metadata": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "metadata", s)
	},
	"major-holders": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "major-holders", s)
	},
	"institutional-holders": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "institutional-holders", s)
	},
	"mutualfund-holders": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "mutualfund-holders", s)
	},
	"insider-transactions": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "insider-transactions", s)
	},
	"insider-purchases": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "insider-purchases", s)
	},
	"insider-roster": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "insider-roster", s)
	},
	"upgrades": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "upgrades", s)
	},
	"calendar": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "calendar", s)
	},
	"earnings-dates": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "earnings-dates", s)
	},
	"sec-filings": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "sec-filings", s)
	},
	"sustainability": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "sustainability", s)
	},
	"recommendations": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "recommendations", s)
	},
	"recommendations-summary": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "recommendations-summary", s)
	},
	"options": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "options", s)
	},
	"isin": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooDispatch(ctx, "isin", s)
	},
	// Legacy chart-based history — 30-day lookback window via facade.
	"history": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.YahooFetchHistory(ctx, s)
	},
	// Legacy scrape-based fundamentals/analysis/insights/news — all route
	// through facade.Scrape* which already wraps svc/scrape.
	"income": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeFinancials(ctx, s, fc.RunID)
	},
	"balance": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeBalanceSheet(ctx, s, fc.RunID)
	},
	"cashflow": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeCashFlow(ctx, s, fc.RunID)
	},
	"earnings-history": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysisDimension(ctx, "earnings-history", s, fc.RunID)
	},
	"eps-trend": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysisDimension(ctx, "eps-trend", s, fc.RunID)
	},
	"eps-revisions": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysisDimension(ctx, "eps-revisions", s, fc.RunID)
	},
	"earnings-estimates": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysisDimension(ctx, "earnings-estimates", s, fc.RunID)
	},
	"revenue-estimates": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysisDimension(ctx, "revenue-estimates", s, fc.RunID)
	},
	"growth-estimates": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysisDimension(ctx, "growth-estimates", s, fc.RunID)
	},
	"price-targets": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalystInsights(ctx, s, fc.RunID)
	},
	"news": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeNews(ctx, s, fc.RunID)
	},
}
