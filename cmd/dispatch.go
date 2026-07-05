package cmd

import (
	"context"
	"time"

	"github.com/bizshuk/yfinance-go/facade"
	"github.com/bizshuk/yfinance-go/svc/yahoo"
)

// FetchContext bundles everything a command needs to fetch its data.
type FetchContext struct {
	Root  *facade.Client // top-level client (provides Scrape*, Fetch*)
	Y     *yahoo.Client
	RunID string
}

// fetchFunc fetches a single command's data; result must be JSON-marshalable.
type fetchFunc func(ctx context.Context, fc *FetchContext, symbol string) (any, error)

// commandRegistry maps Python-style command names to fetchers.
var commandRegistry = map[string]fetchFunc{
	// New quoteSummary-based fetchers (authed yahoo.Client).
	"info": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchInfo(ctx, s)
	},
	"actions": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchActions(ctx, s)
	},
	"metadata": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchMetadata(ctx, s)
	},
	"major-holders": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchHolders(ctx, s)
	},
	"institutional-holders": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchHolders(ctx, s)
	},
	"mutualfund-holders": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchHolders(ctx, s)
	},
	"insider-transactions": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchInsider(ctx, s)
	},
	"insider-purchases": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchInsider(ctx, s)
	},
	"insider-roster": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchInsider(ctx, s)
	},
	"upgrades": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchUpgrades(ctx, s)
	},
	"calendar": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchCalendar(ctx, s)
	},
	"earnings-dates": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchEarningsDates(ctx, s)
	},
	"sec-filings": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchSecFilings(ctx, s)
	},
	"sustainability": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchESG(ctx, s)
	},
	"recommendations": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchRecommendationTrend(ctx, s)
	},
	"recommendations-summary": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchRecommendationTrend(ctx, s)
	},
	"options": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchOptions(ctx, s)
	},
	"isin": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Y.FetchISIN(ctx, s)
	},
	// Legacy chart-based history.
	"history": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.FetchDailyBars(ctx, s, time.Now().AddDate(0, 0, -30), time.Now(), true, fc.RunID)
	},
	// Legacy scrape-based fundamentals/analysis/insights/news.
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
		return fc.Root.ScrapeAnalysis(ctx, s, fc.RunID)
	},
	"eps-trend": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysis(ctx, s, fc.RunID)
	},
	"eps-revisions": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysis(ctx, s, fc.RunID)
	},
	"earnings-estimates": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysis(ctx, s, fc.RunID)
	},
	"revenue-estimates": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysis(ctx, s, fc.RunID)
	},
	"growth-estimates": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalysis(ctx, s, fc.RunID)
	},
	"price-targets": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeAnalystInsights(ctx, s, fc.RunID)
	},
	"news": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
		return fc.Root.ScrapeNews(ctx, s, fc.RunID)
	},
}
