package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScrapeFallbackEndpointOnlyMapsEquivalentSurface(t *testing.T) {
	endpoint, ok := scrapeFallbackEndpoint("fundamentals")
	require.True(t, ok)
	require.Equal(t, "financials", endpoint)

	for _, unsupported := range []string{"quote", "daily-bars", "profile", ""} {
		_, ok = scrapeFallbackEndpoint(unsupported)
		require.Falsef(t, ok, "%s must not claim an equivalent scrape fallback", unsupported)
	}
}

func TestShouldFallbackToScrapeMatchesErrorsCaseInsensitively(t *testing.T) {
	require.True(t, shouldFallbackToScrape(errors.New("HTTP 503 SERVICE UNAVAILABLE")))
	require.True(t, shouldFallbackToScrape(errors.New("Authentication required")))
	require.False(t, shouldFallbackToScrape(errors.New("invalid symbol")))
}

func TestWorkerReusesRunIDAcrossRequests(t *testing.T) {
	worker := NewWorker(0, nil, nil, discardLogger(), nil, "stable-run")

	first := worker.generateWorkRequest([]string{"AAPL"}, []string{"quote"})
	second := worker.generateWorkRequest([]string{"MSFT"}, []string{"daily-bars"})

	require.Equal(t, "stable-run", first.RunID)
	require.Equal(t, first.RunID, second.RunID)
}
