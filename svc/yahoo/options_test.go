package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeOptions(t *testing.T) {
	raw := []byte(`{"optionChain":{"result":[{
	  "expirationDates":[1701000000,1701600000],
	  "strikes":[100.0,110.0],
	  "options":[{"expirationDate":1701000000,
	    "calls":[{"strike":{"raw":100},"lastPrice":{"raw":5.2},"volume":{"raw":120}}],
	    "puts":[{"strike":{"raw":100},"lastPrice":{"raw":2.1}}]}]
	}],"error":null}}`)

	d, err := DecodeOptions(raw)
	require.NoError(t, err)
	require.Len(t, d.ExpirationDates, 2)
	require.Len(t, d.Options, 1)
	require.Len(t, d.Options[0].Calls, 1)
	require.NotNil(t, d.Options[0].Calls[0].LastPrice.Raw)
	require.Equal(t, 5.2, *d.Options[0].Calls[0].LastPrice.Raw)
}

func TestDecodeOptions_EmptyResult(t *testing.T) {
	_, err := DecodeOptions([]byte(`{"optionChain":{"result":[],"error":null}}`))
	require.Error(t, err)
}
