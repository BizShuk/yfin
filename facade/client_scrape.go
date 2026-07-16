// client_scrape.go — the HTML-scrape surface of facade.Client: the 8
// `Scrape*` methods (Financials, BalanceSheet, CashFlow, KeyStatistics,
// Analysis, AnalystInsights, News, AllFundamentals) that fetch Yahoo pages
// via `svc/scrape` and convert the DTOs straight to model types.
package facade

import (
	"context"
	"fmt"
	"time"

	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/svc/scrape"
)

// ScrapeFinancials fetches financials data and returns the plain SDK
// FundamentalsSnapshot. After the ampy-proto removal, the scrape DTO is
// converted directly to a model FundamentalsSnapshot (no proto hop).
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

	return model.ScrapeFinancialsToSnapshot(dto, mic), nil
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

	return model.ScrapeBalanceSheetToSnapshot(dto, mic), nil
}

// ScrapeCashFlow fetches cash flow data and returns the plain SDK
// FundamentalsSnapshot.
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

	return model.ScrapeCashFlowToSnapshot(dto, mic), nil
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

	return model.ScrapeKeyStatisticsToSnapshot(dto, mic), nil
}

// ScrapeAnalysis fetches analysis data and returns the plain SDK
// FundamentalsSnapshot.
func (c *Client) ScrapeAnalysis(ctx context.Context, symbol string, runID string) (*FundamentalsSnapshot, error) {
	dto, err := c.ScrapeAnalysisData(ctx, symbol, runID)
	if err != nil {
		return nil, err
	}
	return model.ScrapeAnalysisToSnapshot(dto, dto.Market), nil
}

// ScrapeAnalysisData fetches and parses the complete Yahoo analysis surface.
func (c *Client) ScrapeAnalysisData(ctx context.Context, symbol, runID string) (*model.ComprehensiveAnalysisDTO, error) {
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

	return dto, nil
}

// ScrapeAnalysisDimension fetches the shared analysis page and projects the
// command-specific yfinance-compatible dimension.
func (c *Client) ScrapeAnalysisDimension(ctx context.Context, command, symbol, runID string) (any, error) {
	dto, err := c.ScrapeAnalysisData(ctx, symbol, runID)
	if err != nil {
		return nil, err
	}
	return projectAnalysisDimension(command, dto)
}

func projectAnalysisDimension(command string, dto *model.ComprehensiveAnalysisDTO) (any, error) {
	if dto == nil {
		return nil, fmt.Errorf("analysis data is nil")
	}
	switch command {
	case "earnings-history":
		return dto.EarningsHistory, nil
	case "eps-trend":
		return dto.EPSTrend, nil
	case "eps-revisions":
		return dto.EPSRevisions, nil
	case "earnings-estimates":
		return dto.EarningsEstimate, nil
	case "revenue-estimates":
		return dto.RevenueEstimate, nil
	case "growth-estimates":
		return dto.GrowthEstimate, nil
	default:
		return nil, fmt.Errorf("unsupported analysis command %q", command)
	}
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

	return model.ScrapeAnalystInsightsToSnapshot(dto, mic), nil
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

	return model.ScrapeNewsToItems(articles), nil
}

// ScrapeAllFundamentals fetches all fundamentals data and returns the plain
// SDK FundamentalsSnapshot slice. Fans out to each per-endpoint scrape helper
// which itself returns plain types.
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
