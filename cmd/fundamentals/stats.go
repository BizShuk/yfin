// stats.go — `comprehensive-stats` cobra subcommand：擷取完整 key statistics
// 含當前值與 5 年歷史。
package fundamentals

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/cmd/format"
	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/obsv"
	"github.com/spf13/cobra"
)

// comprehensiveStatsConfig holds configuration for comprehensive statistics command
type comprehensiveStatsConfig struct {
	Ticker  string
	Preview bool
}

// newComprehensiveStatsCmd returns the `comprehensive-stats` cobra command.
func newComprehensiveStatsCmd() *cobra.Command {
	cfg := &comprehensiveStatsConfig{}
	c := &cobra.Command{
		Use:   "comprehensive-stats",
		Short: "擷取完整 key statistics 含 5 年歷史資料 (Extract comprehensive key statistics with historical data)",
		Long: `擷取完整 key statistics，含當前值與 5 年歷史資料。
使用 YAML 設定的 regex 模式從 Yahoo Finance 萃取所有 key statistics。
(Extract comprehensive key statistics including current values and 5-year historical data.
This command uses YAML-configured regex patterns to extract all key statistics from Yahoo Finance.)

範例 (Examples):
  yfin comprehensive-stats --ticker AAPL
  yfin comprehensive-stats --ticker MSFT --preview`,
		RunE: func(c *cobra.Command, args []string) error { return runComprehensiveStats(c, cfg) },
	}
	c.Flags().StringVar(&cfg.Ticker, "ticker", "", "Stock symbol to analyze (e.g., AAPL)")
	c.Flags().BoolVar(&cfg.Preview, "preview", false, "Show preview of extracted data")
	return c
}

// runComprehensiveStats executes the comprehensive statistics command
func runComprehensiveStats(cobraCmd *cobra.Command, cfg *comprehensiveStatsConfig) error {
	if cfg.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}
	runID := cmd.Global.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_comprehensive_stats_%d", time.Now().Unix())
	}

	loader := config.NewLoader(cmd.Global.ConfigFile)
	ycfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}
	scrapeCfg := ycfg.GetScrapeConfig()
	if !scrapeCfg.Enabled {
		fmt.Fprintf(os.Stderr, "ERROR: Scraping is disabled in configuration\n")
		os.Exit(cmd.ExitConfigError)
	}

	ctx := context.Background()
	disableTracing, _ := cobraCmd.Flags().GetBool("observability-disable-tracing")
	disableMetrics, _ := cobraCmd.Flags().GetBool("observability-disable-metrics")
	obsvConfig := &obsv.Config{
		ServiceName:       "yfinance-go",
		ServiceVersion:    cmd.Version,
		Environment:       ycfg.App.Env,
		CollectorEndpoint: ycfg.Observability.Tracing.OTLP.Endpoint,
		TraceProtocol:     "grpc",
		SampleRatio:       ycfg.Observability.Tracing.OTLP.SampleRatio,
		LogLevel:          ycfg.Observability.Logs.Level,
		MetricsAddr:       ycfg.Observability.Metrics.Prometheus.Addr,
		MetricsEnabled:    ycfg.Observability.Metrics.Prometheus.Enabled && !disableMetrics,
		TracingEnabled:    ycfg.Observability.Tracing.OTLP.Enabled && !disableTracing,
	}
	if obsvErr := obsv.Init(ctx, obsvConfig); obsvErr != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to initialize observability: %v\n", obsvErr)
		os.Exit(cmd.ExitConfigError)
	}
	defer func() { _ = obsv.Shutdown(ctx) }()

	client, err := cmd.CreateClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(cmd.ExitGeneral)
	}
	_ = scrapeCfg
	return runComprehensiveStatsExtraction(ctx, client, cfg.Ticker, runID)
}

// runComprehensiveStatsExtraction executes comprehensive statistics extraction
func runComprehensiveStatsExtraction(ctx context.Context, client *facade.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for comprehensive stats extraction")
	}
	fmt.Printf("COMPREHENSIVE STATISTICS EXTRACTION ticker=%s\n", ticker)
	extractionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	body, meta, err := client.ScrapeFetch(extractionCtx, ticker, "key-statistics")
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
		meta.Host, meta.Status, meta.Bytes, meta.Gzip)

	dto, err := facade.ParseComprehensiveKeyStatistics(body, ticker, "NMS")
	if err != nil {
		return fmt.Errorf("failed to parse comprehensive statistics: %w", err)
	}
	format.ComprehensiveStatistics(dto)
	return nil
}
