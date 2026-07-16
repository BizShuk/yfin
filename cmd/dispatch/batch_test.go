// batch_test.go — tests for `runBatchForTicker`: stub registry verifies success writes JSON, cache hit on re-run yields `skipped`, and failure path writes `_failed/<ticker>.<cmd>.err`. Capacity: 2 test functions + 1 stub registry helper + 1 `errBoom` sentinel.
package dispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

func stubRegistry() map[string]fetchFunc {
	return map[string]fetchFunc{
		"info": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
			return map[string]string{"symbol": s}, nil
		},
	}
}

func TestRunBatchForTicker_WritesOutputAndRespectsCache(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)

	res := runBatchForTicker(context.Background(), nil, stubRegistry(), "AAPL",
		[]string{"info"}, false, root, now)
	require.Equal(t, "success", res.Commands["info"])

	out := filepath.Join(root, "info", "AAPL.2026-06-23.json")
	b, err := os.ReadFile(out)
	require.NoError(t, err)
	var got map[string]string
	require.NoError(t, json.Unmarshal(b, &got))
	require.Equal(t, "AAPL", got["symbol"])

	// monthly tier: same month → skip
	res2 := runBatchForTicker(context.Background(), nil, stubRegistry(), "AAPL",
		[]string{"info"}, false, root, now)
	require.Equal(t, "skipped", res2.Commands["info"])
}

func TestRunBatchForTicker_RecordsFailure(t *testing.T) {
	registry := map[string]fetchFunc{
		"bad": func(ctx context.Context, fc *FetchContext, s string) (any, error) {
			return nil, errBoom
		},
	}

	root := t.TempDir()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)
	res := runBatchForTicker(context.Background(), nil, registry, "AAPL",
		[]string{"bad"}, false, root, now)
	require.Equal(t, "failed", res.Commands["bad"])

	errPath := filepath.Join(root, "_failed", "AAPL.bad.err")
	_, err := os.Stat(errPath)
	require.NoError(t, err)
}

func TestRunBatchForTickerReportsWriteFailure(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "info"), []byte("not a directory"), 0o644))
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)

	res := runBatchForTicker(context.Background(), nil, stubRegistry(), "AAPL",
		[]string{"info"}, true, root, now)

	require.Equal(t, "failed", res.Commands["info"])
	require.Contains(t, res.Errors, "info")
	_, err := os.Stat(filepath.Join(root, "info", "AAPL.2026-06-23.json"))
	require.Error(t, err)
}

func TestRunBatchForTickerReportsErrorArtifactFailure(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "_failed"), []byte("not a directory"), 0o644))
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)
	registry := map[string]fetchFunc{
		"bad": func(context.Context, *FetchContext, string) (any, error) { return nil, errBoom },
	}

	res := runBatchForTicker(context.Background(), nil, registry, "AAPL",
		[]string{"bad"}, true, root, now)

	require.Equal(t, "failed", res.Commands["bad"])
	require.ErrorContains(t, res.Errors["bad"], "write error artifact")
}

func TestRunBatchForTickerClassifiesNotFoundStatus(t *testing.T) {
	for _, statusCode := range []int{404, 422} {
		t.Run(fmt.Sprintf("status_%d", statusCode), func(t *testing.T) {
			root := t.TempDir()
			registry := map[string]fetchFunc{
				"info": func(context.Context, *FetchContext, string) (any, error) {
					return nil, httpx.NewHTTPError(statusCode, "not found", nil)
				},
			}

			res := runBatchForTicker(context.Background(), nil, registry, "AAPL",
				[]string{"info"}, true, root, time.Now())

			require.Equal(t, "not_found", res.Commands["info"])
			require.Empty(t, res.Errors)
			require.NoFileExists(t, filepath.Join(root, "_failed", "AAPL.info.err"))
		})
	}
}

func TestRunBatchRejectsNonPositiveWorkers(t *testing.T) {
	err := runBatch(context.Background(), batchOptions{maxWorkers: 0}, batchDeps{})
	require.EqualError(t, err, "max-workers must be greater than zero")
}

func TestReadEmbeddedTickerList(t *testing.T) {
	tickers, err := readEmbeddedTickerList()
	require.NoError(t, err)
	require.Contains(t, tickers, "2330.TW")
}

func TestBatchCommandUsesInjectedClient(t *testing.T) {
	var called atomic.Bool
	registry := make(map[string]fetchFunc, len(commandOrder))
	for _, command := range commandOrder {
		registry[command] = func(_ context.Context, _ *FetchContext, _ string) (any, error) {
			return map[string]string{"command": command}, nil
		}
	}
	registry["info"] = func(_ context.Context, fc *FetchContext, _ string) (any, error) {
		require.NotNil(t, fc)
		require.NotNil(t, fc.Root)
		called.Store(true)
		return map[string]string{"ok": "true"}, nil
	}
	deps := batchDeps{
		newClient: func() (*facade.Client, error) { return facade.NewClient(), nil },
		dataDir:   t.TempDir,
		readTickers: func() ([]string, error) {
			return []string{"AAPL"}, nil
		},
		now:      func() time.Time { return time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC) },
		registry: registry,
	}
	command := newBatchCmd(deps)
	command.SetArgs([]string{"--ticker", "AAPL", "--max-workers", "1", "--force"})

	require.NoError(t, command.Execute())
	require.True(t, called.Load())
}

func TestRunBatchStopsDispatchAfterCancellation(t *testing.T) {
	var calls atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	deps := batchDeps{
		newClient: func() (*facade.Client, error) { return facade.NewClient(), nil },
		dataDir:   t.TempDir,
		readTickers: func() ([]string, error) {
			return []string{"AAPL", "MSFT"}, nil
		},
		now: time.Now,
		registry: map[string]fetchFunc{
			"info": func(_ context.Context, _ *FetchContext, _ string) (any, error) {
				calls.Add(1)
				return nil, nil
			},
		},
	}

	require.ErrorIs(t, runBatch(ctx, batchOptions{maxWorkers: 1}, deps), context.Canceled)
	require.Zero(t, calls.Load())
}

type errString string

func (e errString) Error() string { return string(e) }

var errBoom = errString("boom")
