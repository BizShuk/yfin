package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeInfo_MergesModules(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "assetProfile":{"sector":"Technology","industry":"Semiconductors",
	    "fullTimeEmployees":50000,"longBusinessSummary":"..."},
	  "summaryDetail":{"marketCap":{"raw":600000000000},
	    "trailingPE":{"raw":18.5},"dividendYield":{"raw":0.018}},
	  "quoteType":{"longName":"TSMC","symbol":"2330.TW","quoteType":"EQUITY"}
	}],"error":null}}`)

	info, err := DecodeInfo(raw)
	require.NoError(t, err)
	require.Equal(t, "Technology", info["sector"])
	require.Equal(t, "TSMC", info["longName"])
	require.InDelta(t, 6.0e11, info["marketCap"], 1)
	require.InDelta(t, 18.5, info["trailingPE"], 0.001)
}

func TestDecodeInfo_EmptyResult(t *testing.T) {
	_, err := DecodeInfo([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
	require.Error(t, err)
}
