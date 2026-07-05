package facade

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bizshuk/yfin/svc/emit"
	"github.com/bizshuk/yfin/svc/norm"
	"github.com/bizshuk/yfin/svc/scrape"
	"github.com/bizshuk/yfin/svc/yahoo"
	"github.com/bizshuk/yfin/utils/httpx"
)

// Client provides a high-level interface for fetching Yahoo Finance data
type Client struct {
	yahooClient  *yahoo.Client
	scrapeClient scrape.Client
	micCache     map[string]string // Cache for MIC inference to avoid repeated API calls
	micCacheMu   sync.RWMutex      // Mutex for MIC cache
}

// NewClient creates a new Yahoo Finance client with default configuration
func NewClient() *Client {
	config := httpx.DefaultConfig()
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, "")
	scrapeClient := scrape.NewClient(scrape.DefaultConfig(), httpClient)

	return &Client{
		yahooClient:  yahooClient,
		scrapeClient: scrapeClient,
		micCache:     make(map[string]string),
	}
}

// NewClientWithConfig creates a new Yahoo Finance client with custom configuration
func NewClientWithConfig(config *httpx.Config) *Client {
	httpClient := httpx.NewClient(config)
	yahooClient := yahoo.NewClient(httpClient, config.BaseURL)
	scrapeClient := scrape.NewClient(scrape.DefaultConfig(), httpClient)

	return &Client{
		yahooClient:  yahooClient,
		scrapeClient: scrapeClient,
		micCache:     make(map[string]string),
	}
}

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
	batch, err := norm.NormalizeBars(bars, meta, runID)
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
	q, err := norm.NormalizeQuote(quotes[0], runID)
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
	snap, err := norm.NormalizeFundamentals(fundamentals, symbol, runID)
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
	batch, err := norm.NormalizeBars(bars, meta, runID)
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

	batch, err := norm.NormalizeBars(bars, meta, runID)
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

	batch, err := norm.NormalizeBars(bars, meta, runID)
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

	info, err := norm.NormalizeCompanyInfo(meta, runID)
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

	md, err := norm.NormalizeMarketData(meta, runID)
	if err != nil {
		return nil, err
	}
	return FromMarketData(md), nil
}

// Scraping Functions - Return AMPY-PROTO Data

// inferMICForSymbol attempts to infer the MIC code for a symbol by fetching company info
// Uses caching to avoid repeated API calls for the same symbol
func (c *Client) inferMICForSymbol(ctx context.Context, symbol string) string {
	// Check cache first
	c.micCacheMu.RLock()
	if mic, found := c.micCache[symbol]; found {
		c.micCacheMu.RUnlock()
		return mic
	}
	c.micCacheMu.RUnlock()

	// Cache miss - fetch company info
	companyInfo, err := c.FetchCompanyInfo(ctx, symbol, "mic-inference")
	if err != nil {
		// If we can't fetch company info, cache empty string to avoid repeated failures
		c.micCacheMu.Lock()
		c.micCache[symbol] = ""
		c.micCacheMu.Unlock()
		return ""
	}

	mic := norm.InferMIC(companyInfo.Exchange, companyInfo.FullExchangeName)

	// Cache the result
	c.micCacheMu.Lock()
	c.micCache[symbol] = mic
	c.micCacheMu.Unlock()

	return mic
}

// ScrapeFinancials fetches financials data and returns the plain SDK
// FundamentalsSnapshot. The internal scrape → emit → proto pipeline still
// runs; only the last hop (proto → SDK struct) is new.
func (c *Client) ScrapeFinancials(ctx context.Context, symbol string, runID string) (*FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/financials", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch financials: %w", err)
	}

	mic := c.inferMICForSymbol(ctx, symbol)

	dto, err := scrape.ParseComprehensiveFinancials(body, symbol, mic)
	if err != nil {
		return nil, fmt.Errorf("failed to parse financials: %w", err)
	}

	snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map financials: %w", err)
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no financials data found")
	}

	return fromProtoFundamentals(snapshots[0]), nil
}

// ScrapeBalanceSheet fetches balance sheet data and returns the plain SDK
// FundamentalsSnapshot.
func (c *Client) ScrapeBalanceSheet(ctx context.Context, symbol string, runID string) (*FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/balance-sheet", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balance sheet: %w", err)
	}

	mic := c.inferMICForSymbol(ctx, symbol)

	dto, err := scrape.ParseComprehensiveFinancials(body, symbol, mic)
	if err != nil {
		return nil, fmt.Errorf("failed to parse balance sheet: %w", err)
	}

	snap, err := emit.MapBalanceSheetDTO(dto, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map balance sheet: %w", err)
	}
	return fromProtoFundamentals(snap), nil
}

// ScrapeCashFlow fetches cash flow data and returns the plain SDK FundamentalsSnapshot.
func (c *Client) ScrapeCashFlow(ctx context.Context, symbol string, runID string) (*FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/cash-flow", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cash flow: %w", err)
	}

	mic := c.inferMICForSymbol(ctx, symbol)

	dto, err := scrape.ParseComprehensiveFinancials(body, symbol, mic)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cash flow: %w", err)
	}

	snap, err := emit.MapCashFlowDTO(dto, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map cash flow: %w", err)
	}
	return fromProtoFundamentals(snap), nil
}

// ScrapeKeyStatistics fetches key statistics data and returns the plain SDK
// FundamentalsSnapshot.
func (c *Client) ScrapeKeyStatistics(ctx context.Context, symbol string, runID string) (*FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/key-statistics", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key statistics: %w", err)
	}

	mic := c.inferMICForSymbol(ctx, symbol)

	dto, err := scrape.ParseComprehensiveKeyStatistics(body, symbol, mic)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key statistics: %w", err)
	}

	snap, err := emit.MapKeyStatisticsDTO(dto, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map key statistics: %w", err)
	}
	return fromProtoFundamentals(snap), nil
}

// ScrapeAnalysis fetches analysis data and returns the plain SDK FundamentalsSnapshot.
func (c *Client) ScrapeAnalysis(ctx context.Context, symbol string, runID string) (*FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/analysis", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analysis: %w", err)
	}

	mic := c.inferMICForSymbol(ctx, symbol)

	dto, err := scrape.ParseAnalysis(body, symbol, mic)
	if err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %w", err)
	}

	snap, err := emit.MapAnalysisDTO(dto, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map analysis: %w", err)
	}
	return fromProtoFundamentals(snap), nil
}

// ScrapeAnalystInsights fetches analyst insights data and returns the plain
// SDK FundamentalsSnapshot.
func (c *Client) ScrapeAnalystInsights(ctx context.Context, symbol string, runID string) (*FundamentalsSnapshot, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/analyst-insights", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analyst insights: %w", err)
	}

	mic := c.inferMICForSymbol(ctx, symbol)

	dto, err := scrape.ParseAnalystInsights(body, symbol, mic)
	if err != nil {
		return nil, fmt.Errorf("failed to parse analyst insights: %w", err)
	}

	snap, err := emit.MapAnalystInsightsDTO(dto, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map analyst insights: %w", err)
	}
	return fromProtoFundamentals(snap), nil
}

// ScrapeNews fetches news data and returns the plain SDK []NewsItem slice.
func (c *Client) ScrapeNews(ctx context.Context, symbol string, runID string) ([]NewsItem, error) {
	url := fmt.Sprintf("https://finance.yahoo.com/quote/%s/news", symbol)
	body, _, err := c.scrapeClient.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}

	articles, _, err := scrape.ParseNews(body, "https://finance.yahoo.com", time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to parse news: %w", err)
	}

	protoArticles, err := emit.MapNewsItems(articles, symbol, runID, "yfinance-go")
	if err != nil {
		return nil, fmt.Errorf("failed to map news: %w", err)
	}

	return fromProtoNews(protoArticles), nil
}

// ScrapeAllFundamentals fetches all fundamentals data and returns the plain
// SDK FundamentalsSnapshot slice. The internals fan out to each per-endpoint
// scrape helper which itself returns plain types — no proto leaks.
func (c *Client) ScrapeAllFundamentals(ctx context.Context, symbol string, runID string) ([]*FundamentalsSnapshot, error) {
	var snapshots []*FundamentalsSnapshot

	financials, err := c.ScrapeFinancials(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape financials: %w", err)
	}
	snapshots = append(snapshots, financials)

	balanceSheet, err := c.ScrapeBalanceSheet(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape balance sheet: %w", err)
	}
	snapshots = append(snapshots, balanceSheet)

	cashFlow, err := c.ScrapeCashFlow(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape cash flow: %w", err)
	}
	snapshots = append(snapshots, cashFlow)

	keyStats, err := c.ScrapeKeyStatistics(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape key statistics: %w", err)
	}
	snapshots = append(snapshots, keyStats)

	analysis, err := c.ScrapeAnalysis(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape analysis: %w", err)
	}
	snapshots = append(snapshots, analysis)

	analystInsights, err := c.ScrapeAnalystInsights(ctx, symbol, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape analyst insights: %w", err)
	}
	snapshots = append(snapshots, analystInsights)

	return snapshots, nil
}

// isAuthenticationError checks if an error indicates authentication is required
func isAuthenticationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized") || strings.Contains(errStr, "authentication")
}
