package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadTickerList(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ticker_list.csv")
	require.NoError(t, os.WriteFile(p, []byte("market, ticker\nTPEx, 3081.TWO\nTWSE, 2330.TW\n"), 0o644))

	got, err := ReadTickerList(p)
	require.NoError(t, err)
	require.Equal(t, []string{"3081.TWO", "2330.TW"}, got)
}

func TestReadTickerList_SkipsBlankLines(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ticker_list.csv")
	require.NoError(t, os.WriteFile(p, []byte("market, ticker\n\nTWSE, 2330.TW\n\n"), 0o644))

	got, err := ReadTickerList(p)
	require.NoError(t, err)
	require.Equal(t, []string{"2330.TW"}, got)
}
