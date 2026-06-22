package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseISINResponse(t *testing.T) {
	body := "0\tAAPL|US0378331005|Apple Inc.\n1\tMSFT|US5949181045|Microsoft"
	isin, err := parseISIN(body, "AAPL")
	require.NoError(t, err)
	require.Equal(t, "US0378331005", isin)
}

func TestParseISINResponse_NotFound(t *testing.T) {
	_, err := parseISIN("0\tFOO|US123|Foo", "AAPL")
	require.Error(t, err)
}
