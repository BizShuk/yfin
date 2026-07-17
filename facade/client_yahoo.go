// client_yahoo.go — the Yahoo API surface of facade.Client: chart data,
// annual fundamentals-timeseries statements, and the ticker news XHR. All
// methods return plain, reflection-free SDK/model structs.
package facade

import (
	"context"
	"fmt"
	"time"

	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/svc/yahoo"
)

// FetchDailyBars fetches daily bars for a symbol and returns the plain SDK
// BarBatch (float64 prices, decoupled from the internal ScaledDecimal).
func (c *Client) FetchDailyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*BarBatch, error) {
	// Fetch raw data
	barsResp, err := c.yahooClient.FetchDailyBars(ctx, symbol, start, end, adjusted)
	if err != nil {
		return nil, err
	}

	// Extract bars and metadata
	bars, err := barsResp.GetBars()
	if err != nil {
		return nil, err
	}

	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	// Normalize then convert to the plain SDK surface.
	batch, err := model.NormalizeBars(bars, meta, runID)
	if err != nil {
		return nil, err
	}
	return FromBarBatch(batch), nil
}

// FetchQuote fetches a quote for a symbol and returns the plain SDK Quote.
func (c *Client) FetchQuote(ctx context.Context, symbol string, runID string) (*Quote, error) {
	// Fetch raw data
	quoteResp, err := c.yahooClient.FetchQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}

	// Extract quotes
	quotes := quoteResp.GetQuotes()
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes found")
	}

	// Normalize then convert to the plain SDK surface.
	q, err := model.NormalizeQuote(quotes[0], runID)
	if err != nil {
		return nil, err
	}
	return FromQuote(q), nil
}

// FetchFundamentalsQuarterly fetches quarterly fundamentals for a symbol and
// returns the plain SDK FundamentalsSnapshot. Note: this endpoint requires the
// Yahoo Finance paid subscription; 401-class errors are surfaced with a clearer
// message that callers can detect via isAuthenticationError.
func (c *Client) FetchFundamentalsQuarterly(ctx context.Context, symbol string, runID string) (*FundamentalsSnapshot, error) {
	// Fetch raw data
	fundResp, err := c.yahooClient.FetchFundamentalsQuarterly(ctx, symbol)
	if err != nil {
		if isAuthenticationError(err) {
			return nil, fmt.Errorf("fundamentals data requires Yahoo Finance paid subscription: %w", err)
		}
		return nil, err
	}

	// Extract fundamentals
	fundamentals, err := fundResp.GetFundamentals()
	if err != nil {
		return nil, err
	}

	// Normalize then convert to the plain SDK surface.
	snap, err := model.NormalizeFundamentals(fundamentals, symbol, runID)
	if err != nil {
		return nil, err
	}
	return FromFundamentalsSnapshot(snap), nil
}

// FetchIntradayBars fetches intraday bars for a symbol (1m, 5m, 15m, 30m, 60m
// intervals) and returns the plain SDK BarBatch.
func (c *Client) FetchIntradayBars(ctx context.Context, symbol string, start, end time.Time, interval string, runID string) (*BarBatch, error) {
	// Fetch raw data
	barsResp, err := c.yahooClient.FetchIntradayBars(ctx, symbol, start, end, interval)
	if err != nil {
		return nil, err
	}

	// Extract bars and metadata
	bars, err := barsResp.GetBars()
	if err != nil {
		return nil, err
	}

	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	// Normalize then convert to the plain SDK surface.
	batch, err := model.NormalizeBars(bars, meta, runID)
	if err != nil {
		return nil, err
	}
	return FromBarBatch(batch), nil
}

// FetchWeeklyBars fetches weekly bars for a symbol and returns the plain SDK BarBatch.
func (c *Client) FetchWeeklyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*BarBatch, error) {
	barsResp, err := c.yahooClient.FetchWeeklyBars(ctx, symbol, start, end, adjusted)
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

	batch, err := model.NormalizeBars(bars, meta, runID)
	if err != nil {
		return nil, err
	}
	return FromBarBatch(batch), nil
}

// FetchMonthlyBars fetches monthly bars for a symbol and returns the plain SDK BarBatch.
func (c *Client) FetchMonthlyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*BarBatch, error) {
	barsResp, err := c.yahooClient.FetchMonthlyBars(ctx, symbol, start, end, adjusted)
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

	batch, err := model.NormalizeBars(bars, meta, runID)
	if err != nil {
		return nil, err
	}
	return FromBarBatch(batch), nil
}

// FetchCompanyInfo fetches basic company information from chart metadata
// and returns the plain SDK CompanyInfo.
func (c *Client) FetchCompanyInfo(ctx context.Context, symbol string, runID string) (*CompanyInfo, error) {
	// Use chart endpoint to get company info from metadata
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

	info, err := model.NormalizeCompanyInfo(meta, runID)
	if err != nil {
		return nil, err
	}
	return FromCompanyInfo(info), nil
}

// FetchMarketData fetches comprehensive market data (price, volume, 52-week
// range, etc.) and returns the plain SDK MarketData.
func (c *Client) FetchMarketData(ctx context.Context, symbol string, runID string) (*MarketData, error) {
	// Use chart endpoint to get comprehensive market data
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

	md, err := model.NormalizeMarketData(meta, runID)
	if err != nil {
		return nil, err
	}
	return FromMarketData(md), nil
}

// FetchIncomeStatement fetches the latest annual income statement from
// Yahoo's fundamentals-timeseries endpoint.
func (c *Client) FetchIncomeStatement(ctx context.Context, symbol string) (*FundamentalsSnapshot, error) {
	return c.fetchFinancialStatement(ctx, symbol, yahoo.IncomeStatement)
}

// FetchBalanceSheet fetches the latest annual balance sheet from Yahoo's
// fundamentals-timeseries endpoint.
func (c *Client) FetchBalanceSheet(ctx context.Context, symbol string) (*FundamentalsSnapshot, error) {
	return c.fetchFinancialStatement(ctx, symbol, yahoo.BalanceSheet)
}

// FetchCashFlowStatement fetches the latest annual cash-flow statement from
// Yahoo's fundamentals-timeseries endpoint.
func (c *Client) FetchCashFlowStatement(ctx context.Context, symbol string) (*FundamentalsSnapshot, error) {
	return c.fetchFinancialStatement(ctx, symbol, yahoo.CashFlow)
}

func (c *Client) fetchFinancialStatement(ctx context.Context, symbol string, kind yahoo.StatementKind) (*FundamentalsSnapshot, error) {
	snapshot, err := c.yahooClient.FetchFinancialStatement(ctx, symbol, kind)
	if err != nil {
		return nil, err
	}
	snapshot.MIC = c.inferMICForSymbol(ctx, symbol)
	return snapshot, nil
}

// FetchNews fetches the latest ticker news from Yahoo's JSON XHR endpoint.
func (c *Client) FetchNews(ctx context.Context, symbol string) ([]NewsItem, error) {
	return c.yahooClient.FetchNews(ctx, symbol)
}
