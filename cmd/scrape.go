// scrape.go — `scrape` cobra subcommand orchestrating the four scrape modes
// (`--check` connectivity test / `--preview-json` extractor dry-run /
// `--preview-news` news parser dry-run / `--preview-proto` full proto
// emission dry-run). This file owns the ScrapeConfig, the scrapeCmd
// registration, flag binding, validation, runScrape dispatch, and the
// per-mode runners. DTO → stdout rendering lives in scrape_format.go.
// Capacity: 1 `ScrapeConfig` + 1 var + 1 `scrapeCmd` + 1 `init()` (8 flags)
// + runScrape + validateScrapeFlags + createScrapeClient +
// runScrapeCheck + runScrapePreviewJSON + runScrapePreviewNews +
// runScrapePreviewProto + buildScrapeURL + formatRelativeTime + truncateString.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/svc/emit"
	"github.com/bizshuk/yfin/svc/scrape"
	"github.com/bizshuk/yfin/utils/obsv"
	"github.com/spf13/cobra"
)

// ScrapeConfig holds configuration for the scrape command
type ScrapeConfig struct {
	Check        bool
	Ticker       string
	Endpoint     string
	Endpoints    string // Comma-separated list of endpoints for preview-json
	Preview      bool
	PreviewJSON  bool
	PreviewNews  bool // Preview news articles without emitting proto
	PreviewProto bool // Preview proto summaries without full output
	Force        bool
}

var scrapeConfig ScrapeConfig

// scrapeCmd represents the scrape command
var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Web scraping operations",
	Long: `Web scraping operations for Yahoo Finance data.
This command provides access to scraping functionality when API endpoints are unavailable.

Examples:
  yfin scrape --check --ticker AAPL --endpoint profile --preview
  yfin scrape --check --ticker MSFT --endpoint key-statistics --preview
  yfin scrape --preview-json --ticker AAPL --endpoints key-statistics,financials,analysis,profile
  yfin scrape --preview-news --ticker AAPL
  yfin scrape --preview-proto --ticker AAPL --endpoints financials,analysis,profile,news`,
	RunE: runScrape,
}

func init() {
	// Scrape command flags
	scrapeCmd.Flags().BoolVar(&scrapeConfig.Check, "check", false, "Check scraping connectivity (no parsing)")
	scrapeCmd.Flags().StringVar(&scrapeConfig.Ticker, "ticker", "", "Stock symbol to scrape (e.g., AAPL)")
	scrapeCmd.Flags().StringVar(&scrapeConfig.Endpoint, "endpoint", "", "Endpoint to scrape (profile, key-statistics, financials, balance-sheet, cash-flow, analysis, analyst-insights, news)")
	scrapeCmd.Flags().StringVar(&scrapeConfig.Endpoints, "endpoints", "", "Comma-separated list of endpoints for preview-json (e.g., key-statistics,financials,analysis,profile)")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.Preview, "preview", false, "Show preview without parsing")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.PreviewJSON, "preview-json", false, "Preview JSON extraction without emitting proto")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.PreviewNews, "preview-news", false, "Preview news articles without emitting proto")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.PreviewProto, "preview-proto", false, "Preview proto summaries with counts, periods, and metadata")
	scrapeCmd.Flags().BoolVar(&scrapeConfig.Force, "force", false, "Force scraping even if API is available")
	rootCmd.AddCommand(scrapeCmd)
}

// runScrape executes the scrape command
func runScrape(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateScrapeFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_scrape_%d", time.Now().Unix())
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

	// Execute scrape check
	if scrapeConfig.Check {
		return runScrapeCheck(ctx, scrapeClient, scrapeConfig.Ticker, scrapeConfig.Endpoint, runID)
	}

	// Execute preview-json mode
	if scrapeConfig.PreviewJSON {
		return runScrapePreviewJSON(ctx, scrapeClient, scrapeConfig.Ticker, scrapeConfig.Endpoints, runID)
	}

	// Execute preview-news mode
	if scrapeConfig.PreviewNews {
		return runScrapePreviewNews(ctx, scrapeClient, scrapeConfig.Ticker, runID)
	}

	// Execute preview-proto mode
	if scrapeConfig.PreviewProto {
		return runScrapePreviewProto(ctx, scrapeClient, scrapeConfig.Ticker, scrapeConfig.Endpoints, runID)
	}

	fmt.Fprintf(os.Stderr, "ERROR: Either --check, --preview-json, --preview-news, or --preview-proto mode is required\n")
	os.Exit(ExitGeneral)
	return nil
}

// validateScrapeFlags validates scrape command flags
func validateScrapeFlags() error {
	// Check that either --check, --preview-json, --preview-news, or --preview-proto is specified
	if !scrapeConfig.Check && !scrapeConfig.PreviewJSON && !scrapeConfig.PreviewNews && !scrapeConfig.PreviewProto {
		return fmt.Errorf("either --check, --preview-json, --preview-news, or --preview-proto flag is required")
	}

	// All modes require ticker
	if scrapeConfig.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}

	// Check mode requires endpoint
	if scrapeConfig.Check {
		if scrapeConfig.Endpoint == "" {
			return fmt.Errorf("--endpoint is required for --check mode")
		}

		// Validate endpoint
		validEndpoints := []string{"profile", "key-statistics", "financials", "balance-sheet", "cash-flow", "analysis", "analyst-insights", "news"}
		valid := false
		for _, ep := range validEndpoints {
			if scrapeConfig.Endpoint == ep {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("--endpoint must be one of: %v", validEndpoints)
		}
	}

	// Preview-json mode requires endpoints
	if scrapeConfig.PreviewJSON {
		if scrapeConfig.Endpoints == "" {
			return fmt.Errorf("--endpoints is required for --preview-json mode")
		}

		// Validate endpoints
		endpointList := strings.Split(scrapeConfig.Endpoints, ",")
		validEndpoints := []string{"profile", "key-statistics", "financials", "balance-sheet", "cash-flow", "analysis", "analyst-insights", "news"}
		for _, ep := range endpointList {
			ep = strings.TrimSpace(ep)
			if ep == "" {
				continue
			}
			valid := false
			for _, validEp := range validEndpoints {
				if ep == validEp {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid endpoint '%s' in --endpoints", ep)
			}
		}
	}

	// Preview-proto mode requires endpoints
	if scrapeConfig.PreviewProto {
		if scrapeConfig.Endpoints == "" {
			return fmt.Errorf("--endpoints is required for --preview-proto mode")
		}

		// Validate endpoints
		endpointList := strings.Split(scrapeConfig.Endpoints, ",")
		validEndpoints := []string{"profile", "key-statistics", "financials", "balance-sheet", "cash-flow", "analysis", "analyst-insights", "news"}
		for _, ep := range endpointList {
			ep = strings.TrimSpace(ep)
			if ep == "" {
				continue
			}
			valid := false
			for _, validEp := range validEndpoints {
				if ep == validEp {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid endpoint '%s' in --endpoints", ep)
			}
		}
	}

	return nil
}

// createScrapeClient creates a scrape client with configuration
func createScrapeClient(cfg *config.ScrapeConfig) (scrape.Client, error) {
	// Convert config to scrape.Config
	scrapeCfg := &scrape.Config{
		Enabled:   cfg.Enabled,
		UserAgent: cfg.UserAgent,
		TimeoutMs: cfg.TimeoutMs,
		QPS:       cfg.QPS,
		Burst:     cfg.Burst,
		Retry: scrape.RetryConfig{
			Attempts:   cfg.Retry.Attempts,
			BaseMs:     cfg.Retry.BaseMs,
			MaxDelayMs: cfg.Retry.MaxDelayMs,
		},
		RobotsPolicy: cfg.RobotsPolicy,
		CacheTTLMs:   cfg.CacheTTLMs,
		Endpoints: scrape.EndpointConfig{
			KeyStatistics: cfg.Endpoints.KeyStatistics,
			Financials:    cfg.Endpoints.Financials,
			Analysis:      cfg.Endpoints.Analysis,
			Profile:       cfg.Endpoints.Profile,
			News:          cfg.Endpoints.News,
		},
	}

	// Create scrape client
	return scrape.NewClient(scrapeCfg, nil)
}

// runScrapeCheck runs a scrape connectivity check
func runScrapeCheck(ctx context.Context, client scrape.Client, ticker, endpoint, runID string) error {
	// Build URL for the endpoint
	url := buildScrapeURL(ticker, endpoint)

	// Fetch the page
	body, meta, err := client.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %v", url, err)
	}

	// Print results
	fmt.Printf("SCRAPE CHECK host=%s url=%s status=%d bytes=%d gzip=%t redirects=%d latency_p50≈%dms\n",
		meta.Host,
		meta.URL,
		meta.Status,
		meta.Bytes,
		meta.Gzip,
		meta.Redirects,
		meta.Duration.Milliseconds())

	// Show the full content (no truncation)
	fmt.Printf("CONTENT PREVIEW: %s\n", string(body))

	return nil
}

// runScrapePreviewNews executes the preview-news mode for testing news parser
func runScrapePreviewNews(ctx context.Context, client scrape.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for preview-news mode")
	}

	fmt.Printf("PREVIEW NEWS ticker=%s\n", ticker)

	// Create a timeout context (30 seconds max)
	newsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Build URL and fetch
	url := buildScrapeURL(ticker, "news")
	body, meta, err := client.Fetch(newsCtx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch news for %s: %v", ticker, err)
	}

	fmt.Printf("FETCH META: host=%s status=%d bytes=%d gzip=%t redirects=%d latency=%dms\n",
		meta.Host, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())

	// Parse news
	now := time.Now()
	baseURL := fmt.Sprintf("https://%s", meta.Host)
	articles, stats, err := scrape.ParseNews(body, baseURL, now)
	if err != nil {
		return fmt.Errorf("failed to parse news: %v", err)
	}

	// Print summary
	fmt.Printf("\n%s news: found=%d deduped=%d returned=%d as_of=%s\n",
		ticker, stats.TotalFound, stats.Deduped, stats.TotalReturned, stats.AsOf.Format(time.RFC3339))

	if stats.NextPageHint != "" {
		fmt.Printf("Next page hint: %s\n", stats.NextPageHint)
	}

	// Print articles in table format
	if len(articles) > 0 {
		fmt.Printf("\nARTICLES:\n")
		for i, article := range articles {
			timeStr := "unknown"
			if article.PublishedAt != nil {
				timeStr = formatRelativeTime(*article.PublishedAt, now)
			}

			// Truncate title for display
			title := article.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}

			fmt.Printf("%2d) %-8s | %-15s | %s\n", i+1, timeStr, truncateString(article.Source, 15), title)

			// Show related tickers if any
			if len(article.RelatedTickers) > 0 {
				fmt.Printf("    Tickers: %s\n", strings.Join(article.RelatedTickers, ", "))
			}
		}
	}

	return nil
}

// formatRelativeTime formats a time relative to now for display
func formatRelativeTime(t, now time.Time) string {
	diff := now.Sub(t)

	if diff < time.Minute {
		return "now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else {
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// runScrapePreviewJSON executes the preview-json mode for testing extractors
func runScrapePreviewJSON(ctx context.Context, client scrape.Client, ticker, endpoints, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for preview-json mode")
	}

	if endpoints == "" {
		return fmt.Errorf("endpoints is required for preview-json mode")
	}

	// Parse endpoints
	endpointList := strings.Split(endpoints, ",")
	for i, ep := range endpointList {
		endpointList[i] = strings.TrimSpace(ep)
	}

	fmt.Printf("PREVIEW JSON EXTRACTION ticker=%s endpoints=%s\n", ticker, endpoints)

	// Process each endpoint with individual timeouts
	for _, endpoint := range endpointList {
		if endpoint == "" {
			continue
		}

		fmt.Printf("\n--- %s ---\n", strings.ToUpper(endpoint))

		// Create a timeout context for each endpoint (15 seconds max)
		endpointCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

		// Build URL and fetch
		url := buildScrapeURL(ticker, endpoint)
		body, meta, err := client.Fetch(endpointCtx, url)
		cancel() // Always cancel the context

		if err != nil {
			fmt.Printf("ERROR: Failed to fetch %s: %v\n", url, err)
			continue
		}

		fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
			meta.Host, meta.Status, meta.Bytes, meta.Gzip)

		// Parse based on endpoint type
		switch endpoint {
		case "key-statistics":
			if dto, err := scrape.ParseComprehensiveKeyStatistics(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveStatisticsSummary(dto)
			}
		case "profile":
			if dto, err := scrape.ParseComprehensiveProfile(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveProfileSummary(dto)
			}
		case "financials":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveFinancialsSummary(dto)
			}
		case "balance-sheet", "cash-flow":
			// For balance sheet and cash flow, we need to fetch financials page to get currency
			financialsURL := buildScrapeURL(ticker, "financials")
			fmt.Printf("FETCHING CURRENCY: %s\n", financialsURL)

			financialsBody, financialsMeta, err := client.Fetch(ctx, financialsURL)
			if err != nil {
				fmt.Printf("CURRENCY FETCH ERROR: %v\n", err)
				// Continue with original parsing but currency will default to USD
				if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "NMS"); err != nil {
					fmt.Printf("PARSE ERROR: %v\n", err)
				} else {
					printComprehensiveFinancialsSummary(dto)
				}
			} else {
				fmt.Printf("CURRENCY FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
					financialsMeta.Host, financialsMeta.Status, financialsMeta.Bytes, financialsMeta.Gzip)

				// Parse the current endpoint (balance-sheet or cash-flow) with currency from financials
				if dto, err := scrape.ParseComprehensiveFinancialsWithCurrency(body, financialsBody, ticker, "NMS"); err != nil {
					fmt.Printf("PARSE ERROR: %v\n", err)
				} else {
					printComprehensiveFinancialsSummary(dto)
				}
			}
		case "analysis":
			if dto, err := scrape.ParseAnalysis(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printAnalysisSummary(dto)
			}
		case "analyst-insights":
			if dto, err := scrape.ParseAnalystInsights(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printAnalystInsightsSummary(dto)
			}
		default:
			fmt.Printf("UNSUPPORTED ENDPOINT: %s (only key-statistics, profile, financials, balance-sheet, cash-flow, analysis, and analyst-insights are supported)\n", endpoint)
		}
	}

	return nil
}

// buildScrapeURL builds the URL for a given ticker and endpoint
func buildScrapeURL(ticker, endpoint string) string {
	baseURL := "https://finance.yahoo.com"

	switch endpoint {
	case "profile":
		return fmt.Sprintf("%s/quote/%s/profile", baseURL, ticker)
	case "key-statistics":
		return fmt.Sprintf("%s/quote/%s/key-statistics", baseURL, ticker)
	case "financials":
		return fmt.Sprintf("%s/quote/%s/financials", baseURL, ticker)
	case "balance-sheet":
		return fmt.Sprintf("%s/quote/%s/balance-sheet", baseURL, ticker)
	case "cash-flow":
		return fmt.Sprintf("%s/quote/%s/cash-flow", baseURL, ticker)
	case "analysis":
		return fmt.Sprintf("%s/quote/%s/analysis", baseURL, ticker)
	case "analyst-insights":
		return fmt.Sprintf("%s/quote/%s/analyst-insights", baseURL, ticker)
	case "news":
		return fmt.Sprintf("%s/quote/%s/news", baseURL, ticker)
	default:
		return fmt.Sprintf("%s/quote/%s", baseURL, ticker)
	}
}

// runScrapePreviewProto executes the preview-proto mode for testing proto emission
func runScrapePreviewProto(ctx context.Context, client scrape.Client, ticker, endpoints, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for preview-proto mode")
	}

	if endpoints == "" {
		return fmt.Errorf("endpoints is required for preview-proto mode")
	}

	// Parse endpoints
	endpointList := strings.Split(endpoints, ",")
	for i, ep := range endpointList {
		endpointList[i] = strings.TrimSpace(ep)
	}

	fmt.Printf("PREVIEW PROTO EMISSION ticker=%s endpoints=%s\n", ticker, endpoints)

	// Create mapper configuration
	mapperConfig := emit.ScrapeMapperConfig{
		RunID:    runID,
		Producer: fmt.Sprintf("yfin-%s", version),
		Source:   "yfinance-go/scrape",
		TraceID:  "", // Could be extracted from context if available
	}

	// mapper := emit.NewScrapeMapper(mapperConfig) // Not used in this function

	// Process each endpoint
	for _, endpoint := range endpointList {
		if endpoint == "" {
			continue
		}

		fmt.Printf("\n--- %s ---\n", strings.ToUpper(endpoint))

		// Create a timeout context for each endpoint (15 seconds max)
		endpointCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

		// Build URL and fetch
		url := buildScrapeURL(ticker, endpoint)
		body, meta, err := client.Fetch(endpointCtx, url)
		cancel() // Always cancel the context

		if err != nil {
			fmt.Printf("ERROR: Failed to fetch %s: %v\n", url, err)
			continue
		}

		fmt.Printf("FETCH META: host=%s status=%d bytes=%d gzip=%t redirects=%d latency=%dms\n",
			meta.Host, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())

		// Parse and map based on endpoint type
		switch endpoint {
		case "financials":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				// Use the comprehensive mapping for more complete data
				if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					for _, snapshot := range snapshots {
						printFundamentalsSnapshot(snapshot)
					}
				}
			}

		case "profile":
			if dto, err := scrape.ParseComprehensiveProfile(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if result, err := emit.MapProfileDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printProfileResult(result)
				}
			}

		case "news":
			if articles, stats, err := scrape.ParseNews(body, "https://finance.yahoo.com", time.Now()); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if protoArticles, err := emit.MapNewsItems(articles, ticker, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printNewsArticles(protoArticles, stats)
				}
			}

		case "balance-sheet":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				// Balance sheet data is included in comprehensive financials
				if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					for _, snapshot := range snapshots {
						printFundamentalsSnapshot(snapshot)
					}
				}
			}

		case "cash-flow":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				// Cash flow data is included in comprehensive financials
				if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					for _, snapshot := range snapshots {
						printFundamentalsSnapshot(snapshot)
					}
				}
			}

		case "key-statistics":
			if dto, err := scrape.ParseComprehensiveKeyStatistics(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if snapshot, err := emit.MapKeyStatisticsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printFundamentalsSnapshot(snapshot)
				}
			}

		case "analysis":
			if dto, err := scrape.ParseAnalysis(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if snapshot, err := emit.MapAnalysisDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printFundamentalsSnapshot(snapshot)
				}
			}

		case "analyst-insights":
			if dto, err := scrape.ParseAnalystInsights(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				if snapshot, err := emit.MapAnalystInsightsDTO(dto, runID, mapperConfig.Producer); err != nil {
					fmt.Printf("MAPPING ERROR: %v\n", err)
				} else {
					printFundamentalsSnapshot(snapshot)
				}
			}

		default:
			fmt.Printf("PROTO MAPPING: endpoint '%s' not yet supported for proto emission\n", endpoint)
			fmt.Printf("Supported endpoints: financials, balance-sheet, cash-flow, key-statistics, analysis, analyst-insights, profile, news\n")
		}
	}

	return nil
}
