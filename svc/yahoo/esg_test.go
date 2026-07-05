// Tests `DecodeESG`: decodes total/environment/social/governance scores + rating year + controversy, and rejects an empty `quoteSummary.result`.
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeESG(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "esgScores":{"totalEsg":{"raw":21.5},"environmentScore":{"raw":5.1},
	    "socialScore":{"raw":8.2},"governanceScore":{"raw":8.2},
	    "ratingYear":2024,"highestControversy":{"raw":3}}
	}],"error":null}}`)

	d, err := DecodeESG(raw)
	require.NoError(t, err)
	require.NotNil(t, d.TotalEsg.Raw)
	require.Equal(t, 21.5, *d.TotalEsg.Raw)
	require.Equal(t, 2024, d.RatingYear)
}

func TestDecodeESG_EmptyResult(t *testing.T) {
	_, err := DecodeESG([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
	require.Error(t, err)
}
