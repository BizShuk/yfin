// Tests the annual fundamentals-timeseries request contract and conversion to
// the stable model.FundamentalsSnapshot surface.
package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

func TestDecodeFinancialStatementSelectsLatestAnnualPoints(t *testing.T) {
	raw := []byte(`{
      "timeseries": {
        "result": [
          {
            "meta": {"type": ["annualTotalRevenue"]},
            "annualTotalRevenue": [
              {"asOfDate":"2023-09-30","periodType":"12M","currencyCode":"USD","reportedValue":{"raw":100}},
              {"asOfDate":"2024-09-30","periodType":"12M","currencyCode":"USD","reportedValue":{"raw":120}}
            ]
          },
          {
            "meta": {"type": ["annualNetIncomeCommonStockholders"]},
            "annualNetIncomeCommonStockholders": [
              {"asOfDate":"2024-09-30","periodType":"12M","currencyCode":"USD","reportedValue":{"raw":25}}
            ]
          },
          {
            "meta": {"type": ["annualBasicEPS"]},
            "annualBasicEPS": [
              {"asOfDate":"2024-09-30","periodType":"12M","currencyCode":"USD","reportedValue":{"raw":0}}
            ]
          },
          {
            "meta": {"type": ["annualDilutedEPS"]},
            "annualDilutedEPS": [
              {"asOfDate":"2024-09-30","periodType":"12M","currencyCode":"USD","reportedValue":{}}
            ]
          }
        ],
        "error": null
      }
    }`)

	snapshot, err := decodeFinancialStatement(raw, "AAPL", IncomeStatement)
	require.NoError(t, err)
	require.Equal(t, "AAPL", snapshot.Symbol)
	require.Equal(t, "yahoo/fundamentals-timeseries/income", snapshot.Source)
	require.Equal(t, time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC), snapshot.AsOf)
	require.Len(t, snapshot.Lines, 3)

	require.Equal(t, "total_revenue", snapshot.Lines[0].Key)
	require.Equal(t, 120.0, snapshot.Lines[0].Value)
	require.Equal(t, "USD", snapshot.Lines[0].CurrencyCode)
	require.Equal(t, time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), snapshot.Lines[0].PeriodStart)
	require.Equal(t, snapshot.AsOf, snapshot.Lines[0].PeriodEnd)
	require.Equal(t, "net_income", snapshot.Lines[1].Key)
	require.Equal(t, 25.0, snapshot.Lines[1].Value)
	require.Equal(t, "eps_basic", snapshot.Lines[2].Key)
	require.Zero(t, snapshot.Lines[2].Value, "reported zero is data, not a missing value")
}

func TestDecodeFinancialStatementRejectsEmptyMappedData(t *testing.T) {
	_, err := decodeFinancialStatement([]byte(`{"timeseries":{"result":[],"error":null}}`), "AAPL", BalanceSheet)
	require.ErrorContains(t, err, "no annual balance data")
}

func TestFetchFinancialStatementRequestContract(t *testing.T) {
	var receivedTypes []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/ws/fundamentals-timeseries/v1/finance/timeseries/AAPL", r.URL.Path)
		require.Equal(t, "AAPL", r.URL.Query().Get("symbol"))
		require.Equal(t, "1483142400", r.URL.Query().Get("period1"))
		require.NotEmpty(t, r.URL.Query().Get("period2"))
		receivedTypes = strings.Split(r.URL.Query().Get("type"), ",")
		_, _ = w.Write([]byte(`{"timeseries":{"result":[{"meta":{"type":["annualTotalRevenue"]},"annualTotalRevenue":[{"asOfDate":"2024-09-30","currencyCode":"USD","reportedValue":{"raw":120}}]}],"error":null}}`))
	}))
	defer server.Close()

	client := NewClient(httpx.NewClient(httpx.DefaultConfig()), "")
	client.timeseriesBaseURL = server.URL
	snapshot, err := client.FetchFinancialStatement(context.Background(), "AAPL", IncomeStatement)
	require.NoError(t, err)
	require.Len(t, snapshot.Lines, 1)
	require.Len(t, receivedTypes, 21)
	require.Equal(t, "annualTotalRevenue", receivedTypes[0])
	require.Contains(t, receivedTypes, "annualNetIncomeCommonStockholders")
	require.Contains(t, receivedTypes, "annualNormalizedEBITDA")
}

func TestFetchFinancialStatementRejectsUnsupportedKind(t *testing.T) {
	client := NewClient(httpx.NewClient(httpx.DefaultConfig()), "")
	_, err := client.FetchFinancialStatement(context.Background(), "AAPL", StatementKind("unknown"))
	require.ErrorContains(t, err, "unsupported statement kind")
}
