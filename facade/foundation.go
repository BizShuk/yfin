package facade

import (
	"context"
	"fmt"
	"strings"
)

var foundationCommandOrder = []string{
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

// FoundationCommands returns the ordered Yahoo data surfaces supported by
// FetchFoundation.
func FoundationCommands() []string {
	return append([]string(nil), foundationCommandOrder...)
}

// FetchFoundation dispatches one yfinance-compatible foundation command
// through yfin-owned Yahoo API or scrape adapters.
func (c *Client) FetchFoundation(
	ctx context.Context,
	command string,
	symbol string,
	runID string,
) (any, error) {
	command = strings.TrimSpace(command)
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return nil, fmt.Errorf("facade: foundation symbol is required")
	}

	switch command {
	case "info":
		return c.YahooDispatch(ctx, "info", symbol)
	case "history":
		return c.YahooFetchHistory(ctx, symbol)
	case "actions":
		return c.YahooDispatch(ctx, "actions", symbol)
	case "income":
		return c.FetchIncomeStatement(ctx, symbol)
	case "balance":
		return c.FetchBalanceSheet(ctx, symbol)
	case "cashflow":
		return c.FetchCashFlowStatement(ctx, symbol)
	case "major-holders":
		return c.YahooDispatch(ctx, "major-holders", symbol)
	case "institutional-holders":
		return c.YahooDispatch(ctx, "institutional-holders", symbol)
	case "mutualfund-holders":
		return c.YahooDispatch(ctx, "mutualfund-holders", symbol)
	case "insider-transactions":
		return c.YahooDispatch(ctx, "insider-transactions", symbol)
	case "insider-purchases":
		return c.YahooDispatch(ctx, "insider-purchases", symbol)
	case "insider-roster":
		return c.YahooDispatch(ctx, "insider-roster", symbol)
	case "recommendations":
		return c.YahooDispatch(ctx, "recommendations", symbol)
	case "recommendations-summary":
		return c.YahooDispatch(ctx, "recommendations-summary", symbol)
	case "upgrades":
		return c.YahooDispatch(ctx, "upgrades", symbol)
	case "earnings-dates":
		return c.YahooDispatch(ctx, "earnings-dates", symbol)
	case "earnings-history":
		return c.ScrapeAnalysisDimension(ctx, "earnings-history", symbol, runID)
	case "eps-trend":
		return c.ScrapeAnalysisDimension(ctx, "eps-trend", symbol, runID)
	case "eps-revisions":
		return c.ScrapeAnalysisDimension(ctx, "eps-revisions", symbol, runID)
	case "earnings-estimates":
		return c.ScrapeAnalysisDimension(ctx, "earnings-estimates", symbol, runID)
	case "revenue-estimates":
		return c.ScrapeAnalysisDimension(ctx, "revenue-estimates", symbol, runID)
	case "growth-estimates":
		return c.ScrapeAnalysisDimension(ctx, "growth-estimates", symbol, runID)
	case "price-targets":
		return c.ScrapeAnalystInsights(ctx, symbol, runID)
	case "news":
		return c.FetchNews(ctx, symbol)
	case "calendar":
		return c.YahooDispatch(ctx, "calendar", symbol)
	case "sec-filings":
		return c.YahooDispatch(ctx, "sec-filings", symbol)
	case "sustainability":
		return c.YahooDispatch(ctx, "sustainability", symbol)
	case "isin":
		return c.YahooDispatch(ctx, "isin", symbol)
	case "options":
		return c.YahooDispatch(ctx, "options", symbol)
	case "metadata":
		return c.YahooDispatch(ctx, "metadata", symbol)
	default:
		return nil, fmt.Errorf("facade: unsupported foundation command %q", command)
	}
}

func isFoundationCommand(command string) bool {
	for _, supported := range foundationCommandOrder {
		if command == supported {
			return true
		}
	}
	return false
}
