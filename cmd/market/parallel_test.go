package market

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolveWorkerLimit(t *testing.T) {
	tests := []struct {
		name           string
		cliWorkers     int
		globalWorkers  int
		perHostWorkers int
		itemCount      int
		want           int
		wantErr        string
	}{
		{
			name:           "yaml defaults respect per-host cap",
			globalWorkers:  64,
			perHostWorkers: 32,
			itemCount:      100,
			want:           32,
		},
		{
			name:           "cli override may lower worker count",
			cliWorkers:     4,
			globalWorkers:  64,
			perHostWorkers: 32,
			itemCount:      100,
			want:           4,
		},
		{
			name:           "cli override remains capped per host",
			cliWorkers:     48,
			globalWorkers:  64,
			perHostWorkers: 32,
			itemCount:      100,
			want:           32,
		},
		{
			name:           "worker count does not exceed item count",
			globalWorkers:  64,
			perHostWorkers: 32,
			itemCount:      2,
			want:           2,
		},
		{
			name:           "empty workload needs no workers",
			globalWorkers:  64,
			perHostWorkers: 32,
			itemCount:      0,
			want:           0,
		},
		{
			name:           "negative cli value is invalid",
			cliWorkers:     -1,
			globalWorkers:  64,
			perHostWorkers: 32,
			itemCount:      2,
			wantErr:        "--concurrency must be >= 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveWorkerLimit(
				test.cliWorkers,
				test.globalWorkers,
				test.perHostWorkers,
				test.itemCount,
			)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestMapBoundedLimitsConcurrencyAndPreservesOrder(t *testing.T) {
	const workers = 2

	var active atomic.Int32
	var maximum atomic.Int32
	var ready sync.Once
	allWorkersReady := make(chan struct{})
	release := make(chan struct{})
	resultsReady := make(chan []int, 1)

	go func() {
		resultsReady <- mapBounded([]int{1, 2, 3, 4}, workers, func(value int) int {
			current := active.Add(1)
			for {
				observed := maximum.Load()
				if current <= observed || maximum.CompareAndSwap(observed, current) {
					break
				}
			}
			if current == workers {
				ready.Do(func() {
					close(allWorkersReady)
				})
			}

			<-release
			active.Add(-1)
			return value * 10
		})
	}()

	select {
	case <-allWorkersReady:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("workers did not run concurrently")
	}
	require.Equal(t, int32(workers), maximum.Load())
	close(release)

	require.Equal(t, []int{10, 20, 30, 40}, <-resultsReady)
}
