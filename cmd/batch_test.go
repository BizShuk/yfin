package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// stubRegistry overrides commandRegistry for the test. We register a single
// "info" command that returns a fixed map.
func stubRegistry(t *testing.T) {
	t.Helper()
	prev := commandRegistry
	commandRegistry = map[string]fetchFunc{
		"info": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
			return map[string]string{"symbol": s}, nil
		},
	}
	t.Cleanup(func() { commandRegistry = prev })
}

func TestRunBatchForTicker_WritesOutputAndRespectsCache(t *testing.T) {
	stubRegistry(t)
	root := t.TempDir()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)

	res := runBatchForTicker(context.Background(), nil, "AAPL",
		[]string{"info"}, false, root, now)
	require.Equal(t, "success", res.Commands["info"])

	out := filepath.Join(root, "info", "AAPL.2026-06-23.json")
	b, err := os.ReadFile(out)
	require.NoError(t, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(b, &got))
	require.Equal(t, "AAPL", got["symbol"])

	// monthly tier: same month → skip
	res2 := runBatchForTicker(context.Background(), nil, "AAPL",
		[]string{"info"}, false, root, now)
	require.Equal(t, "skipped", res2.Commands["info"])
}

func TestRunBatchForTicker_RecordsFailure(t *testing.T) {
	prev := commandRegistry
	commandRegistry = map[string]fetchFunc{
		"bad": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
			return nil, errBoom
		},
	}
	t.Cleanup(func() { commandRegistry = prev })

	root := t.TempDir()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)
	res := runBatchForTicker(context.Background(), nil, "AAPL",
		[]string{"bad"}, false, root, now)
	require.Equal(t, "failed", res.Commands["bad"])

	errPath := filepath.Join(root, "_failed", "AAPL.bad.err")
	_, err := os.Stat(errPath)
	require.NoError(t, err)
}

type errString string

func (e errString) Error() string { return string(e) }

var errBoom = errString("boom")
