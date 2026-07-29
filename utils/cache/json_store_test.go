package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONStoreRoundTripAndReplace(t *testing.T) {
	store, err := NewJSONStore(t.TempDir())
	require.NoError(t, err)

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	path := filepath.Join("yahoo", "bars", "AAPL", "request.json")

	require.NoError(t, store.Write(path, payload{Name: "first", Count: 1}))
	require.NoError(t, store.Write(path, payload{Name: "second", Count: 2}))

	var got payload
	require.NoError(t, store.Read(path, &got))
	require.Equal(t, payload{Name: "second", Count: 2}, got)

	info, err := os.Stat(filepath.Join(store.Root(), path))
	require.NoError(t, err)
	require.Equal(t, fs.FileMode(0o644), info.Mode().Perm())
}

func TestJSONStoreReadMissingPreservesNotExist(t *testing.T) {
	store, err := NewJSONStore(t.TempDir())
	require.NoError(t, err)

	var got map[string]any
	err = store.Read("missing.json", &got)
	require.Error(t, err)
	require.True(t, errors.Is(err, fs.ErrNotExist))
}

func TestJSONStoreListReturnsSortedJSONFiles(t *testing.T) {
	store, err := NewJSONStore(t.TempDir())
	require.NoError(t, err)

	require.NoError(t, store.Write("bars/AAPL/z.json", map[string]int{"n": 3}))
	require.NoError(t, store.Write("bars/AAPL/a.json", map[string]int{"n": 1}))
	require.NoError(t, os.WriteFile(filepath.Join(store.Root(), "bars", "AAPL", "ignore.txt"), []byte("x"), 0o644))

	got, err := store.List("bars/AAPL")
	require.NoError(t, err)
	require.Equal(t, []string{"bars/AAPL/a.json", "bars/AAPL/z.json"}, got)
}

func TestJSONStoreRejectsPathsOutsideRoot(t *testing.T) {
	store, err := NewJSONStore(t.TempDir())
	require.NoError(t, err)

	invalid := []string{"", ".", "..", "../outside.json", "nested/../../outside.json", "/tmp/outside.json"}
	for _, path := range invalid {
		t.Run(path, func(t *testing.T) {
			require.Error(t, store.Write(path, map[string]bool{"bad": true}))
			var got any
			require.Error(t, store.Read(path, &got))
			_, err := store.List(path)
			require.Error(t, err)
		})
	}
}
