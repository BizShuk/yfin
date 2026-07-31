package facade

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

type staticScrapeClient struct {
	body    []byte
	lastURL string
	err     error
}

func (client *staticScrapeClient) Fetch(_ context.Context, url string) ([]byte, *model.FetchMeta, error) {
	client.lastURL = url
	return client.body, &model.FetchMeta{}, client.err
}

func TestProjectAnalysisDimension(t *testing.T) {
	dto := &model.ComprehensiveAnalysisDTO{}
	dto.EarningsHistory.Currency = "history"
	dto.EPSTrend.Currency = "trend"
	dto.EPSRevisions.Currency = "revisions"
	dto.EarningsEstimate.Currency = "earnings"
	dto.RevenueEstimate.Currency = "revenue"
	growth := "growth"
	dto.GrowthEstimate.CurrentQtr = &growth

	tests := []struct {
		command string
		want    any
	}{
		{command: "earnings-history", want: dto.EarningsHistory},
		{command: "eps-trend", want: dto.EPSTrend},
		{command: "eps-revisions", want: dto.EPSRevisions},
		{command: "earnings-estimates", want: dto.EarningsEstimate},
		{command: "revenue-estimates", want: dto.RevenueEstimate},
		{command: "growth-estimates", want: dto.GrowthEstimate},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got, err := projectAnalysisDimension(tt.command, dto)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	_, err := projectAnalysisDimension("unsupported", dto)
	require.EqualError(t, err, `unsupported analysis command "unsupported"`)
}

func TestScrapeProfileDataFetchesAndParsesProfileSurface(t *testing.T) {
	const profileHTML = `<script type="application/json" data-test="assetProfile">` +
		`{"body":"{\"quoteSummary\":{\"result\":[{\"assetProfile\":{\"sector\":\"Technology\",\"fullTimeEmployees\":164000}}]}}"}` +
		`</script>`
	scraper := &staticScrapeClient{body: []byte(profileHTML)}
	client := &Client{scrapeClient: scraper}

	dto, err := client.ScrapeProfileData(context.Background(), "AAPL", "test-run")

	require.NoError(t, err)
	require.Equal(t, "https://finance.yahoo.com/quote/AAPL/profile", scraper.lastURL)
	require.Equal(t, "AAPL", dto.Symbol)
	require.Equal(t, "Technology", dto.Sector)
	require.NotNil(t, dto.FullTimeEmployees)
	require.EqualValues(t, 164000, *dto.FullTimeEmployees)
}

func TestScrapeFinancialsDataPreservesSourceCurrency(t *testing.T) {
	scraper := &staticScrapeClient{body: []byte(`<div>Currency in EUR</div>`)}
	client := &Client{scrapeClient: scraper}

	dto, err := client.ScrapeFinancialsData(context.Background(), "SAP", "test-run")

	require.NoError(t, err)
	require.Equal(t, "https://finance.yahoo.com/quote/SAP/financials", scraper.lastURL)
	require.Equal(t, "SAP", dto.Symbol)
	require.Equal(t, "EUR", dto.Currency)
}

func TestConfigureScrapeAppliesLoadedPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := NewClientWithConfig(httpx.DefaultConfig())
	cfg := &config.ScrapeConfig{
		HTTP:         httpx.DefaultConfig(),
		Enabled:      true,
		RobotsPolicy: "ignore",
		CacheTTLMs:   60_000,
	}
	require.NoError(t, client.ConfigureScrape(cfg))

	body, meta, err := client.scrapeClient.Fetch(context.Background(), server.URL+"/quote/AAPL/profile")
	require.NoError(t, err)
	require.Equal(t, []byte("ok"), body)
	require.Equal(t, "ignore", meta.RobotsPolicy)
}

func TestScrapeConfigFromConfigPreservesOwnedSettings(t *testing.T) {
	httpConfig := httpx.DefaultConfig()
	cfg := &config.ScrapeConfig{
		HTTP:         httpConfig,
		Enabled:      true,
		RobotsPolicy: "warn",
		CacheTTLMs:   42_000,
		Endpoints: config.ScrapeEndpointConfig{
			KeyStatistics: true,
			Financials:    false,
			Analysis:      true,
			Profile:       false,
			News:          true,
		},
	}

	got := scrapeConfigFromConfig(cfg)

	require.Same(t, httpConfig, got.HTTP)
	require.True(t, got.Enabled)
	require.Equal(t, "warn", got.RobotsPolicy)
	require.Equal(t, 42_000, got.CacheTTLMs)
	require.True(t, got.Endpoints.KeyStatistics)
	require.False(t, got.Endpoints.Financials)
	require.True(t, got.Endpoints.Analysis)
	require.False(t, got.Endpoints.Profile)
	require.True(t, got.Endpoints.News)
}
