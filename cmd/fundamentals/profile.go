// profile.go — `comprehensive-profile` cobra subcommand：擷取公司基本資料與
// 高階主管資訊。
package fundamentals

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/config/types"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/obsv"
	"github.com/spf13/cobra"
)

// comprehensiveProfileConfig holds configuration for comprehensive profile command
type comprehensiveProfileConfig struct {
	Ticker  string
	Preview bool
}

// newComprehensiveProfileCmd returns the `comprehensive-profile` cobra command.
func newComprehensiveProfileCmd() *cobra.Command {
	cfg := &comprehensiveProfileConfig{}
	c := &cobra.Command{
		Use:   "comprehensive-profile",
		Short: "擷取公司基本資料與高階主管資訊 (Extract comprehensive company profile information)",
		Long: `擷取公司基本資料，包含公司細節、key executives 與業務摘要。
(Extract comprehensive company profile information including company details,
key executives, and business summary from Yahoo Finance.)

範例 (Examples):
  yfin comprehensive-profile --ticker AAPL
  yfin comprehensive-profile --ticker MSFT --preview`,
		RunE: func(c *cobra.Command, args []string) error { return runComprehensiveProfile(c, cfg) },
	}
	c.Flags().StringVar(&cfg.Ticker, "ticker", "", "Stock symbol to analyze (e.g., AAPL)")
	c.Flags().BoolVar(&cfg.Preview, "preview", false, "Show preview of extracted data")
	return c
}

// runComprehensiveProfile executes the comprehensive profile command
func runComprehensiveProfile(cobraCmd *cobra.Command, cfg *comprehensiveProfileConfig) error {
	if cfg.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}
	runID := cmd.Global.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_comprehensive_profile_%d", time.Now().Unix())
	}

	loader := types.NewLoader(cmd.Global.ConfigFile)
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
	return runComprehensiveProfileExtraction(ctx, client, cfg.Ticker, runID)
}

// runComprehensiveProfileExtraction executes comprehensive profile extraction
func runComprehensiveProfileExtraction(ctx context.Context, client *facade.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for comprehensive profile extraction")
	}
	fmt.Printf("COMPREHENSIVE PROFILE EXTRACTION ticker=%s\n", ticker)
	extractionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	body, meta, err := client.ScrapeFetch(extractionCtx, ticker, "profile")
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
		meta.Host, meta.Status, meta.Bytes, meta.Gzip)

	dto, err := facade.ParseComprehensiveProfile(body, ticker, "NMS")
	if err != nil {
		return fmt.Errorf("failed to parse comprehensive profile: %w", err)
	}
	printComprehensiveProfileSummary(dto)
	return nil
}
