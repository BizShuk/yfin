package yahoo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

func TestYahooAuthCircuitDoesNotBlockChart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/v1/test/getcrumb":
			w.WriteHeader(http.StatusTooManyRequests)
		case strings.HasPrefix(r.URL.Path, "/v8/finance/chart/AAPL"):
			_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"symbol":"AAPL","currency":"USD"}}],"error":null}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := httpx.DefaultConfig()
	config.MaxAttempts = 1
	config.FailureThreshold = 1
	config.QPS = 100
	config.Burst = 10
	transport := httpx.NewClient(config)
	client := NewClientWithAuth(transport, server.URL,
		NewCrumbManager(transport, server.URL, server.URL))

	_, err := client.FetchQuoteSummary(context.Background(), "AAPL", []string{"assetProfile"})
	require.Error(t, err)
	metadata, err := client.FetchMetadata(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Equal(t, "AAPL", metadata.Symbol)
	_, err = client.FetchQuoteSummary(context.Background(), "AAPL", []string{"assetProfile"})
	require.True(t, errors.Is(err, httpx.ErrCircuitOpen), "auth family must remain open: %v", err)
}

func TestYahooRequestSitesDeclareCircuitGroups(t *testing.T) {
	expected := map[string]map[string]int{
		"auth.go":           {"circuitGroupAuth": 2},
		"quotesummary.go":   {"circuitGroupAuth": 1},
		"client.go":         {"circuitGroupChart": 5},
		"timeseries.go":     {"circuitGroupTimeseries": 1},
		"options.go":        {"circuitGroupOptions": 1},
		"news.go":           {"circuitGroupNews": 1},
		"earnings_dates.go": {"circuitGroupWeb": 1},
	}
	for file, groups := range expected {
		source, err := os.ReadFile(file)
		require.NoError(t, err)
		for group, count := range groups {
			require.GreaterOrEqual(t, strings.Count(string(source), group), count,
				"%s must declare %s", file, group)
		}
	}
}
