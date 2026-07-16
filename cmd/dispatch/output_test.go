package dispatch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJSONAtomicRejectsUnmarshalableValueWithoutTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "info", "AAPL.2026-06-23.json")

	err := writeJSONAtomic(path, func() {})
	require.Error(t, err)
	require.NoFileExists(t, path)
}

func TestWriteJSONAtomicWritesCompleteDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "info", "AAPL.2026-06-23.json")

	require.NoError(t, writeJSONAtomic(path, map[string]string{"symbol": "AAPL"}))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.JSONEq(t, `{"symbol":"AAPL"}`, string(data))
	require.Equal(t, byte('\n'), data[len(data)-1])
}
