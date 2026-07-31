// client_scrape_financials.go — complete Yahoo financials scrape surface.
package facade

import (
	"context"
	"fmt"

	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/svc/scrape"
)

// ScrapeFinancialsData fetches and parses the complete Yahoo financials DTO
// without reducing source-level metadata such as currency.
func (c *Client) ScrapeFinancialsData(
	ctx context.Context,
	symbol string,
	runID string,
) (*model.ComprehensiveFinancialsDTO, error) {
	body, _, err := c.scrapeClient.Fetch(ctx, BuildScrapeURL(symbol, "financials"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch financials: %w", err)
	}

	dto, err := scrape.ParseComprehensiveFinancials(body, symbol, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse financials: %w", err)
	}
	return dto, nil
}
