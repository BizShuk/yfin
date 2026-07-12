// fetch.go — thin helpers that route the yfin CLI through svc/yahoo and
// svc/norm so the CLI keeps the full ScaledDecimal/emit proto precision
// needed for bus publishing. After Step 6 of plans/spicy-singing-swan.md,
// the public facade.Client returns plain SDK structs; the CLI's
// print*/handle* code still wants the norm.* shape, so it calls these
// helpers directly instead of going through the SDK surface.
//
// The helpers are EXPORTED so that sub-packages (cmd/market, cmd/fundamentals)
// can call them. They bridge svc/yahoo → svc/norm and stay co-located with
// cmd/ root rather than in each sub-package to keep the wiring logic in
// one place.
//
// Capacity: 4 fetch helpers (FetchDailyBarsNorm / FetchQuoteNorm /
// FetchFundamentalsNorm / FetchMarketDataNorm) + 1 paid-auth detector.

package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bizshuk/yfin/svc/norm"
	"github.com/bizshuk/yfin/svc/yahoo"
)

// FetchDailyBarsNorm fetches daily bars via yahoo.Client and returns the
// internal *norm.NormalizedBarBatch that emit.EmitBarBatch expects.
//
// Mirrors facade.Client.FetchDailyBars but skips the FromBarBatch conversion —
// the yfin CLI's downstream code (printBarsPreview, handleBusPublishing)
// already handles norm.* types and uses ScaledDecimal for sub-cent precision.
func FetchDailyBarsNorm(ctx context.Context, yahooClient *yahoo.Client, symbol string, start, end time.Time, adjusted bool, runID string) (*norm.NormalizedBarBatch, error) {
	if yahooClient == nil {
		return nil, fmt.Errorf("FetchDailyBarsNorm: yahooClient is nil")
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

// FetchQuoteNorm fetches a quote via yahoo.Client and returns the internal
// *norm.NormalizedQuote. Mirrors facade.Client.FetchQuote's internals but
// skips the FromQuote conversion.
func FetchQuoteNorm(ctx context.Context, yahooClient *yahoo.Client, symbol, runID string) (*norm.NormalizedQuote, error) {
	if yahooClient == nil {
		return nil, fmt.Errorf("FetchQuoteNorm: yahooClient is nil")
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

// FetchFundamentalsNorm fetches quarterly fundamentals via yahoo.Client and
// returns *norm.NormalizedFundamentalsSnapshot. The 401-class error handling
// is preserved so the CLI can emit ExitPaidFeature.
func FetchFundamentalsNorm(ctx context.Context, yahooClient *yahoo.Client, symbol, runID string) (*norm.NormalizedFundamentalsSnapshot, error) {
	if yahooClient == nil {
		return nil, fmt.Errorf("FetchFundamentalsNorm: yahooClient is nil")
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

// FetchMarketDataNorm fetches comprehensive market data via yahoo.Client and
// returns *norm.NormalizedMarketData. Mirrors facade.Client.FetchMarketData.
func FetchMarketDataNorm(ctx context.Context, yahooClient *yahoo.Client, symbol, runID string) (*norm.NormalizedMarketData, error) {
	if yahooClient == nil {
		return nil, fmt.Errorf("FetchMarketDataNorm: yahooClient is nil")
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
