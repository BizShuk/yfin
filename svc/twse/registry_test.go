// registry_test.go — `Registry` completeness check (23 entries across 5 boards) + per-endpoint metadata spot-checks (board/path/needs*) + `Dispatcher.Call` transport check. Capacity: 23-entry coverage + 6-case metadata table + 1 dispatcher test.
package twse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

// TestDispatcher_Call_DispatchesViaClient exercises Dispatcher.Call over a
// real (latency-free) httpx transport pointed at a local httptest server.
// It lives here rather than in cmd/twse so that the CLI package needs no
// svc/twse import.
func TestDispatcher_Call_DispatchesViaClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stat":  "OK",
			"title": "twse:MI_INDEX",
			"data":  [][]string{{"1", "2"}},
			"date":  "20260620",
		})
	}))
	defer srv.Close()

	hc := httpx.NewClient(&httpx.Config{
		Timeout:          5 * time.Second,
		MaxAttempts:      1,
		BackoffBaseMs:    1,
		MaxDelayMs:       10,
		QPS:              1000,
		Burst:            1000,
		CircuitWindow:    time.Second,
		FailureThreshold: 1000,
		ResetTimeout:     time.Second,
	})
	d := NewDispatcher(NewClientWithURL(hc, srv.URL))

	raw, err := d.Call(context.Background(), "MI_INDEX", "20260620", url.Values{})
	require.NoError(t, err)
	require.NotNil(t, raw)
}

func TestRegistry_CoversAllEndpoints(t *testing.T) {
	want := []string{
		// afterTrading
		"MI_INDEX", "STOCK_DAY", "BWIBBU_d", "MI_INDEX_PLUS", "MI_INDEX_ODD",
		"MI_5MINS", "TWTB4U",
		// marginTrading
		"MI_MARGN",
		// fund
		"T86", "MI_QFIIS", "BFI82U", "TWT38U", "TWT43U", "TWT44U",
		// block
		"BFIAUU", "BFIAUU_STOCK", "BFIMUU", "BFIAUU_YEAR",
		// statistics
		"FMTQIK", "STOCK_DAY_AVG", "FMSRFK", "BFIAMU", "MI_WEEK",
	}
	for _, name := range want {
		_, ok := Registry[name]
		require.Truef(t, ok, "endpoint %q missing from registry", name)
	}
	require.Len(t, Registry, len(want))
}

func TestRegistry_EndpointMetadataIsCorrect(t *testing.T) {
	// Spot-check a few endpoints to make sure metadata is consistent.
	cases := []struct {
		name       string
		board      string
		path       string
		needsStock bool
		needsMonth bool
	}{
		{"MI_INDEX", "afterTrading", "/afterTrading/MI_INDEX", false, false},
		{"STOCK_DAY", "afterTrading", "/afterTrading/STOCK_DAY", true, false},
		{"T86", "fund", "/fund/T86", false, false},
		{"BFIAUU_STOCK", "block", "/block/BFIAUU", true, false},
		{"FMTQIK", "statistics", "/exchangeReport/FMTQIK", false, true},
		{"MI_MARGN", "marginTrading", "/marginTrading/MI_MARGN", false, false},
	}
	for _, tc := range cases {
		ep, ok := Registry[tc.name]
		require.Truef(t, ok, "%s missing", tc.name)
		require.Equalf(t, tc.board, ep.Board, "%s board", tc.name)
		require.Equalf(t, tc.path, ep.Path, "%s path", tc.name)
		require.Equalf(t, tc.needsStock, ep.NeedsStock, "%s needsStock", tc.name)
		require.Equalf(t, tc.needsMonth, ep.NeedsMonth, "%s needsMonth", tc.name)
		require.NotEmptyf(t, ep.Description, "%s description", tc.name)
	}
}
