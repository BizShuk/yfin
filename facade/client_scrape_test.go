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
