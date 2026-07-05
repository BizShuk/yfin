// registry_test.go — `Registry` completeness check (23 entries across 5 boards) + per-endpoint metadata spot-checks (board/path/needs*). Capacity: 23-entry coverage + 6-case metadata table.
package twse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
