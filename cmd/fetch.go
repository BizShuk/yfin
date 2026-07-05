// cmd/fetch.go — thin helpers that route the yfin CLI through svc/yahoo and
// internal/norm so the CLI keeps the full ScaledDecimal/emit proto precision
// needed for bus publishing. As of Step 6 of plans/spicy-singing-swan.md, the
// public facade.Client returns plain SDK structs (facade.BarBatch, facade.Quote,
// facade.FundamentalsSnapshot, facade.MarketData); the CLI's print*/handle*
// code still wants the norm.* shape, so it calls these helpers directly
// instead of going through the SDK surface. Capacity: 5 fetch helpers (`fetchDailyBarsNorm`/`fetchQuoteNorm`/`fetchFundamentalsNorm`/`fetchMarketDataNorm`/`isPaidYahooAuthError`).

package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bizshuk/yfin/svc/norm"
	"github.com/bizshuk/yfin/svc/yahoo"
)

// fetchDailyBarsNorm fetches daily bars via yahoo.Client and returns the
// internal *norm.NormalizedBarBatch that emit.EmitBarBatch expects.
//
// Mirrors facade.Client.FetchDailyBars but skips the FromBarBatch conversion —
// the yfin CLI's downstream code (printBarsPreview, handleBusPublishing)
// already handles norm.* types and uses ScaledDecimal for sub-cent precision.
func fetchDailyBarsNorm(ctx context.Context, yahooClient *yahoo.Client, symbol string, start, end time.Time, adjusted bool, runID string) (*norm.NormalizedBarBatch, error) {
	if yahooClient == nil {
		return nil, fmt.Errorf("fetchDailyBarsNorm: yahooClient is nil")
	}

	barsResp, err := yahooClient.FetchDailyBars(ctx, symbol, start, end, adjusted)
	if err != nil {
		return nil, err
	}

	bars, err := barsResp.GetBars()
	if err != nil {
		return nil, err
	}

	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	return norm.NormalizeBars(bars, meta, runID)
}

// fetchQuoteNorm fetches a quote via yahoo.Client and returns the internal
// *norm.NormalizedQuote. Mirrors facade.Client.FetchQuote's internals but
// skips the FromQuote conversion.
func fetchQuoteNorm(ctx context.Context, yahooClient *yahoo.Client, symbol, runID string) (*norm.NormalizedQuote, error) {
	if yahooClient == nil {
		return nil, fmt.Errorf("fetchQuoteNorm: yahooClient is nil")
	}

	quoteResp, err := yahooClient.FetchQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}

	quotes := quoteResp.GetQuotes()
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes found")
	}

	return norm.NormalizeQuote(quotes[0], runID)
}

// fetchFundamentalsNorm fetches quarterly fundamentals via yahoo.Client and
// returns *norm.NormalizedFundamentalsSnapshot. The 401-class error handling
// is preserved so the CLI can emit ExitPaidFeature.
func fetchFundamentalsNorm(ctx context.Context, yahooClient *yahoo.Client, symbol, runID string) (*norm.NormalizedFundamentalsSnapshot, error) {
	if yahooClient == nil {
		return nil, fmt.Errorf("fetchFundamentalsNorm: yahooClient is nil")
	}

	fundResp, err := yahooClient.FetchFundamentalsQuarterly(ctx, symbol)
	if err != nil {
		if isPaidYahooAuthError(err) {
			return nil, fmt.Errorf("fundamentals data requires Yahoo Finance paid subscription: %w", err)
		}
		return nil, err
	}

	fundamentals, err := fundResp.GetFundamentals()
	if err != nil {
		return nil, err
	}

	return norm.NormalizeFundamentals(fundamentals, symbol, runID)
}

// fetchMarketDataNorm fetches comprehensive market data via yahoo.Client and
// returns *norm.NormalizedMarketData. Mirrors facade.Client.FetchMarketData.
func fetchMarketDataNorm(ctx context.Context, yahooClient *yahoo.Client, symbol, runID string) (*norm.NormalizedMarketData, error) {
	if yahooClient == nil {
		return nil, fmt.Errorf("fetchMarketDataNorm: yahooClient is nil")
	}

	// Use the same chart-endpoint window (last 1 day) that facade.Client uses.
	end := time.Now()
	start := end.AddDate(0, 0, -1)

	barsResp, err := yahooClient.FetchDailyBars(ctx, symbol, start, end, true)
	if err != nil {
		return nil, err
	}

	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	return norm.NormalizeMarketData(meta, runID)
}

// isPaidYahooAuthError mirrors facade.Client.isAuthenticationError. The CLI
// uses this to surface ExitPaidFeature (code 2) instead of a generic error.
func isPaidYahooAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "Unauthorized") ||
		strings.Contains(errStr, "authentication")
}
