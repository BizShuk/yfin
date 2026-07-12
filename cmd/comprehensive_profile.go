// comprehensive_profile.go — `comprehensive-profile` cobra subcommand that
// fetches the profile page and prints the parsed `ComprehensiveProfileDTO`
// (company info, executives, risks) to stdout via `printComprehensiveProfileSummary`.
// DTO → stdout rendering lives in this file because no other subcommand
// consumes that DTO.
// Capacity: 1 `ComprehensiveProfileConfig` + 1 var + 1 `comprehensiveProfileCmd` +
// 1 `init()` (2 flags) + runComprehensiveProfile + runComprehensiveProfileExtraction
// + printComprehensiveProfileSummary.
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

// ComprehensiveProfileConfig holds configuration for comprehensive profile command
type ComprehensiveProfileConfig struct {
	Ticker  string
	Preview bool
}

var comprehensiveProfileConfig ComprehensiveProfileConfig

// comprehensiveProfileCmd represents the comprehensive profile command
var comprehensiveProfileCmd = &cobra.Command{
	Use:   "comprehensive-profile",
	Short: "Extract comprehensive company profile information",
	Long: `Extract comprehensive company profile information including company details,
key executives, and business summary from Yahoo Finance.

Examples:
  yfin comprehensive-profile --ticker AAPL
  yfin comprehensive-profile --ticker MSFT --preview`,
	RunE: runComprehensiveProfile,
}

func init() {
	// Comprehensive profile command flags
	comprehensiveProfileCmd.Flags().StringVar(&comprehensiveProfileConfig.Ticker, "ticker", "", "Stock symbol to analyze (e.g., AAPL)")
	comprehensiveProfileCmd.Flags().BoolVar(&comprehensiveProfileConfig.Preview, "preview", false, "Show preview of extracted data")
	rootCmd.AddCommand(comprehensiveProfileCmd)
}

// runComprehensiveProfile executes the comprehensive profile command
func runComprehensiveProfile(cmd *cobra.Command, args []string) error {
	// Validate flags
	if comprehensiveProfileConfig.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_comprehensive_profile_%d", time.Now().Unix())
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

	// Execute comprehensive profile extraction
	return runComprehensiveProfileExtraction(ctx, scrapeClient, comprehensiveProfileConfig.Ticker, runID)
}

// runComprehensiveProfileExtraction executes comprehensive profile extraction
func runComprehensiveProfileExtraction(ctx context.Context, client scrape.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for comprehensive profile extraction")
	}

	fmt.Printf("COMPREHENSIVE PROFILE EXTRACTION ticker=%s\n", ticker)

	// Create a timeout context (30 seconds max)
	extractionCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build URL for profile endpoint
	url := buildScrapeURL(ticker, "profile")
	body, meta, err := client.Fetch(extractionCtx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
		meta.Host, meta.Status, meta.Bytes, meta.Gzip)

	// Parse comprehensive profile
	comprehensiveDTO, err := scrape.ParseComprehensiveProfile(body, ticker, "NMS")
	if err != nil {
		return fmt.Errorf("failed to parse comprehensive profile: %w", err)
	}

	// Print comprehensive profile summary
	printComprehensiveProfileSummary(comprehensiveDTO)

	return nil
}

// printComprehensiveProfileSummary prints a summary of comprehensive profile
func printComprehensiveProfileSummary(dto *scrape.ComprehensiveProfileDTO) {
	fmt.Printf("COMPREHENSIVE PROFILE: symbol=%s\n", dto.Symbol)

	// Company Information
	fmt.Printf("COMPANY INFORMATION:\n")
	if dto.CompanyName != "" {
		fmt.Printf("  Company Name: %s\n", dto.CompanyName)
	}
	if dto.ShortName != "" {
		fmt.Printf("  Short Name: %s\n", dto.ShortName)
	}
	if dto.Address1 != "" {
		fmt.Printf("  Address: %s\n", dto.Address1)
	}
	if dto.City != "" && dto.State != "" {
		fmt.Printf("  City, State: %s, %s\n", dto.City, dto.State)
	}
	if dto.Zip != "" {
		fmt.Printf("  ZIP: %s\n", dto.Zip)
	}
	if dto.Country != "" {
		fmt.Printf("  Country: %s\n", dto.Country)
	}
	if dto.Phone != "" {
		fmt.Printf("  Phone: %s\n", dto.Phone)
	}
	if dto.Website != "" {
		fmt.Printf("  Website: %s\n", dto.Website)
	}
	if dto.Industry != "" {
		fmt.Printf("  Industry: %s\n", dto.Industry)
	}
	if dto.Sector != "" {
		fmt.Printf("  Sector: %s\n", dto.Sector)
	}
	if dto.FullTimeEmployees != nil {
		fmt.Printf("  Full Time Employees: %d\n", *dto.FullTimeEmployees)
	}
	if dto.BusinessSummary != "" {
		// Truncate business summary if too long
		summary := dto.BusinessSummary
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}
		fmt.Printf("  Business Summary: %s\n", summary)
	}

	// Key Executives
	if len(dto.Executives) > 0 {
		fmt.Printf("KEY EXECUTIVES:\n")
		for i, exec := range dto.Executives {
			if i >= 5 { // Limit to top 5 executives
				break
			}
			fmt.Printf("  %d. %s", i+1, exec.Name)
			if exec.Title != "" {
				fmt.Printf(" - %s", exec.Title)
			}
			if exec.YearBorn != nil {
				fmt.Printf(" (Born: %d)", *exec.YearBorn)
			}
			if exec.TotalPay != nil {
				fmt.Printf(" - Total Pay: $%.2fM", float64(*exec.TotalPay)/1e6)
			}
			fmt.Printf("\n")
		}
	}

	// Additional Information
	fmt.Printf("ADDITIONAL INFORMATION:\n")
	if dto.MaxAge != nil {
		fmt.Printf("  Max Age: %d\n", *dto.MaxAge)
	}
	if dto.AuditRisk != nil {
		fmt.Printf("  Audit Risk: %d\n", *dto.AuditRisk)
	}
	if dto.BoardRisk != nil {
		fmt.Printf("  Board Risk: %d\n", *dto.BoardRisk)
	}
	if dto.CompensationRisk != nil {
		fmt.Printf("  Compensation Risk: %d\n", *dto.CompensationRisk)
	}
	if dto.ShareHolderRightsRisk != nil {
		fmt.Printf("  Share Holder Rights Risk: %d\n", *dto.ShareHolderRightsRisk)
	}
	if dto.OverallRisk != nil {
		fmt.Printf("  Overall Risk: %d\n", *dto.OverallRisk)
	}
}
