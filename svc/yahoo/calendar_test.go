package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeCalendar(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "calendarEvents":{
	    "earnings":{"earningsDate":[{"raw":1701000000}],
	      "earningsAverage":{"raw":2.1},"revenueAverage":{"raw":100000000}},
	    "exDividendDate":{"raw":1699000000},
	    "dividendDate":{"raw":1699500000}}
	}],"error":null}}`)

	d, err := DecodeCalendar(raw)
	require.NoError(t, err)
	require.Len(t, d.EarningsDates, 1)
	require.Equal(t, int64(1701000000), d.EarningsDates[0])
	require.NotNil(t, d.ExDividendDate.Raw)
	require.Equal(t, int64(1699000000), *d.ExDividendDate.Raw)
	require.NotNil(t, d.EarningsAverage.Raw)
}

func TestDecodeCalendar_EmptyResult(t *testing.T) {
	_, err := DecodeCalendar([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
	require.Error(t, err)
}
