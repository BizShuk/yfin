package dispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bizshuk/yfin/facade"
	"github.com/stretchr/testify/require"
)

func TestRunBatchLifecycleAllCommands(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	deps := integrationBatchDeps(t, root, now, "")

	require.NoError(t, runBatch(context.Background(), batchOptions{
		ticker:     "AAPL",
		maxWorkers: 1,
		force:      true,
	}, deps))

	for _, command := range commandOrder {
		path := filepath.Join(root, "raw", command, "AAPL.2026-07-16.json")
		data, err := os.ReadFile(path)
		require.NoError(t, err, command)
		var got map[string]string
		require.NoError(t, json.Unmarshal(data, &got), command)
		require.Equal(t, command, got["command"])
		require.Equal(t, "AAPL", got["symbol"])
	}
}

func TestRunBatchReturnsErrorAndPreservesSuccessfulArtifacts(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	deps := integrationBatchDeps(t, root, now, "income")

	err := runBatch(context.Background(), batchOptions{
		ticker:     "AAPL",
		maxWorkers: 1,
		force:      true,
	}, deps)

	require.EqualError(t, err, "batch completed with 1 failed command(s)")
	require.FileExists(t, filepath.Join(root, "raw", "info", "AAPL.2026-07-16.json"))
	require.FileExists(t, filepath.Join(root, "raw", "metadata", "AAPL.2026-07-16.json"))
	require.NoFileExists(t, filepath.Join(root, "raw", "income", "AAPL.2026-07-16.json"))
	require.FileExists(t, filepath.Join(root, "raw", "_failed", "AAPL.income.err"))
}

func integrationBatchDeps(t *testing.T, root string, now time.Time, failCommand string) batchDeps {
	t.Helper()
	registry := make(map[string]fetchFunc, len(commandOrder))
	for _, command := range commandOrder {
		registry[command] = func(_ context.Context, _ *FetchContext, symbol string) (any, error) {
			if command == failCommand {
				return nil, fmt.Errorf("forced %s failure", command)
			}
			return map[string]string{"command": command, "symbol": symbol}, nil
		}
	}
	return batchDeps{
		newClient:   func() (*facade.Client, error) { return facade.NewClient(), nil },
		dataDir:     func() string { return root },
		readTickers: func() ([]string, error) { return []string{"AAPL"}, nil },
		now:         func() time.Time { return now },
		registry:    registry,
	}
}
