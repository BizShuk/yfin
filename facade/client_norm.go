// client_norm.go — the ScaledDecimal-preserving surface of facade.Client.
//
// Norm-returning helpers — same fetch logic as FetchDailyBars / FetchQuote /
// FetchFundamentalsQuarterly / FetchMarketData, but stop at the
// `model.Normalize*` step and return the internal `*model.Normalized*` types
// instead of converting to plain SDK. These exist for callers that need
// ScaledDecimal precision (e.g. cmd's emit→proto pipeline). External SDK
// consumers should keep using the plain Fetch* methods.
//
// Naming convention: `Fetch<Kind>Norm` returns the Norm-internal shape.

package facade

import (
	"context"
	"fmt"
	"time"

	"github.com/bizshuk/yfin/model"
)

// FetchDailyBarsNorm fetches daily bars and returns the internal
// *model.NormalizedBarBatch (ScaledDecimal precision preserved).
func (c *Client) FetchDailyBarsNorm(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*model.NormalizedBarBatch, error) {
	barsResp, err := c.yahooClient.FetchDailyBars(ctx, symbol, start, end, adjusted)
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

	return model.NormalizeBars(bars, meta, runID)
}

// FetchQuoteNorm fetches a quote and returns the internal
// *model.NormalizedQuote.
func (c *Client) FetchQuoteNorm(ctx context.Context, symbol string, runID string) (*model.NormalizedQuote, error) {
	quoteResp, err := c.yahooClient.FetchQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}

	quotes := quoteResp.GetQuotes()
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes found")
	}

	return model.NormalizeQuote(quotes[0], runID)
}

// FetchFundamentalsNorm fetches quarterly fundamentals and returns the
// internal *model.NormalizedFundamentalsSnapshot. The 401-class error
// handling is preserved (paid-subscription endpoint) so callers can detect
// via the "paid subscription" / "401" / "Unauthorized" string in the error.
func (c *Client) FetchFundamentalsNorm(ctx context.Context, symbol string, runID string) (*model.NormalizedFundamentalsSnapshot, error) {
	fundResp, err := c.yahooClient.FetchFundamentalsQuarterly(ctx, symbol)
	if err != nil {
		if isAuthenticationError(err) {
			return nil, fmt.Errorf("fundamentals data requires Yahoo Finance paid subscription: %w", err)
		}
		return nil, err
	}

	fundamentals, err := fundResp.GetFundamentals()
	if err != nil {
		return nil, err
	}

	return model.NormalizeFundamentals(fundamentals, symbol, runID)
}

// FetchMarketDataNorm fetches comprehensive market data and returns the
// internal *model.NormalizedMarketData. Uses the same chart-endpoint window
// (last 1 day) as FetchMarketData.
func (c *Client) FetchMarketDataNorm(ctx context.Context, symbol string, runID string) (*model.NormalizedMarketData, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -1)

	barsResp, err := c.yahooClient.FetchDailyBars(ctx, symbol, start, end, true)
	if err != nil {
		return nil, err
	}

	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	return model.NormalizeMarketData(meta, runID)
}
