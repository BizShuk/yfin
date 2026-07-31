package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemorySummaryUsesMonitorOwnedPeaksAndSignedGrowth(t *testing.T) {
	finalTime := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	monitor := &MemoryMonitor{
		initialMemory:     200,
		peakMemory:        300,
		initialGoroutines: 4,
		peakGoroutines:    8,
		samples: []MemorySample{{
			Timestamp:  finalTime,
			AllocBytes: 100,
			Goroutines: 3,
		}},
	}

	summary := monitor.Summary()

	require.Equal(t, uint64(200), summary.InitialMemory)
	require.Equal(t, uint64(300), summary.PeakMemory)
	require.Equal(t, 4, summary.InitialGoroutines)
	require.Equal(t, 8, summary.PeakGoroutines)
	require.Equal(t, finalTime, summary.Final.Timestamp)
	require.Equal(t, int64(-100), summary.MemoryGrowthBytes())
	require.Equal(t, -1, summary.GoroutineGrowth())
}
