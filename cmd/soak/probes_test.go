package main

import (
	"testing"

	"github.com/bizshuk/yfin/model"
	"github.com/stretchr/testify/require"
)

func TestCorrectnessProbesCompareProvidedRealSourceValues(t *testing.T) {
	marketCap := model.ScaledDecimal{Scaled: 105_000_000_000, Scale: 2}
	trailingPE := model.ScaledDecimal{Scaled: 2_600, Scale: 2}
	employees := int64(155_000)
	stats := &model.ComprehensiveKeyStatisticsDTO{}
	stats.Current.MarketCap = &marketCap
	stats.Current.TrailingPE = &trailingPE
	dataset := probeDataset{
		info: map[string]any{
			"marketCap":         1_000_000_000.0,
			"trailingPE":        25.5,
			"fullTimeEmployees": 150_000.0,
			"sector":            "Information Technology",
			"financialCurrency": "USD",
		},
		stats:      stats,
		financials: &model.ComprehensiveFinancialsDTO{Currency: "USD"},
		profile: &model.ComprehensiveProfileDTO{
			Sector:            "Technology",
			FullTimeEmployees: &employees,
		},
	}

	require.True(t, probeMarketCap("AAPL", dataset).Passed)
	require.True(t, probePERatio("AAPL", dataset).Passed)
	require.True(t, probeEmployeeCount("AAPL", dataset).Passed)
	require.True(t, probeSector("AAPL", dataset).Passed)
	require.True(t, probeCurrency("AAPL", dataset).Passed)
}

func TestCorrectnessProbeFailsWhenSourceValueIsUnavailable(t *testing.T) {
	stats := &model.ComprehensiveKeyStatisticsDTO{}
	result := probeMarketCap("AAPL", probeDataset{
		info:  map[string]any{},
		stats: stats,
	})

	require.False(t, result.Passed)
	require.Contains(t, result.Details, "unavailable")
}
