// client_scrape_profile.go — complete Yahoo profile scrape surface.
package facade

import (
	"context"
	"fmt"

	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/svc/scrape"
)

// ScrapeProfileData fetches and parses the complete Yahoo company profile.
func (c *Client) ScrapeProfileData(
	ctx context.Context,
	symbol string,
	runID string,
) (*model.ComprehensiveProfileDTO, error) {
	body, _, err := c.scrapeClient.Fetch(ctx, BuildScrapeURL(symbol, "profile"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}

	dto, err := scrape.ParseComprehensiveProfile(body, symbol, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}
	return dto, nil
}
