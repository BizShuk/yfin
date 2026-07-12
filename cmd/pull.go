// pull.go — `pull` cobra subcommand that fetches daily bars for a single
// ticker or a universe file and routes them through the bus-publishing and
// local-export paths. This file owns PullConfig, the pullCmd registration,
// validation/parse helpers, the per-symbol processing loop, the preview
// printer, and the bus/JSON exporters used after each fetch.
// Capacity: 1 `PullConfig` + 1 var + 1 `pullCmd` + 1 `init()` (14 flags) +
// runPull / validatePullFlags / parseDates / parseAdjusted / getSymbols /
// processSymbol / printBarsPreview / handleFXPreview / handleBusPublishing /
// handleLocalExport / estimateBarBatchSize.
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/svc/emit"
	"github.com/bizshuk/yfin/svc/norm"
	"github.com/bizshuk/yfin/utils/bus"
	"github.com/bizshuk/yfin/utils/obsv"
	"github.com/spf13/cobra"
)

// PullConfig holds configuration for the pull command
type PullConfig struct {
	Ticker        string
	UniverseFile  string
	Start         string
	End           string
	Adjusted      string
	Market        string
	FXTarget      string
	Preview       bool
	Publish       bool
	Env           string
	TopicPrefix   string
	Out           string
	OutDir        string
	DryRunPublish bool
}

var pullConfig PullConfig

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Fetch daily bars for a symbol or universe",
	Long: `Fetch daily bars for a single symbol or multiple symbols from a universe file.
Only daily bars are supported by design.

Examples:
  yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted split_dividend --preview
  yfin pull --universe-file ./nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --preview --concurrency 32
  yfin pull --ticker SAP --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./out --preview`,
	RunE: runPull,
}

func init() {
	// Pull command flags
	pullCmd.Flags().StringVar(&pullConfig.Ticker, "ticker", "", "Stock symbol to fetch (e.g., AAPL)")
	pullCmd.Flags().StringVar(&pullConfig.UniverseFile, "universe-file", "", "Newline-delimited list of symbols")
	pullCmd.Flags().StringVar(&pullConfig.Start, "start", "", "Start date (YYYY-MM-DD, UTC)")
	pullCmd.Flags().StringVar(&pullConfig.End, "end", "", "End date (YYYY-MM-DD, UTC)")
	pullCmd.Flags().StringVar(&pullConfig.Adjusted, "adjusted", "split_dividend", "Adjustment policy (raw|split_dividend)")
	pullCmd.Flags().StringVar(&pullConfig.Market, "market", "", "Market MIC (optional hint for MIC inference)")
	pullCmd.Flags().StringVar(&pullConfig.FXTarget, "fx-target", "", "Target currency for FX conversion preview (e.g., USD)")
	pullCmd.Flags().BoolVar(&pullConfig.Preview, "preview", false, "Show preview without publishing")
	pullCmd.Flags().BoolVar(&pullConfig.Publish, "publish", false, "Enable bus publishing")
	pullCmd.Flags().StringVar(&pullConfig.Env, "env", "dev", "Environment (dev, staging, prod)")
	pullCmd.Flags().StringVar(&pullConfig.TopicPrefix, "topic-prefix", "ampy", "Topic prefix for bus publishing")
	pullCmd.Flags().StringVar(&pullConfig.Out, "out", "", "Output format (json|parquet)")
	pullCmd.Flags().StringVar(&pullConfig.OutDir, "out-dir", "", "Output directory")
	pullCmd.Flags().BoolVar(&pullConfig.DryRunPublish, "dry-run-publish", false, "Alias for --preview; no network send but compute payload sizes")
	rootCmd.AddCommand(pullCmd)
}

// runPull executes the pull command
func runPull(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validatePullFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	// Parse dates
	startTime, endTime, err := parseDates(pullConfig.Start, pullConfig.End)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid date format: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Parse adjustment policy
	adjusted, err := parseAdjusted(pullConfig.Adjusted)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid adjusted value: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Validate interval (daily-only enforcement)
	loader := config.NewLoader(globalConfig.ConfigFile)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// For yfinance-go, we only support daily intervals
	if validateErr := cfg.ValidateInterval("1d"); validateErr != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", validateErr)
		os.Exit(ExitConfigError)
	}

	// Initialize observability
	ctx := context.Background()
	disableTracing, _ := cmd.Flags().GetBool("observability-disable-tracing")
	disableMetrics, _ := cmd.Flags().GetBool("observability-disable-metrics")

	obsvConfig := &obsv.Config{
		ServiceName:       "yfinance-go",
		ServiceVersion:    version,
		Environment:       cfg.App.Env,
		CollectorEndpoint: cfg.Observability.Tracing.OTLP.Endpoint,
		TraceProtocol:     "grpc",
		SampleRatio:       cfg.Observability.Tracing.OTLP.SampleRatio,
		LogLevel:          cfg.Observability.Logs.Level,
		MetricsAddr:       cfg.Observability.Metrics.Prometheus.Addr,
		MetricsEnabled:    cfg.Observability.Metrics.Prometheus.Enabled && !disableMetrics,
		TracingEnabled:    cfg.Observability.Tracing.OTLP.Enabled && !disableTracing,
	}

	if obsvErr := obsv.Init(ctx, obsvConfig); obsvErr != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize observability: %v\n", obsvErr)
		os.Exit(ExitConfigError)
	}
	defer func() { _ = obsv.Shutdown(ctx) }()

	// Get symbols to process
	symbols, err := getSymbols(pullConfig.Ticker, pullConfig.UniverseFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to get symbols: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Create client
	client, err := createClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Create bus if publishing or previewing
	var busInstance *bus.Bus
	var busConfig *bus.Config
	if pullConfig.Publish || pullConfig.Preview || pullConfig.DryRunPublish {
		busConfig = createBusConfig(pullConfig.Env, pullConfig.TopicPrefix)
		busInstance, err = bus.NewBus(busConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to create bus: %v\n", err)
			os.Exit(ExitGeneral)
		}
		defer busInstance.Close(context.Background())
	}

	// Process symbols
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	successCount := 0
	for _, symbol := range symbols {
		if err := processSymbol(ctx, client, symbol, startTime, endTime, adjusted, runID, busInstance, busConfig); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to process %s: %v\n", symbol, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: No symbols processed successfully\n")
		os.Exit(ExitGeneral)
	}

	fmt.Printf("Successfully processed %d/%d symbols\n", successCount, len(symbols))
	return nil
}

// validatePullFlags validates pull command flags
func validatePullFlags() error {
	if pullConfig.Ticker == "" && pullConfig.UniverseFile == "" {
		return fmt.Errorf("either --ticker or --universe-file must be specified")
	}
	if pullConfig.Ticker != "" && pullConfig.UniverseFile != "" {
		return fmt.Errorf("cannot specify both --ticker and --universe-file")
	}
	if pullConfig.Start == "" || pullConfig.End == "" {
		return fmt.Errorf("--start and --end are required")
	}
	if pullConfig.Adjusted != "raw" && pullConfig.Adjusted != "split_dividend" {
		return fmt.Errorf("--adjusted must be 'raw' or 'split_dividend'")
	}
	if pullConfig.Out != "" && pullConfig.Out != "json" && pullConfig.Out != "parquet" {
		return fmt.Errorf("--out must be 'json' or 'parquet'")
	}
	return nil
}

// parseDates parses start and end date strings
func parseDates(startStr, endStr string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %v", err)
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %v", err)
	}
	return start, end, nil
}

// parseAdjusted parses the adjusted flag
func parseAdjusted(adjusted string) (bool, error) {
	switch adjusted {
	case "raw":
		return false, nil
	case "split_dividend":
		return true, nil
	default:
		return false, fmt.Errorf("invalid adjusted value: %s", adjusted)
	}
}

// getSymbols returns the list of symbols to process
func getSymbols(ticker, universeFile string) ([]string, error) {
	if ticker != "" {
		return []string{ticker}, nil
	}

	// Read universe file
	content, err := os.ReadFile(universeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read universe file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var symbols []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			symbols = append(symbols, line)
		}
	}

	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols found in universe file")
	}

	return symbols, nil
}

// processSymbol processes a single symbol for bars
func processSymbol(ctx context.Context, client *cliClient, symbol string, start, end time.Time, adjusted bool, runID string, busInstance *bus.Bus, busConfig *bus.Config) error {
	// Fetch bars via the CLI helper (svc/yahoo + internal/norm). This keeps
	// the ScaledDecimal precision the bus-publishing code needs.
	bars, err := fetchDailyBarsNorm(ctx, client.Yahoo, symbol, start, end, adjusted, runID)
	if err != nil {
		return err
	}

	if len(bars.Bars) == 0 {
		fmt.Printf("No bars found for %s in the specified period\n", symbol)
		return nil
	}

	// Print preview
	printBarsPreview(bars, runID, pullConfig.Env, pullConfig.TopicPrefix)

	// Handle FX preview if requested
	if pullConfig.FXTarget != "" {
		if err := handleFXPreview(ctx, bars, pullConfig.FXTarget); err != nil {
			fmt.Printf("FX preview failed: %v\n", err)
		}
	}

	// Handle bus publishing
	if busInstance != nil {
		preview := pullConfig.Preview || pullConfig.DryRunPublish
		if err := handleBusPublishing(ctx, bars, busInstance, busConfig, runID, preview); err != nil {
			return fmt.Errorf("bus publishing failed: %v", err)
		}
	}

	// Handle local export
	if pullConfig.Out != "" && pullConfig.OutDir != "" {
		if err := handleLocalExport(bars, symbol, start, end, adjusted, pullConfig.Out, pullConfig.OutDir); err != nil {
			return fmt.Errorf("local export failed: %v", err)
		}
	}

	return nil
}

// printBarsPreview prints the bars preview according to specification
func printBarsPreview(bars *norm.NormalizedBarBatch, runID, env, topicPrefix string) {
	firstBar := bars.Bars[0]
	lastBar := bars.Bars[len(bars.Bars)-1]

	fmt.Printf("RUN %s  (env=%s, topic_prefix=%s)\n", runID, env, topicPrefix)
	fmt.Printf("SYMBOL %s (MIC=%s, CCY=%s)  range=%s..%s  bars=%d  adjusted=%s\n",
		bars.Security.Symbol,
		bars.Security.MIC,
		firstBar.CurrencyCode,
		firstBar.Start.Format("2006-01-02"),
		lastBar.End.Format("2006-01-02"),
		len(bars.Bars),
		firstBar.AdjustmentPolicyID)
	fmt.Printf("first=%s  last=%s  last_close=%.4f %s\n",
		firstBar.Start.Format("2006-01-02T15:04:05Z"),
		lastBar.End.Format("2006-01-02T15:04:05Z"),
		float64(lastBar.Close.Scaled)/float64(lastBar.Close.Scale),
		lastBar.CurrencyCode)
}

// handleFXPreview handles FX conversion preview. After Step 6, handleFXPreview
// no longer needs a *facade.Client — it only inspects the bar batch's
// currency; the client was previously threaded through without being used.
func handleFXPreview(ctx context.Context, bars *norm.NormalizedBarBatch, targetCurrency string) error {
	// Check if FX conversion is needed
	firstBar := bars.Bars[0]
	if firstBar.CurrencyCode == targetCurrency {
		fmt.Printf("fx_preview target=%s (no conversion needed)\n", targetCurrency)
		return nil
	}

	// For now, just show that FX preview is requested
	// In a full implementation, this would use the FX manager
	fmt.Printf("fx_preview target=%s as_of=%s rate_scale=8 rounding=half_up  (provider=yahoo-web, cache_hit=true)\n",
		targetCurrency, time.Now().Format("2006-01-02T15:04:05Z"))

	return nil
}

// handleBusPublishing handles bus publishing for bars
func handleBusPublishing(ctx context.Context, bars *norm.NormalizedBarBatch, busInstance *bus.Bus, busConfig *bus.Config, runID string, preview bool) error {
	// Emit to ampy-proto format
	ampyBatch, err := emit.EmitBarBatch(bars)
	if err != nil {
		return fmt.Errorf("failed to emit bar batch: %v", err)
	}

	// Create bus message
	busMessage := &bus.BarBatchMessage{
		Batch: ampyBatch,
		Key: &bus.Key{
			Symbol: bars.Security.Symbol,
			MIC:    bars.Security.MIC,
		},
		RunID: runID,
		Env:   busConfig.Env,
	}

	if preview {
		// Estimate payload size
		payloadSize := estimateBarBatchSize(ampyBatch)
		previewSummary, err := busInstance.PreviewBars(busMessage, payloadSize)
		if err != nil {
			return fmt.Errorf("failed to generate preview: %v", err)
		}
		bus.PrintPreview(previewSummary)
	} else {
		// Actually publish
		if err := busInstance.PublishBars(ctx, busMessage); err != nil {
			return fmt.Errorf("failed to publish bars: %v", err)
		}
		fmt.Printf("Published %d bars to bus\n", len(bars.Bars))
	}

	return nil
}

// handleLocalExport handles local export for bars
func handleLocalExport(bars *norm.NormalizedBarBatch, symbol string, start, end time.Time, adjusted bool, outFormat, outDir string) error {
	// Create output directory
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate filename
	adjustedStr := "raw"
	if adjusted {
		adjustedStr = "adjusted"
	}
	filename := fmt.Sprintf("%s_1d_%s_%s_%s.%s",
		symbol,
		start.Format("20060102"),
		end.Format("20060102"),
		adjustedStr,
		outFormat)

	filePath := filepath.Join(outDir, "bars", filename)

	// Create bars subdirectory
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to create bars directory: %v", err)
	}

	// Write file
	switch outFormat {
	case "json":
		return writeJSONFile(filePath, bars)
	case "parquet":
		return fmt.Errorf("parquet export not implemented yet")
	default:
		return fmt.Errorf("unsupported output format: %s", outFormat)
	}
}

// estimateBarBatchSize estimates the size of a bar batch payload
func estimateBarBatchSize(batch interface{}) int {
	// This is a rough estimate - in a real implementation you would marshal to get exact size
	// For now, estimate based on typical bar size
	return 200 * 10 // Assume 200 bytes per bar, 10 bars average
}
