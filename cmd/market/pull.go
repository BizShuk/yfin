// pull.go — `pull` cobra subcommand：擷取每日 bars。透過 facade 取得
// ScaledDecimal 精度的 *model.NormalizedBarBatch，再本地 JSON 匯出。
// 同 package 的 client_json.go 提供 writeJSONFile 共用。
package market

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/utils/obsv"
	"github.com/spf13/cobra"
)

// pullConfig holds configuration for the pull command
type pullConfig struct {
	Ticker       string
	UniverseFile string
	Start        string
	End          string
	Adjusted     string
	Market       string
	FXTarget     string
	Out          string
	OutDir       string
}

type pullFetchResult struct {
	bars *model.NormalizedBarBatch
	err  error
}

// newPullCmd returns the `pull` cobra command.
func newPullCmd() *cobra.Command {
	cfg := &pullConfig{}
	c := &cobra.Command{
		Use:   "pull",
		Short: "擷取單一 symbol 或 universe 的每日 bars (Fetch daily bars for a symbol or universe)",
		Long: `擷取單一 symbol 或 universe 檔中多個 symbols 的每日 bars；僅支援 daily 區間。
(Fetch daily bars for a single symbol or multiple symbols from a universe file.
Only daily bars are supported by design.)

範例 (Examples):
  yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted split_dividend
  yfin pull --universe-file ./nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --concurrency 32
  yfin pull --ticker SAP --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./out`,
		RunE: func(c *cobra.Command, args []string) error { return runPull(c, cfg) },
	}
	// Pull command flags
	c.Flags().StringVar(&cfg.Ticker, "ticker", "", "Stock symbol to fetch (e.g., AAPL)")
	c.Flags().StringVar(&cfg.UniverseFile, "universe-file", "", "Newline-delimited list of symbols")
	c.Flags().StringVar(&cfg.Start, "start", "", "Start date (YYYY-MM-DD, UTC)")
	c.Flags().StringVar(&cfg.End, "end", "", "End date (YYYY-MM-DD, UTC)")
	c.Flags().StringVar(&cfg.Adjusted, "adjusted", "split_dividend", "Adjustment policy (raw|split_dividend)")
	c.Flags().StringVar(&cfg.Market, "market", "", "Market MIC or Yahoo exchange code (optional metadata hint)")
	c.Flags().StringVar(&cfg.FXTarget, "fx-target", "", "Validate output currency; cross-currency conversion is unavailable (e.g., USD)")
	c.Flags().StringVar(&cfg.Out, "out", "", "Output format (json)")
	c.Flags().StringVar(&cfg.OutDir, "out-dir", "", "Output directory")
	return c
}

// runPull executes the pull command
func runPull(cobraCmd *cobra.Command, cfg *pullConfig) error {
	// Validate flags
	if err := validatePullFlags(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// Generate run ID if not provided
	runID := cmd.Global.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	// Parse dates
	startTime, endTime, err := parseDates(cfg.Start, cfg.End)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid date format: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// Parse adjustment policy
	adjusted, err := parseAdjusted(cfg.Adjusted)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid adjusted value: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// Validate interval (daily-only enforcement)
	loader := config.NewLoader(cmd.Global.ConfigFile)
	ycfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// yfin supports daily intervals only.
	if validateErr := ycfg.ValidateInterval("1d"); validateErr != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", validateErr)
		os.Exit(cmd.ExitConfigError)
	}

	// Initialize observability. Persistent flags (--observability-disable-*)
	// live on the root command; pull reads them through the cobra Command.
	ctx := context.Background()
	disableTracing, _ := cobraCmd.Flags().GetBool("observability-disable-tracing")
	disableMetrics, _ := cobraCmd.Flags().GetBool("observability-disable-metrics")

	obsvConfig := &obsv.Config{
		ServiceName:       "yfin",
		ServiceVersion:    cmd.Version,
		Environment:       ycfg.App.Env,
		CollectorEndpoint: ycfg.Observability.Tracing.OTLP.Endpoint,
		TraceProtocol:     "grpc",
		SampleRatio:       ycfg.Observability.Tracing.OTLP.SampleRatio,
		LogLevel:          cmd.ResolveLogLevel(cobraCmd, ycfg.Observability.Logs.Level),
		MetricsAddr:       ycfg.Observability.Metrics.Prometheus.Addr,
		MetricsEnabled:    ycfg.Observability.Metrics.Prometheus.Enabled && !disableMetrics,
		TracingEnabled:    ycfg.Observability.Tracing.OTLP.Enabled && !disableTracing,
	}

	if obsvErr := obsv.Init(ctx, obsvConfig); obsvErr != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize observability: %v\n", obsvErr)
		os.Exit(cmd.ExitConfigError)
	}
	defer func() { _ = obsv.Shutdown(ctx) }()

	// Get symbols to process
	symbols, err := getSymbols(cfg.Ticker, cfg.UniverseFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to get symbols: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// Create client
	client, err := cmd.CreateClientFromConfig(ycfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(cmd.ExitGeneral)
	}

	workerLimit, err := resolveWorkerLimit(
		cmd.Global.Concurrency,
		ycfg.Concurrency.GlobalWorkers,
		ycfg.Concurrency.PerHostWorkers,
		len(symbols),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid concurrency: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// Fetch concurrently, then process output in the original symbol order.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	fetchResults := mapBounded(symbols, workerLimit, func(symbol string) pullFetchResult {
		bars, fetchErr := client.FetchDailyBarsNorm(ctx, symbol, startTime, endTime, adjusted, runID)
		return pullFetchResult{bars: bars, err: fetchErr}
	})

	successCount := 0
	for index, symbol := range symbols {
		result := fetchResults[index]
		if result.err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to process %s: %v\n", symbol, result.err)
			continue
		}
		if err := processPulledSymbol(ctx, result.bars, symbol, startTime, endTime, adjusted, runID, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to process %s: %v\n", symbol, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: No symbols processed successfully\n")
		os.Exit(cmd.ExitGeneral)
	}

	fmt.Printf("Successfully processed %d/%d symbols\n", successCount, len(symbols))
	return nil
}

// validatePullFlags validates pull command flags
func validatePullFlags(cfg *pullConfig) error {
	if cfg.Ticker == "" && cfg.UniverseFile == "" {
		return fmt.Errorf("either --ticker or --universe-file must be specified")
	}
	if cfg.Ticker != "" && cfg.UniverseFile != "" {
		return fmt.Errorf("cannot specify both --ticker and --universe-file")
	}
	if cfg.Start == "" || cfg.End == "" {
		return fmt.Errorf("--start and --end are required")
	}
	if cfg.Adjusted != "raw" && cfg.Adjusted != "split_dividend" {
		return fmt.Errorf("--adjusted must be 'raw' or 'split_dividend'")
	}
	if cfg.Out != "" && cfg.Out != "json" {
		return fmt.Errorf("--out must be 'json'")
	}
	if cfg.Out != "" && cfg.OutDir == "" {
		return fmt.Errorf("--out-dir is required when --out is set")
	}
	if cfg.Out == "" && cfg.OutDir != "" {
		return fmt.Errorf("--out is required when --out-dir is set")
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

// processPulledSymbol renders and exports one already-fetched bar batch.
func processPulledSymbol(ctx context.Context, bars *model.NormalizedBarBatch, symbol string, start, end time.Time, adjusted bool, runID string, cfg *pullConfig) error {
	if cfg == nil {
		return fmt.Errorf("pull configuration is required")
	}
	if err := applyMarketHint(bars, cfg.Market); err != nil {
		return err
	}
	if len(bars.Bars) == 0 {
		fmt.Printf("No bars found for %s in the specified period\n", symbol)
		return nil
	}

	// Print preview
	printBarsPreview(bars, runID)

	// Handle FX preview if requested
	if cfg.FXTarget != "" {
		if err := handleFXPreview(ctx, bars, cfg.FXTarget); err != nil {
			return err
		}
	}

	// Handle local export
	if cfg.Out != "" && cfg.OutDir != "" {
		if err := handleLocalExport(bars, symbol, start, end, adjusted, cfg.Out, cfg.OutDir); err != nil {
			return fmt.Errorf("local export failed: %v", err)
		}
	}

	return nil
}

// applyMarketHint normalizes a known Yahoo exchange code to its MIC, then
// fills missing provider metadata without overwriting a known value. The hint
// does not select a Yahoo endpoint or rewrite the input symbol.
func applyMarketHint(bars *model.NormalizedBarBatch, hint string) error {
	if bars == nil {
		return fmt.Errorf("bar batch is required")
	}
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return nil
	}
	if inferredMIC := model.InferMIC(hint, hint); inferredMIC != "" {
		hint = inferredMIC
	} else {
		hint = strings.ToUpper(hint)
	}
	if len(hint) < 3 || len(hint) > 4 {
		return fmt.Errorf("invalid market MIC %q", hint)
	}
	for _, character := range hint {
		if (character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') {
			return fmt.Errorf("invalid market MIC %q", hint)
		}
	}
	if bars.Security.MIC != "" {
		return nil
	}
	bars.Security.MIC = hint
	return nil
}

// printBarsPreview prints the bars preview according to specification
func printBarsPreview(bars *model.NormalizedBarBatch, runID string) {
	firstBar := bars.Bars[0]
	lastBar := bars.Bars[len(bars.Bars)-1]

	fmt.Printf("RUN %s\n", runID)
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
		lastBar.Close.Float64(),
		lastBar.CurrencyCode)
}

// handleFXPreview validates whether the requested target already matches the
// fetched data. yfin does not currently own an FX rate provider, so it must
// reject cross-currency requests instead of reporting a fabricated preview.
func handleFXPreview(ctx context.Context, bars *model.NormalizedBarBatch, targetCurrency string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if bars == nil || len(bars.Bars) == 0 {
		return fmt.Errorf("FX preview requires at least one bar")
	}

	firstBar := bars.Bars[0]
	sourceCurrency := strings.ToUpper(strings.TrimSpace(firstBar.CurrencyCode))
	targetCurrency = strings.ToUpper(strings.TrimSpace(targetCurrency))
	if targetCurrency == "" {
		return fmt.Errorf("FX target currency is required")
	}
	if sourceCurrency == targetCurrency {
		fmt.Printf("fx_preview target=%s (no conversion needed)\n", targetCurrency)
		return nil
	}

	return fmt.Errorf(
		"FX conversion is unavailable: source currency %s does not match target currency %s",
		sourceCurrency,
		targetCurrency,
	)
}

// handleLocalExport handles local export for bars
func handleLocalExport(bars *model.NormalizedBarBatch, symbol string, start, end time.Time, adjusted bool, outFormat, outDir string) error {
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
	default:
		return fmt.Errorf("unsupported output format: %s", outFormat)
	}
}
