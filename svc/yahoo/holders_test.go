package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeHolders_ParsesBreakdownAndInstitutions(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "majorHoldersBreakdown":{"insidersPercentHeld":{"raw":0.012,"fmt":"1.20%"},
	    "institutionsPercentHeld":{"raw":0.74,"fmt":"74.00%"}},
	  "institutionOwnership":{"ownershipList":[
	    {"organization":"Vanguard","pctHeld":{"raw":0.08,"fmt":"8.00%"},
	     "position":{"raw":1000000},"value":{"raw":250000000}}]}
	}],"error":null}}`)

	d, err := DecodeHolders(raw)
	require.NoError(t, err)
	require.NotNil(t, d.MajorBreakdown.InstitutionsPercentHeld.Raw)
	require.Equal(t, 0.74, *d.MajorBreakdown.InstitutionsPercentHeld.Raw)
	require.Len(t, d.InstitutionOwnership, 1)
	require.Equal(t, "Vanguard", d.InstitutionOwnership[0].Organization)
}

func TestDecodeHolders_EmptyResult(t *testing.T) {
	_, err := DecodeHolders([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
	require.Error(t, err)
}

func TestDecodeHolders_ParsesMajorDirectHolders(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "majorDirectHolders":{"holders":[
	    {"organization":"BlackRock","positionDirect":{"raw":5000000},
	     "positionDirectDate":{"raw":1700000000}, "valueDirect":{"raw":800000000}}
	  ]}
	}],"error":null}}`)

	d, err := DecodeHolders(raw)
	require.NoError(t, err)
	require.Len(t, d.MajorDirectHolders, 1)
	require.Equal(t, "BlackRock", d.MajorDirectHolders[0].Organization)
	require.NotNil(t, d.MajorDirectHolders[0].PositionDirect.Raw)
	require.Equal(t, int64(5000000), *d.MajorDirectHolders[0].PositionDirect.Raw)
}
