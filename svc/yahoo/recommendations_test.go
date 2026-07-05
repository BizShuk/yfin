// Tests `DecodeRecommendationTrend`: decodes `strongBuy/buy/hold/sell/strongSell` rows keyed by period, and rejects an empty `quoteSummary.result`.
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeRecommendationTrend(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "recommendationTrend":{"trend":[
	    {"period":"0m","strongBuy":5,"buy":10,"hold":3,"sell":1,"strongSell":0}]}
	}],"error":null}}`)

	rows, err := DecodeRecommendationTrend(raw)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, 5, rows[0].StrongBuy)
	require.Equal(t, "0m", rows[0].Period)
}

func TestDecodeRecommendationTrend_EmptyResult(t *testing.T) {
	_, err := DecodeRecommendationTrend([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
	require.Error(t, err)
}
