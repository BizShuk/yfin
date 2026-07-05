// batch.go — `batch` cobra subcommand + worker-pool driver that fans every ticker through every entry in `commandRegistry`, honoring the tiered cache and writing per-command JSON files (plus `_failed` error logs). Capacity: 1 `batchCmd` + 3 flags + `runBatchForTicker`/`runBatch` worker loop with bounded concurrency (`batchMaxWorkers` semaphore).
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bizshuk/yfin/utils/cache"
	"github.com/spf13/cobra"
)

const batchRetries = 3

type tickerResult struct {
	Ticker   string
	Commands map[string]string
}

// runBatchForTicker fetches each command for a single ticker, honoring the
// tiered cache. fc may be nil in tests (registries that ignore it).
func runBatchForTicker(ctx context.Context, fc *FetchContext, ticker string,
	commands []string, force bool, rawDir string, now time.Time,
) tickerResult {
	res := tickerResult{Ticker: ticker, Commands: map[string]string{}}
	for _, command := range commands {
		if cache.ShouldSkip(command, ticker, force, rawDir, now) {
			res.Commands[command] = "skipped"
			continue
		}
		fn, ok := commandRegistry[command]
		if !ok {
			res.Commands[command] = "failed"
			continue
		}

		var lastErr error
		var data any
		for attempt := 0; attempt < batchRetries; attempt++ {
			data, lastErr = fn(ctx, fc, ticker)
			if lastErr == nil {
				break
			}
			if attempt < batchRetries-1 {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
			}
		}
		if lastErr != nil {
			errPath := filepath.Join(rawDir, "_failed", fmt.Sprintf("%s.%s.err", ticker, command))
			_ = os.MkdirAll(filepath.Dir(errPath), 0o755)
			_ = os.WriteFile(errPath, []byte(lastErr.Error()), 0o644)
			res.Commands[command] = "failed"
			continue
		}

		outPath := filepath.Join(rawDir, command, fmt.Sprintf("%s.%s.json", ticker, now.Format("2006-01-02")))
		_ = os.MkdirAll(filepath.Dir(outPath), 0o755)
		b, _ := json.MarshalIndent(data, "", "  ")
		_ = os.WriteFile(outPath, b, 0o644)
		res.Commands[command] = "success"
	}
	return res
}

var (
	batchTicker     string
	batchMaxWorkers int
	batchForce      bool
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Batch-fetch all commands for a ticker universe (skills/scripts parity)",
	RunE:  runBatch,
}

func init() {
	batchCmd.Flags().StringVar(&batchTicker, "ticker", "", "Single ticker (default: ticker_list.csv)")
	batchCmd.Flags().IntVar(&batchMaxWorkers, "max-workers", 10, "Max concurrent workers")
	batchCmd.Flags().BoolVar(&batchForce, "force", false, "Force re-fetch, ignore cache")
}

func runBatch(cmd *cobra.Command, args []string) error {
	rawDir := filepath.Join(os.Getenv("HOME"), ".config", "stock", "data", "raw")
	now := time.Now()

	var tickers []string
	if batchTicker != "" {
		tickers = []string{batchTicker}
	} else {
		var err error
		tickers, err = cache.ReadTickerList(
			filepath.Join("yf", "references", "ticker_list.csv"))
		if err != nil {
			return err
		}
	}

	allCommands := make([]string, 0, len(commandRegistry))
	for k := range commandRegistry {
		allCommands = append(allCommands, k)
	}

	ctx := context.Background()
	// TODO(phase-21): wire real *yfinance.Client and authed *yahoo.Client.
	// For now, the FetchContext can be nil; commands that don't touch the
	// root client still work, and the new authed fetchers will be wired in
	// Task 21 / e2e docs phase.
	var fc *FetchContext

	sem := make(chan struct{}, batchMaxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var success, skipped, failed int

	for _, t := range tickers {
		wg.Add(1)
		sem <- struct{}{}
		go func(tk string) {
			defer wg.Done()
			defer func() { <-sem }()
			r := runBatchForTicker(ctx, fc, tk, allCommands, batchForce, rawDir, now)
			mu.Lock()
			for _, st := range r.Commands {
				switch st {
				case "success":
					success++
				case "skipped":
					skipped++
				case "failed":
					failed++
				}
			}
			mu.Unlock()
			fmt.Printf("  %s: %d commands processed\n", tk, len(r.Commands))
		}(t)
	}
	wg.Wait()
	fmt.Printf("Done. success=%d skipped=%d failed=%d\n", success, skipped, failed)
	return nil
}
