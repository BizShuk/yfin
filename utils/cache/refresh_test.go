// — Unit tests for `ShouldSkip` covering daily, quarterly, annually tiers and `force=true` cache bypass. Capacity: 3 test funcs × 4 sub-cases.

package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShouldSkip_DailyTierSkipsSameDay(t *testing.T) {
	root := t.TempDir()
	now := time.Now()
	cmdDir := filepath.Join(root, "history")
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	fn := filepath.Join(cmdDir, "AAPL."+now.Format("2006-01-02")+".json")
	require.NoError(t, os.WriteFile(fn, []byte("{}"), 0o644))

	require.True(t, ShouldSkip("history", "AAPL", false, root, now))
	require.False(t, ShouldSkip("history", "AAPL", true, root, now))
}

func TestShouldSkip_QuarterlyTier(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)
	cmdDir := filepath.Join(root, "income")
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	// Apr 10 is same quarter (Q2: Apr-Jun) as Jun 23
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "AAPL.2026-04-10.json"), []byte("{}"), 0o644))
	require.True(t, ShouldSkip("income", "AAPL", false, root, now))

	// An older Q1 artifact does not invalidate the newer Q2 artifact.
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "AAPL.2026-01-05.json"), []byte("{}"), 0o644))
	require.True(t, ShouldSkip("income", "AAPL", false, root, now))
}

func TestShouldSkip_AnnuallyTier(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)
	cmdDir := filepath.Join(root, "isin")
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	// same year (2026) → skip
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "AAPL.2026-02-01.json"), []byte("{}"), 0o644))
	require.True(t, ShouldSkip("isin", "AAPL", false, root, now))
	// A previous-year artifact does not invalidate the newer same-year artifact.
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "AAPL.2025-12-31.json"), []byte("{}"), 0o644))
	require.True(t, ShouldSkip("isin", "AAPL", false, root, now))
}

func TestShouldSkipUsesNewestArtifact(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)
	cmdDir := filepath.Join(root, "history")
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "AAPL.2025-01-01.json"), []byte("{}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "AAPL.2026-06-23.json"), []byte("{}"), 0o644))

	require.True(t, ShouldSkip("history", "AAPL", false, root, now))
}
