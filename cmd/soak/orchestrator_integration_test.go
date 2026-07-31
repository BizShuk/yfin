//go:build integration

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bizshuk/yfin/facade"
)

func TestOrchestratorShortLiveRun(t *testing.T) {
	universeFile := filepath.Join(t.TempDir(), "test_universe.txt")
	if err := os.WriteFile(universeFile, []byte("AAPL\n"), 0o644); err != nil {
		t.Fatalf("write universe: %v", err)
	}

	config := &SoakConfig{
		UniverseFile:  universeFile,
		Endpoints:     "key-statistics",
		Fallback:      "scrape-only",
		Duration:      5 * time.Second,
		Concurrency:   1,
		QPS:           0.5,
		ProbeInterval: time.Hour,
		MemoryCheck:   true,
		RunID:         "integration-soak",
	}
	orchestrator, err := NewOrchestrator(facade.NewClient(), discardLogger(), config)
	if err != nil {
		t.Fatalf("create orchestrator: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := orchestrator.Run(ctx); err != nil {
		t.Fatalf("run soak: %v", err)
	}
	if orchestrator.stats.TotalRequests == 0 {
		t.Fatal("expected requests during live soak")
	}
}
