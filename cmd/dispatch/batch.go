// batch.go — `batch` cobra subcommand + worker-pool driver that fans every
// ticker through every entry in `commandRegistry`, honoring the tiered cache
// and writing per-command JSON files (plus `_failed` error logs).
package dispatch

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	sdkconfig "github.com/bizshuk/gosdk/config"
	rootcmd "github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/cache"
	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/spf13/cobra"
)

type tickerResult struct {
	Ticker   string
	Commands map[string]string
	Errors   map[string]error
}

type batchOptions struct {
	ticker     string
	maxWorkers int
	force      bool
}

type batchDeps struct {
	newClient   func() (*facade.Client, error)
	dataDir     func() string
	readTickers func() ([]string, error)
	now         func() time.Time
	registry    map[string]fetchFunc
}

func newProductionBatchDeps() batchDeps {
	return batchDeps{
		newClient:   rootcmd.CreateClient,
		dataDir:     sdkconfig.GetAppDataDir,
		readTickers: readEmbeddedTickerList,
		now:         func() time.Time { return time.Now().UTC() },
		registry:    commandRegistry,
	}
}

// Register attaches the `batch` subcommand onto rootCmd.
func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newBatchCmd(newProductionBatchDeps()))
}

// runBatchForTicker fetches each command for a single ticker, honoring the
// tiered cache. fc may be nil in tests (registries that ignore it).
func runBatchForTicker(ctx context.Context, fc *FetchContext, registry map[string]fetchFunc, ticker string,
	commands []string, force bool, rawDir string, now time.Time,
) tickerResult {
	res := tickerResult{Ticker: ticker, Commands: map[string]string{}, Errors: map[string]error{}}
	for _, command := range commands {
		if ctx.Err() != nil {
			return res
		}
		if cache.ShouldSkip(command, ticker, force, rawDir, now) {
			res.Commands[command] = "skipped"
			continue
		}
		fn, ok := registry[command]
		if !ok {
			recordCommandFailure(&res, rawDir, ticker, command, fmt.Errorf("command %q is not registered", command))
			continue
		}

		data, err := fn(ctx, fc, ticker)
		if err != nil {
			if isNotFoundError(err) {
				res.Commands[command] = "not_found"
				continue
			}
			recordCommandFailure(&res, rawDir, ticker, command, err)
			continue
		}

		outPath := filepath.Join(rawDir, command, fmt.Sprintf("%s.%s.json", ticker, now.Format("2006-01-02")))
		if err := writeJSONAtomic(outPath, data); err != nil {
			recordCommandFailure(&res, rawDir, ticker, command, err)
			continue
		}
		res.Commands[command] = "success"
	}
	return res
}

func isNotFoundError(err error) bool {
	var statusErr *httpx.HTTPError
	return errors.As(err, &statusErr) &&
		(statusErr.StatusCode == 404 || statusErr.StatusCode == 422)
}

func recordCommandFailure(res *tickerResult, rawDir, ticker, command string, cause error) {
	res.Commands[command] = "failed"
	res.Errors[command] = cause
	errPath := filepath.Join(rawDir, "_failed", fmt.Sprintf("%s.%s.err", ticker, command))
	if err := writeErrorAtomic(errPath, cause); err != nil {
		res.Errors[command] = fmt.Errorf("%w; write error artifact: %v", cause, err)
	}
}

func newBatchCmd(deps batchDeps) *cobra.Command {
	options := batchOptions{maxWorkers: 10}
	c := &cobra.Command{
		Use:   "batch",
		Short: "批次擷取 universe 全部 commands 對齊 skills/scripts 行為 (Batch-fetch all commands for a ticker universe — skills/scripts parity)",
		RunE: func(command *cobra.Command, _ []string) error {
			return runBatch(command.Context(), options, deps)
		},
	}
	c.Flags().StringVar(&options.ticker, "ticker", "", "Single ticker (default: ticker_list.csv)")
	c.Flags().IntVar(&options.maxWorkers, "max-workers", 10, "Max concurrent workers")
	c.Flags().BoolVar(&options.force, "force", false, "Force re-fetch, ignore cache")
	return c
}

func runBatch(ctx context.Context, options batchOptions, deps batchDeps) error {
	if options.maxWorkers <= 0 {
		return fmt.Errorf("max-workers must be greater than zero")
	}

	client, err := deps.newClient()
	if err != nil {
		return fmt.Errorf("create facade client: %w", err)
	}
	if client == nil {
		return fmt.Errorf("create facade client: nil client")
	}
	rawDir := filepath.Join(deps.dataDir(), "raw")
	now := deps.now().UTC()

	var tickers []string
	if options.ticker != "" {
		tickers = []string{options.ticker}
	} else {
		tickers, err = deps.readTickers()
		if err != nil {
			return fmt.Errorf("read ticker universe: %w", err)
		}
	}

	fc := &FetchContext{Root: client}

	sem := make(chan struct{}, options.maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var success, skipped, failed, notFound int

	dispatching := true
	for _, t := range tickers {
		select {
		case <-ctx.Done():
			dispatching = false
		default:
		}
		if !dispatching {
			break
		}
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			dispatching = false
		}
		if !dispatching {
			break
		}
		wg.Add(1)
		go func(tk string) {
			defer wg.Done()
			defer func() { <-sem }()
			r := runBatchForTicker(ctx, fc, deps.registry, tk, commandOrder, options.force, rawDir, now)
			mu.Lock()
			for _, st := range r.Commands {
				switch st {
				case "success":
					success++
				case "skipped":
					skipped++
				case "failed":
					failed++
				case "not_found":
					notFound++
				}
			}
			for command, err := range r.Errors {
				fmt.Printf("  %s/%s: %v\n", tk, command, err)
			}
			mu.Unlock()
			fmt.Printf("  %s: %d commands processed\n", tk, len(r.Commands))
		}(t)
	}
	wg.Wait()
	fmt.Printf("Done. success=%d skipped=%d failed=%d not_found=%d\n", success, skipped, failed, notFound)
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
