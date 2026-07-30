package scrape

import (
	"testing"
	"time"

	"github.com/bizshuk/yfin/model"
	"github.com/stretchr/testify/require"
)

func TestValidateScrapeFlagsRejectsMultipleModes(t *testing.T) {
	tests := []struct {
		name string
		cfg  scrapeConfig
	}{
		{
			name: "check and preview JSON",
			cfg: scrapeConfig{
				Check:       true,
				PreviewJSON: true,
				Ticker:      "AAPL",
				Endpoint:    "profile",
				Endpoints:   "profile",
			},
		},
		{
			name: "check and preview news",
			cfg: scrapeConfig{
				Check:       true,
				PreviewNews: true,
				Ticker:      "AAPL",
				Endpoint:    "news",
			},
		},
		{
			name: "preview JSON and preview news",
			cfg: scrapeConfig{
				PreviewJSON: true,
				PreviewNews: true,
				Ticker:      "AAPL",
				Endpoints:   "news",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScrapeFlags(&tt.cfg)
			require.EqualError(t, err, "exactly one of --check, --preview-json, or --preview-news is required")
		})
	}
}

func TestFormatScrapeCheckHonorsPreviewFlag(t *testing.T) {
	meta := &model.FetchMeta{
		Host:      "finance.yahoo.com",
		URL:       "https://finance.yahoo.com/quote/AAPL/profile",
		Status:    200,
		Bytes:     17,
		Duration:  125 * time.Millisecond,
		Redirects: 1,
	}

	withoutPreview := formatScrapeCheck([]byte("<html>body</html>"), meta, false)
	require.Contains(t, withoutPreview, "SCRAPE CHECK")
	require.NotContains(t, withoutPreview, "CONTENT PREVIEW")

	withPreview := formatScrapeCheck([]byte("<html>body</html>"), meta, true)
	require.Contains(t, withPreview, "CONTENT PREVIEW: <html>body</html>")
}
