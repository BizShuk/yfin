// Tests `DecodeUpgrades`: decodes analyst `upgradeDowngradeHistory` rows (firm, from/toGrade, action, epochGradeDate) and rejects an empty `quoteSummary.result`.
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeUpgrades(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "upgradeDowngradeHistory":{"history":[
	    {"epochGradeDate":1700000000,"firm":"Morgan Stanley",
	     "toGrade":"Overweight","fromGrade":"Equal-Weight","action":"up"}]}
	}],"error":null}}`)

	rows, err := DecodeUpgrades(raw)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "Morgan Stanley", rows[0].Firm)
	require.Equal(t, "up", rows[0].Action)
	require.Equal(t, int64(1700000000), rows[0].EpochGradeDate)
}

func TestDecodeUpgrades_EmptyResult(t *testing.T) {
	_, err := DecodeUpgrades([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
	require.Error(t, err)
}
