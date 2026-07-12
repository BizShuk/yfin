// scrape_run.go — `scrape` cobra subcommand 的 4 種 mode runners + 通用 helper
// (URL builder / 連線 client builder / 時間字串 helper)。scrape.go 只負責
// Register，scrape_format.go 負責 DTO → stdout formatter，本檔負責實際的
// orchestration。
package scrape

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/svc/emit"
	"github.com/bizshuk/yfin/svc/scrape"
	"github.com/bizshuk/yfin/utils/obsv"
	"github.com/spf13/cobra"
)

// scrapeConfig holds configuration for the scrape command
type scrapeConfig struct {
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

// newScrapeCmd returns the `scrape` cobra command.
func newScrapeCmd() *cobra.Command {
	cfg := &scrapeConfig{}
	c := &cobra.Command{
		Use:   "scrape",
		Short: "Yahoo Finance 網頁爬蟲 (Web scraping operations)",
		Long: `當 API endpoints 無法使用時，改走 Yahoo Finance 網頁爬蟲取得資料。
提供四種模式：--check（連線測試）/ --preview-json（extractor 乾跑）/ --preview-news（news parser 乾跑）/ --preview-proto（proto 完整輸出乾跑）。
(Web scraping operations for Yahoo Finance data.
This command provides access to scraping functionality when API endpoints are unavailable.)

範例 (Examples):
  yfin scrape --check --ticker AAPL --endpoint profile --preview
  yfin scrape --check --ticker MSFT --endpoint key-statistics --preview
  yfin scrape --preview-json --ticker AAPL --endpoints key-statistics,financials,analysis,profile
  yfin scrape --preview-news --ticker AAPL
  yfin scrape --preview-proto --ticker AAPL --endpoints financials,analysis,profile,news`,
		RunE: func(c *cobra.Command, args []string) error { return runScrape(c, cfg) },
	}
	c.Flags().BoolVar(&cfg.Check, "check", false, "Check scraping connectivity (no parsing)")
	c.Flags().StringVar(&cfg.Ticker, "ticker", "", "Stock symbol to scrape (e.g., AAPL)")
	c.Flags().StringVar(&cfg.Endpoint, "endpoint", "", "Endpoint to scrape (profile, key-statistics, financials, balance-sheet, cash-flow, analysis, analyst-insights, news)")
	c.Flags().StringVar(&cfg.Endpoints, "endpoints", "", "Comma-separated list of endpoints for preview-json (e.g., key-statistics,financials,analysis,profile)")
	c.Flags().BoolVar(&cfg.Preview, "preview", false, "Show preview without parsing")
	c.Flags().BoolVar(&cfg.PreviewJSON, "preview-json", false, "Preview JSON extraction without emitting proto")
	c.Flags().BoolVar(&cfg.PreviewNews, "preview-news", false, "Preview news articles without emitting proto")
	c.Flags().BoolVar(&cfg.PreviewProto, "preview-proto", false, "Preview proto summaries with counts, periods, and metadata")
	c.Flags().BoolVar(&cfg.Force, "force", false, "Force scraping even if API is available")
	return c
}

// runScrape executes the scrape command
func runScrape(cobraCmd *cobra.Command, cfg *scrapeConfig) error {
	if err := validateScrapeFlags(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	runID := cmd.Global.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_scrape_%d", time.Now().Unix())
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

	scrapeClient, err := createScrapeClient(scrapeCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create scrape client: %v\n", err)
		os.Exit(cmd.ExitGeneral)
	}

	if cfg.Check {
		return runScrapeCheck(ctx, scrapeClient, cfg.Ticker, cfg.Endpoint, runID)
	}
	if cfg.PreviewJSON {
		return runScrapePreviewJSON(ctx, scrapeClient, cfg.Ticker, cfg.Endpoints, runID)
	}
	if cfg.PreviewNews {
		return runScrapePreviewNews(ctx, scrapeClient, cfg.Ticker, runID)
	}
	if cfg.PreviewProto {
		return runScrapePreviewProto(ctx, scrapeClient, cfg.Ticker, cfg.Endpoints, runID)
	}

	fmt.Fprintf(os.Stderr, "ERROR: Either --check, --preview-json, --preview-news, or --preview-proto mode is required\n")
	os.Exit(cmd.ExitGeneral)
	return nil
}

// validateScrapeFlags validates scrape command flags
func validateScrapeFlags(cfg *scrapeConfig) error {
	if !cfg.Check && !cfg.PreviewJSON && !cfg.PreviewNews && !cfg.PreviewProto {
		return fmt.Errorf("either --check, --preview-json, --preview-news, or --preview-proto flag is required")
	}
	if cfg.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}

	validEndpoints := []string{"profile", "key-statistics", "financials", "balance-sheet", "cash-flow", "analysis", "analyst-insights", "news"}

	if cfg.Check {
		if cfg.Endpoint == "" {
			return fmt.Errorf("--endpoint is required for --check mode")
		}
		valid := false
		for _, ep := range validEndpoints {
			if cfg.Endpoint == ep {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("--endpoint must be one of: %v", validEndpoints)
		}
	}

	if cfg.PreviewJSON {
		if cfg.Endpoints == "" {
			return fmt.Errorf("--endpoints is required for --preview-json mode")
		}
		endpointList := strings.Split(cfg.Endpoints, ",")
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

	if cfg.PreviewProto {
		if cfg.Endpoints == "" {
			return fmt.Errorf("--endpoints is required for --preview-proto mode")
		}
		endpointList := strings.Split(cfg.Endpoints, ",")
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
	return scrape.NewClient(scrapeCfg, nil)
}

// runScrapeCheck runs a scrape connectivity check
func runScrapeCheck(ctx context.Context, client scrape.Client, ticker, endpoint, runID string) error {
	url := buildScrapeURL(ticker, endpoint)
	body, meta, err := client.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %v", url, err)
	}

	fmt.Printf("SCRAPE CHECK host=%s url=%s status=%d bytes=%d gzip=%t redirects=%d latency_p50≈%dms\n",
		meta.Host, meta.URL, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())
	fmt.Printf("CONTENT PREVIEW: %s\n", string(body))
	return nil
}

// runScrapePreviewNews executes the preview-news mode for testing news parser
func runScrapePreviewNews(ctx context.Context, client scrape.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for preview-news mode")
	}

	fmt.Printf("PREVIEW NEWS ticker=%s\n", ticker)
	newsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := buildScrapeURL(ticker, "news")
	body, meta, err := client.Fetch(newsCtx, url)
	if err != nil {
		return fmt.Errorf("failed to fetch news for %s: %v", ticker, err)
	}

	fmt.Printf("FETCH META: host=%s status=%d bytes=%d gzip=%t redirects=%d latency=%dms\n",
		meta.Host, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())

	now := time.Now()
	baseURL := fmt.Sprintf("https://%s", meta.Host)
	articles, stats, err := scrape.ParseNews(body, baseURL, now)
	if err != nil {
		return fmt.Errorf("failed to parse news: %v", err)
	}

	fmt.Printf("\n%s news: found=%d deduped=%d returned=%d as_of=%s\n",
		ticker, stats.TotalFound, stats.Deduped, stats.TotalReturned, stats.AsOf.Format(time.RFC3339))

	if stats.NextPageHint != "" {
		fmt.Printf("Next page hint: %s\n", stats.NextPageHint)
	}

	if len(articles) > 0 {
		fmt.Printf("\nARTICLES:\n")
		for i, article := range articles {
			timeStr := "unknown"
			if article.PublishedAt != nil {
				timeStr = formatRelativeTime(*article.PublishedAt, now)
			}
			title := article.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}
			fmt.Printf("%2d) %-8s | %-15s | %s\n", i+1, timeStr, truncateString(article.Source, 15), title)
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

	endpointList := strings.Split(endpoints, ",")
	for i, ep := range endpointList {
		endpointList[i] = strings.TrimSpace(ep)
	}

	fmt.Printf("PREVIEW JSON EXTRACTION ticker=%s endpoints=%s\n", ticker, endpoints)

	for _, endpoint := range endpointList {
		if endpoint == "" {
			continue
		}
		fmt.Printf("\n--- %s ---\n", strings.ToUpper(endpoint))

		endpointCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		url := buildScrapeURL(ticker, endpoint)
		body, meta, err := client.Fetch(endpointCtx, url)
		cancel()

		if err != nil {
			fmt.Printf("ERROR: Failed to fetch %s: %v\n", url, err)
			continue
		}
		fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
			meta.Host, meta.Status, meta.Bytes, meta.Gzip)

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
			financialsURL := buildScrapeURL(ticker, "financials")
			fmt.Printf("FETCHING CURRENCY: %s\n", financialsURL)
			financialsBody, financialsMeta, err := client.Fetch(ctx, financialsURL)
			if err != nil {
				fmt.Printf("CURRENCY FETCH ERROR: %v\n", err)
				if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "NMS"); err != nil {
					fmt.Printf("PARSE ERROR: %v\n", err)
				} else {
					printComprehensiveFinancialsSummary(dto)
				}
			} else {
				fmt.Printf("CURRENCY FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
					financialsMeta.Host, financialsMeta.Status, financialsMeta.Bytes, financialsMeta.Gzip)
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

	endpointList := strings.Split(endpoints, ",")
	for i, ep := range endpointList {
		endpointList[i] = strings.TrimSpace(ep)
	}

	fmt.Printf("PREVIEW PROTO EMISSION ticker=%s endpoints=%s\n", ticker, endpoints)

	mapperConfig := emit.ScrapeMapperConfig{
		RunID:    runID,
		Producer: fmt.Sprintf("yfin-%s", cmd.Version),
		Source:   "yfinance-go/scrape",
		TraceID:  "",
	}

	for _, endpoint := range endpointList {
		if endpoint == "" {
			continue
		}
		fmt.Printf("\n--- %s ---\n", strings.ToUpper(endpoint))

		endpointCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		url := buildScrapeURL(ticker, endpoint)
		body, meta, err := client.Fetch(endpointCtx, url)
		cancel()

		if err != nil {
			fmt.Printf("ERROR: Failed to fetch %s: %v\n", url, err)
			continue
		}
		fmt.Printf("FETCH META: host=%s status=%d bytes=%d gzip=%t redirects=%d latency=%dms\n",
			meta.Host, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())

		switch endpoint {
		case "financials":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
				fmt.Printf("MAPPING ERROR: %v\n", err)
			} else {
				for _, snapshot := range snapshots {
					printFundamentalsSnapshot(snapshot)
				}
			}
		case "profile":
			if dto, err := scrape.ParseComprehensiveProfile(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else if result, err := emit.MapProfileDTO(dto, runID, mapperConfig.Producer); err != nil {
				fmt.Printf("MAPPING ERROR: %v\n", err)
			} else {
				printProfileResult(result)
			}
		case "news":
			if articles, stats, err := scrape.ParseNews(body, "https://finance.yahoo.com", time.Now()); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else if protoArticles, err := emit.MapNewsItems(articles, ticker, runID, mapperConfig.Producer); err != nil {
				fmt.Printf("MAPPING ERROR: %v\n", err)
			} else {
				printNewsArticles(protoArticles, stats)
			}
		case "balance-sheet":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
				fmt.Printf("MAPPING ERROR: %v\n", err)
			} else {
				for _, snapshot := range snapshots {
					printFundamentalsSnapshot(snapshot)
				}
			}
		case "cash-flow":
			if dto, err := scrape.ParseComprehensiveFinancials(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else if snapshots, err := emit.MapComprehensiveFinancialsDTO(dto, runID, mapperConfig.Producer); err != nil {
				fmt.Printf("MAPPING ERROR: %v\n", err)
			} else {
				for _, snapshot := range snapshots {
					printFundamentalsSnapshot(snapshot)
				}
			}
		case "key-statistics":
			if dto, err := scrape.ParseComprehensiveKeyStatistics(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else if snapshot, err := emit.MapKeyStatisticsDTO(dto, runID, mapperConfig.Producer); err != nil {
				fmt.Printf("MAPPING ERROR: %v\n", err)
			} else {
				printFundamentalsSnapshot(snapshot)
			}
		case "analysis":
			if dto, err := scrape.ParseAnalysis(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else if snapshot, err := emit.MapAnalysisDTO(dto, runID, mapperConfig.Producer); err != nil {
				fmt.Printf("MAPPING ERROR: %v\n", err)
			} else {
				printFundamentalsSnapshot(snapshot)
			}
		case "analyst-insights":
			if dto, err := scrape.ParseAnalystInsights(body, ticker, "XNAS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else if snapshot, err := emit.MapAnalystInsightsDTO(dto, runID, mapperConfig.Producer); err != nil {
				fmt.Printf("MAPPING ERROR: %v\n", err)
			} else {
				printFundamentalsSnapshot(snapshot)
			}
		default:
			fmt.Printf("PROTO MAPPING: endpoint '%s' not yet supported for proto emission\n", endpoint)
			fmt.Printf("Supported endpoints: financials, balance-sheet, cash-flow, key-statistics, analysis, analyst-insights, profile, news\n")
		}
	}
	return nil
}
