// Tests `ExtractMetadata` (chart-meta field decode + empty-result error) and `FetchMetadata` (asserts the chart call uses a < 7-day range so cached metadata is reused cheaply).
package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

func TestExtractMetadata(t *testing.T) {
	raw := []byte(`{"chart":{"result":[{
	  "meta":{"symbol":"AAPL","currency":"USD","exchangeName":"NMS",
	    "instrumentType":"EQUITY","timezone":"EST","gmtoffset":-18000,
	    "firstTradeDate":345479400,"regularMarketPrice":150.0}
	}],"error":null}}`)

	m, err := ExtractMetadata(raw)
	require.NoError(t, err)
	require.Equal(t, "AAPL", m.Symbol)
	require.Equal(t, "USD", m.Currency)
	require.Equal(t, "NMS", m.ExchangeName)
}

func TestExtractMetadata_EmptyResult(t *testing.T) {
	_, err := ExtractMetadata([]byte(`{"chart":{"result":[],"error":null}}`))
	require.Error(t, err)
}

func TestFetchMetadata_OneDayRange(t *testing.T) {
	var period1, period2 int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		p1, _ := strconv.ParseInt(q.Get("period1"), 10, 64)
		p2, _ := strconv.ParseInt(q.Get("period2"), 10, 64)
		period1, period2 = p1, p2
		_, _ = w.Write([]byte(`{"chart":{"result":[{"meta":{"symbol":"AAPL","currency":"USD"}}],"error":null}}`))
	}))
	defer srv.Close()

	c := NewClient(httpx.NewClient(httpx.DefaultConfig()), srv.URL)
	_, err := c.FetchMetadata(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Greater(t, period2, period1)
	require.Less(t, period2-period1, int64(7*24*3600), "FetchMetadata should use < 7 day range, got %d sec", period2-period1)
}
