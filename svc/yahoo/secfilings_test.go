package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeSecFilings(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "secFilings":{"filings":[
	    {"date":"2024-01-15","type":"10-K","title":"Annual Report",
	     "edgarUrl":"https://www.sec.gov/...","epochDate":1705276800}]}
	}],"error":null}}`)

	rows, err := DecodeSecFilings(raw)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "10-K", rows[0].Type)
	require.Equal(t, int64(1705276800), rows[0].EpochDate)
}

func TestDecodeSecFilings_EmptyResult(t *testing.T) {
	_, err := DecodeSecFilings([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
	require.Error(t, err)
}
