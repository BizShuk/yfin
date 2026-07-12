// comprehensive_stats.go — `comprehensive-stats` cobra subcommand that fetches
// the key-statistics page and prints the parsed `ComprehensiveKeyStatisticsDTO`
// (current + additional + 5-year historical) to stdout via
// `printComprehensiveStatisticsSummary`. DTO → stdout rendering lives in this
// file because no other subcommand consumes that DTO.
// Capacity: 1 `ComprehensiveStatsConfig` + 1 var + 1 `comprehensiveStatsCmd` +
// 1 `init()` (2 flags) + runComprehensiveStats + runComprehensiveStatsExtraction
// + printComprehensiveStatisticsSummary.
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/svc/scrape"
	"github.com/bizshuk/yfin/utils/obsv"
	"github.com/spf13/cobra"
)

// ComprehensiveStatsConfig holds configuration for comprehensive statistics command
type ComprehensiveStatsConfig struct {
	Ticker  string
	Preview bool
}

var comprehensiveStatsConfig ComprehensiveStatsConfig

// comprehensiveStatsCmd represents the comprehensive statistics command
var comprehensiveStatsCmd = &cobra.Command{
	Use:   "comprehensive-stats",
	Short: "Extract comprehensive key statistics with historical data",
	Long: `Extract comprehensive key statistics including current values and 5-year historical data.
This command uses YAML-configured regex patterns to extract all key statistics from Yahoo Finance.

Examples:
  yfin comprehensive-stats --ticker AAPL
  yfin comprehensive-stats --ticker MSFT --preview`,
	RunE: runComprehensiveStats,
}

func init() {
	// Comprehensive stats command flags
	comprehensiveStatsCmd.Flags().StringVar(&comprehensiveStatsConfig.Ticker, "ticker", "", "Stock symbol to analyze (e.g., AAPL)")
	comprehensiveStatsCmd.Flags().BoolVar(&comprehensiveStatsConfig.Preview, "preview", false, "Show preview of extracted data")
	rootCmd.AddCommand(comprehensiveStatsCmd)
}

// runComprehensiveStats executes the comprehensive statistics command
func runComprehensiveStats(cmd *cobra.Command, args []string) error {
	// Validate flags
	if comprehensiveStatsConfig.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_comprehensive_stats_%d", time.Now().Unix())
	}

	// Load configuration
	loader := config.NewLoader(globalConfig.ConfigFile)
	cfg, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Get scrape configuration
	scrapeCfg := cfg.GetScrapeConfig()
	if !scrapeCfg.Enabled {
		fmt.Fprintf(os.Stderr, "ERROR: Scraping is disabled in configuration\n")
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

	// Create scrape client
	scrapeClient, err := createScrapeClient(scrapeCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create scrape client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Execute comprehensive statistics extraction
	return runComprehensiveStatsExtraction(ctx, scrapeClient, comprehensiveStatsConfig.Ticker, runID)
}

// runComprehensiveStatsExtraction executes comprehensive statistics extraction
func runComprehensiveStatsExtraction(ctx context.Context, client scrape.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for comprehensive stats extraction")
	}

	fmt.Printf("COMPREHENSIVE STATISTICS EXTRACTION ticker=%s\n", ticker)

	// Create a timeout context (30 seconds max)
	extractionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build URL for key-statistics endpoint
	url := buildScrapeURL(ticker, "key-statistics")
	body, meta, err := client.Fetch(extractionCtx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
		meta.Host, meta.Status, meta.Bytes, meta.Gzip)

	// Parse comprehensive statistics
	comprehensiveDTO, err := scrape.ParseComprehensiveKeyStatistics(body, ticker, "NMS")
	if err != nil {
		return fmt.Errorf("failed to parse comprehensive statistics: %w", err)
	}

	// Print comprehensive statistics summary
	printComprehensiveStatisticsSummary(comprehensiveDTO)

	return nil
}

// printComprehensiveStatisticsSummary prints a summary of comprehensive statistics
func printComprehensiveStatisticsSummary(dto *scrape.ComprehensiveKeyStatisticsDTO) {
	fmt.Printf("COMPREHENSIVE STATISTICS: symbol=%s currency=%s\n", dto.Symbol, dto.Currency)

	// Current values
	fmt.Printf("CURRENT VALUES:\n")
	if dto.Current.MarketCap != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.MarketCap.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.MarketCap.Scaled) / multiplier
		fmt.Printf("  Market Cap: %.2fB\n", actualValue/1e9)
	}
	if dto.Current.EnterpriseValue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValue.Scaled) / multiplier
		fmt.Printf("  Enterprise Value: %.2fB\n", actualValue/1e9)
	}
	if dto.Current.ForwardPE != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.ForwardPE.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.ForwardPE.Scaled) / multiplier
		fmt.Printf("  Forward P/E: %.2f\n", actualValue)
	}
	if dto.Current.TrailingPE != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TrailingPE.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TrailingPE.Scaled) / multiplier
		fmt.Printf("  Trailing P/E: %.2f\n", actualValue)
	}
	if dto.Current.PEGRatio != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PEGRatio.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PEGRatio.Scaled) / multiplier
		fmt.Printf("  PEG Ratio: %.2f\n", actualValue)
	}
	if dto.Current.PriceSales != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PriceSales.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PriceSales.Scaled) / multiplier
		fmt.Printf("  Price/Sales: %.2f\n", actualValue)
	}
	if dto.Current.PriceBook != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PriceBook.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PriceBook.Scaled) / multiplier
		fmt.Printf("  Price/Book: %.2f\n", actualValue)
	}
	if dto.Current.EnterpriseValueRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValueRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValueRevenue.Scaled) / multiplier
		fmt.Printf("  Enterprise Value/Revenue: %.2f\n", actualValue)
	}
	if dto.Current.EnterpriseValueEBITDA != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValueEBITDA.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValueEBITDA.Scaled) / multiplier
		fmt.Printf("  Enterprise Value/EBITDA: %.2f\n", actualValue)
	}

	// Additional statistics
	fmt.Printf("ADDITIONAL STATISTICS:\n")
	if dto.Additional.Beta != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.Beta.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.Beta.Scaled) / multiplier
		fmt.Printf("  Beta: %.2f\n", actualValue)
	}
	if dto.Additional.SharesOutstanding != nil {
		fmt.Printf("  Shares Outstanding: %.2fB\n", float64(*dto.Additional.SharesOutstanding)/1e9)
	}
	if dto.Additional.ProfitMargin != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ProfitMargin.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ProfitMargin.Scaled) / multiplier
		fmt.Printf("  Profit Margin: %.2f%%\n", actualValue)
	}
	if dto.Additional.OperatingMargin != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.OperatingMargin.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.OperatingMargin.Scaled) / multiplier
		fmt.Printf("  Operating Margin: %.2f%%\n", actualValue)
	}
	if dto.Additional.ReturnOnAssets != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ReturnOnAssets.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ReturnOnAssets.Scaled) / multiplier
		fmt.Printf("  Return on Assets: %.2f%%\n", actualValue)
	}
	if dto.Additional.ReturnOnEquity != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ReturnOnEquity.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ReturnOnEquity.Scaled) / multiplier
		fmt.Printf("  Return on Equity: %.2f%%\n", actualValue)
	}

	// Historical values
	if len(dto.Historical) > 0 {
		fmt.Printf("HISTORICAL VALUES:\n")
		for _, quarter := range dto.Historical {
			fmt.Printf("  %s:\n", quarter.Date)
			if quarter.MarketCap != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.MarketCap.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.MarketCap.Scaled) / multiplier
				fmt.Printf("    Market Cap: %.2fB\n", actualValue/1e9)
			}
			if quarter.ForwardPE != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.ForwardPE.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.ForwardPE.Scaled) / multiplier
				fmt.Printf("    Forward P/E: %.2f\n", actualValue)
			}
			if quarter.TrailingPE != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.TrailingPE.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.TrailingPE.Scaled) / multiplier
				fmt.Printf("    Trailing P/E: %.2f\n", actualValue)
			}
		}
	}
}
