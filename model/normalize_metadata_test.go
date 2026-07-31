package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeMetadataUsesCanonicalRuntimeIdentity(t *testing.T) {
	t.Parallel()

	bars, err := NormalizeBars(nil, &ChartMeta{Symbol: "AAPL"}, "bars-run")
	require.NoError(t, err)
	assertRuntimeIdentity(t, bars.Meta, "bars-run")

	quote, err := NormalizeQuote(RawQuote{
		Symbol:   "AAPL",
		Currency: "USD",
		Exchange: "NMS",
	}, "quote-run")
	require.NoError(t, err)
	assertRuntimeIdentity(t, quote.Meta, "quote-run")

	market, err := NormalizeMarketData(&ChartMeta{Symbol: "AAPL"}, "market-run")
	require.NoError(t, err)
	assertRuntimeIdentity(t, market.Meta, "market-run")
	require.Equal(t, market.EventTime, market.IngestTime)

	company, err := NormalizeCompanyInfo(&ChartMeta{Symbol: "AAPL"}, "company-run")
	require.NoError(t, err)
	assertRuntimeIdentity(t, company.Meta, "company-run")
	require.Equal(t, company.EventTime, company.IngestTime)
}

func TestNormalizeFundamentalsDerivesProviderMetadata(t *testing.T) {
	t.Parallel()

	firstEnd := time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)
	latestEnd := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	revenue := int64(100)
	assets := int64(200)

	snapshot, err := NormalizeFundamentals(&Fundamentals{
		IncomeStatements: []IncomeStatement{{
			EndDate:      DateValue{Raw: firstEnd.Unix()},
			TotalRevenue: &Value{Raw: &revenue},
		}},
		BalanceSheets: []BalanceSheet{{
			EndDate:     DateValue{Raw: latestEnd.Unix()},
			TotalAssets: &Value{Raw: &assets},
		}},
	}, "AAPL", "fundamentals-run")
	require.NoError(t, err)

	assertRuntimeIdentity(t, snapshot.Meta, "fundamentals-run")
	require.Equal(t, "yahoo-finance", snapshot.Source)
	require.Equal(t, latestEnd, snapshot.AsOf)
	for _, line := range snapshot.Lines {
		require.Empty(t, line.CurrencyCode, "missing upstream currency must remain absent")
	}
}

func TestNormalizeFundamentalsRejectsMissingPeriodDate(t *testing.T) {
	t.Parallel()

	revenue := int64(100)
	_, err := NormalizeFundamentals(&Fundamentals{
		IncomeStatements: []IncomeStatement{{
			TotalRevenue: &Value{Raw: &revenue},
		}},
	}, "AAPL", "fundamentals-run")
	require.ErrorContains(t, err, "no valid fundamentals lines")
}

func assertRuntimeIdentity(t *testing.T, meta Meta, runID string) {
	t.Helper()
	require.Equal(t, runID, meta.RunID)
	require.Equal(t, "yfin", meta.Source)
	require.Equal(t, "local", meta.Producer)
	require.NotEmpty(t, meta.SchemaVersion)
}
