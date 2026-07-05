// Yahoo Finance API client (bars, quotes, fundamentals via chart/quoteSummary).

package yahoo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bizshuk/yfinance-go/utils/httpx"
)

// Client provides methods to fetch and normalize Yahoo Finance data
type Client struct {
	httpClient *httpx.Client
	baseURL    string
	crumb      *CrumbManager
}

// NewClientWithAuth creates a Client wired with a CrumbManager for authenticated endpoints.
func NewClientWithAuth(httpClient *httpx.Client, baseURL string, cm *CrumbManager) *Client {
	c := NewClient(httpClient, baseURL)
	c.crumb = cm
	return c
}

// NewClient creates a new Yahoo Finance client
func NewClient(httpClient *httpx.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://query1.finance.yahoo.com"
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// FetchDailyBars fetches daily bars for a symbol
func (c *Client) FetchDailyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool) (*BarsResponse, error) {
	// Build URL for daily bars
	u, err := c.buildBarsURL(symbol, start, end, adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to build bars URL: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bars: %w", err)
	}
	defer resp.Body.Close()

	// Decode response
	barsResp, err := DecodeBarsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bars response: %w", err)
	}

	// Validate response has data
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	return barsResp, nil
}

// FetchQuote fetches a quote for a symbol using the chart endpoint (since v7 quote endpoint is restricted)
func (c *Client) FetchQuote(ctx context.Context, symbol string) (*QuoteResponse, error) {
	// Use chart endpoint to get quote data from metadata
	// This is more reliable since v7 quote endpoint returns 401
	end := time.Now()
	start := end.AddDate(0, 0, -7) // Get 7 days of data to ensure we have recent data

	barsResp, err := c.FetchDailyBars(ctx, symbol, start, end, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote via chart endpoint: %w", err)
	}

	// Convert chart metadata to quote response
	quoteResp, err := c.convertChartToQuote(barsResp, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to convert chart to quote: %w", err)
	}

	return quoteResp, nil
}

// FundamentalsModules are the quoteSummary modules required for quarterly financials.
var FundamentalsModules = []string{
	"incomeStatementHistoryQuarterly",
	"balanceSheetHistoryQuarterly",
	"cashflowStatementHistoryQuarterly",
}

// FetchFundamentalsQuarterly fetches quarterly fundamentals for a symbol using crumb auth.
func (c *Client) FetchFundamentalsQuarterly(ctx context.Context, symbol string) (*FundamentalsResponse, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, FundamentalsModules)
	if err != nil {
		return nil, err
	}
	return DecodeFundamentalsResponse(raw)
}

// fetchChartRaw returns the raw bytes of the chart response for the given lookback window in days.
func (c *Client) fetchChartRaw(ctx context.Context, symbol string, daysBack int) ([]byte, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -daysBack)
	u, err := c.buildBarsURL(symbol, start, end, true)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// buildBarsURL builds the URL for fetching daily bars
func (c *Client) buildBarsURL(symbol string, start, end time.Time, adjusted bool) (string, error) {
	u, err := url.Parse(c.baseURL + "/v8/finance/chart/" + symbol)
	if err != nil {
		return "", err
	}

	// Add query parameters
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", "1d")
	params.Set("includePrePost", "false")
	params.Set("events", "div,split")

	u.RawQuery = params.Encode()
	return u.String(), nil
}

// convertChartToQuote converts chart metadata to a quote response
func (c *Client) convertChartToQuote(barsResp *BarsResponse, symbol string) (*QuoteResponse, error) {
	if len(barsResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart results found")
	}

	result := barsResp.Chart.Result[0]
	meta := result.Meta

	// Get the most recent close price from the bars data
	var regularMarketPrice *float64
	var regularMarketDayHigh, regularMarketDayLow *float64

	if len(result.Indicators.Quote) > 0 {
		quote := result.Indicators.Quote[0]

		if len(quote.Close) > 0 {
			// Get the last close price
			lastClose := quote.Close[len(quote.Close)-1]
			if lastClose != nil {
				regularMarketPrice = lastClose
			}
		}

		// Get the last high and low
		if len(quote.High) > 0 {
			lastHigh := quote.High[len(quote.High)-1]
			if lastHigh != nil {
				regularMarketDayHigh = lastHigh
			}
		}

		if len(quote.Low) > 0 {
			lastLow := quote.Low[len(quote.Low)-1]
			if lastLow != nil {
				regularMarketDayLow = lastLow
			}
		}
	}

	// Create a quote result from chart metadata
	quoteResult := QuoteResult{
		Symbol:                meta.Symbol,
		Currency:              meta.Currency,
		Exchange:              meta.ExchangeName,
		FullExchangeName:      meta.FullExchangeName,
		ShortName:             meta.ShortName,
		LongName:              meta.LongName,
		RegularMarketPrice:    regularMarketPrice,
		RegularMarketDayHigh:  regularMarketDayHigh,
		RegularMarketDayLow:   regularMarketDayLow,
		RegularMarketVolume:   meta.RegularMarketVolume,
		RegularMarketTime:     &meta.RegularMarketTime,
		ExchangeTimezoneName:  meta.ExchangeTimezoneName,
		GmtOffsetMilliseconds: meta.GmtOffset,
		MarketState:           "REGULAR", // Default to regular market
		QuoteType:             "EQUITY",  // Default to equity
		Language:              "en-US",   // Default language
		Region:                "US",      // Default region
	}

	// Create quote response
	quoteResponse := &QuoteResponse{
		QuoteResponse: QuoteResponseData{
			Result: []QuoteResult{quoteResult},
			Error:  nil,
		},
	}

	return quoteResponse, nil
}

// FetchIntradayBars fetches intraday bars for a symbol
func (c *Client) FetchIntradayBars(ctx context.Context, symbol string, start, end time.Time, interval string) (*BarsResponse, error) {
	// Build URL for intraday bars
	u, err := c.buildIntradayBarsURL(symbol, start, end, interval)
	if err != nil {
		return nil, fmt.Errorf("failed to build intraday bars URL: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch intraday bars: %w", err)
	}
	defer resp.Body.Close()

	// Decode response
	barsResp, err := DecodeBarsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode intraday bars response: %w", err)
	}

	// Validate response has data
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	return barsResp, nil
}

// FetchWeeklyBars fetches weekly bars for a symbol
func (c *Client) FetchWeeklyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool) (*BarsResponse, error) {
	// Build URL for weekly bars
	u, err := c.buildWeeklyBarsURL(symbol, start, end, adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to build weekly bars URL: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weekly bars: %w", err)
	}
	defer resp.Body.Close()

	// Decode response
	barsResp, err := DecodeBarsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode weekly bars response: %w", err)
	}

	// Validate response has data
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	return barsResp, nil
}

// FetchMonthlyBars fetches monthly bars for a symbol
func (c *Client) FetchMonthlyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool) (*BarsResponse, error) {
	// Build URL for monthly bars
	u, err := c.buildMonthlyBarsURL(symbol, start, end, adjusted)
	if err != nil {
		return nil, fmt.Errorf("failed to build monthly bars URL: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monthly bars: %w", err)
	}
	defer resp.Body.Close()

	// Decode response
	barsResp, err := DecodeBarsResponseFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode monthly bars response: %w", err)
	}

	// Validate response has data
	meta := barsResp.GetMetadata()
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	return barsResp, nil
}

// buildIntradayBarsURL builds the URL for fetching intraday bars
func (c *Client) buildIntradayBarsURL(symbol string, start, end time.Time, interval string) (string, error) {
	u, err := url.Parse(c.baseURL + "/v8/finance/chart/" + symbol)
	if err != nil {
		return "", err
	}

	// Add query parameters
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", interval)
	params.Set("includePrePost", "false")
	params.Set("events", "div,split")

	u.RawQuery = params.Encode()
	return u.String(), nil
}

// buildWeeklyBarsURL builds the URL for fetching weekly bars
func (c *Client) buildWeeklyBarsURL(symbol string, start, end time.Time, adjusted bool) (string, error) {
	u, err := url.Parse(c.baseURL + "/v8/finance/chart/" + symbol)
	if err != nil {
		return "", err
	}

	// Add query parameters
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", "1wk")
	params.Set("includePrePost", "false")
	params.Set("events", "div,split")

	u.RawQuery = params.Encode()
	return u.String(), nil
}

// buildMonthlyBarsURL builds the URL for fetching monthly bars
func (c *Client) buildMonthlyBarsURL(symbol string, start, end time.Time, adjusted bool) (string, error) {
	u, err := url.Parse(c.baseURL + "/v8/finance/chart/" + symbol)
	if err != nil {
		return "", err
	}

	// Add query parameters
	params := url.Values{}
	params.Set("period1", strconv.FormatInt(start.Unix(), 10))
	params.Set("period2", strconv.FormatInt(end.Unix(), 10))
	params.Set("interval", "1mo")
	params.Set("includePrePost", "false")
	params.Set("events", "div,split")

	u.RawQuery = params.Encode()
	return u.String(), nil
}
