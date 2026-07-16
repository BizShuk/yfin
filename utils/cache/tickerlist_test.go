// — Unit tests for `ReadTickerList`: basic CSV parsing and blank-line skipping. Capacity: 2 test funcs × 2 fixtures.

package cache

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadTickerList(t *testing.T) {
	got, err := ReadTickerList(strings.NewReader("market,ticker\nTPEx,3081.TWO\nTWSE,2330.TW\n"))
	require.NoError(t, err)
	require.Equal(t, []string{"3081.TWO", "2330.TW"}, got)
}

func TestReadTickerList_SkipsBlankLines(t *testing.T) {
	got, err := ReadTickerList(strings.NewReader("market, ticker\n\nTWSE, 2330.TW\n,\n"))
	require.NoError(t, err)
	require.Equal(t, []string{"2330.TW"}, got)
}
