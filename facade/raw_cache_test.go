package facade

import (
	"errors"
	"io/fs"
	"testing"

	cacheutil "github.com/bizshuk/yfin/utils/cache"
	"github.com/stretchr/testify/require"
)

func TestRawCacheReadLatestJSON(t *testing.T) {
	root := t.TempDir()
	store, err := cacheutil.NewJSONStore(root)
	require.NoError(t, err)
	require.NoError(t, store.Write("yahoo/news/AAPL.2026-07-20.json", map[string]int{"count": 1}))
	require.NoError(t, store.Write("yahoo/news/AAPL.2026-07-28.json", map[string]int{"count": 2}))
	require.NoError(t, store.Write("yahoo/news/MSFT.2026-07-29.json", map[string]int{"count": 99}))

	cache, err := NewRawCache(root)
	require.NoError(t, err)
	var got map[string]int
	asOf, err := cache.ReadLatestJSON("yahoo/news", "AAPL.", &got)
	require.NoError(t, err)
	require.Equal(t, "2026-07-28", asOf)
	require.Equal(t, map[string]int{"count": 2}, got)
}

func TestRawCacheReadLatestJSONMissing(t *testing.T) {
	cache, err := NewRawCache(t.TempDir())
	require.NoError(t, err)

	var got any
	_, err = cache.ReadLatestJSON("yahoo/news", "AAPL.", &got)
	require.Error(t, err)
	require.True(t, errors.Is(err, fs.ErrNotExist))
}
