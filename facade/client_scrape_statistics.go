// client_scrape_statistics.go — complete Yahoo key-statistics scrape surface.
package facade

import (
	"context"
	"fmt"

	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/svc/scrape"
)

// ScrapeKeyStatisticsData fetches and parses the complete Yahoo key-statistics
// surface without reducing it to FundamentalsSnapshot lines.
func (c *Client) ScrapeKeyStatisticsData(
	ctx context.Context,
	symbol string,
	runID string,
) (*model.ComprehensiveKeyStatisticsDTO, error) {
	body, _, err := c.scrapeClient.Fetch(ctx, BuildScrapeURL(symbol, "key-statistics"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key statistics: %w", err)
	}

	dto, err := scrape.ParseComprehensiveKeyStatistics(body, symbol, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse key statistics: %w", err)
	}
	return dto, nil
}
