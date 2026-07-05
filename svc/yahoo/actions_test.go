package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractActions(t *testing.T) {
	raw := []byte(`{"chart":{"result":[{
	  "meta":{"symbol":"AAPL"},
	  "events":{
	    "dividends":{"1700000000":{"amount":0.24,"date":1700000000}},
	    "splits":{"1600000000":{"numerator":4,"denominator":1,"splitRatio":"4:1","date":1600000000}}}
	}],"error":null}}`)

	acts, err := ExtractActions(raw)
	require.NoError(t, err)
	require.Len(t, acts.Dividends, 1)
	require.Equal(t, 0.24, acts.Dividends[0].Amount)
	require.Equal(t, int64(1700000000), acts.Dividends[0].Date)
	require.Len(t, acts.Splits, 1)
	require.Equal(t, "4:1", acts.Splits[0].SplitRatio)
}

func TestExtractActions_NoEvents(t *testing.T) {
	acts, err := ExtractActions([]byte(`{"chart":{"result":[{"meta":{"symbol":"X"}}],"error":null}}`))
	require.NoError(t, err)
	require.Empty(t, acts.Dividends)
	require.Empty(t, acts.Splits)
}
