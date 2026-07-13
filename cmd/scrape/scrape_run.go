// scrape_run.go — `scrape` cobra subcommand 的 3 種 mode runners + 通用 helper
// (時間/字串 helper)。scrape.go 只負責 Register，scrape_format.go 負責 DTO →
// stdout formatter，本檔負責實際的 orchestration。所有 svc 呼叫都透過
// facade（ScrapeFetch / ParseComprehensiveXxx / BuildScrapeURL 等）。
package scrape

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/config/types"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/obsv"
	"github.com/spf13/cobra"
)

// scrapeConfig holds configuration for the scrape command
type scrapeConfig struct {
	Check       bool
	Ticker      string
	Endpoint    string
	Endpoints   string // Comma-separated list of endpoints for preview-json
	Preview     bool
	PreviewJSON bool
	PreviewNews bool // Preview news articles
	Force       bool
}

// newScrapeCmd returns the `scrape` cobra command.
func newScrapeCmd() *cobra.Command {
	cfg := &scrapeConfig{}
	c := &cobra.Command{
		Use:   "scrape",
		Short: "Yahoo Finance 網頁爬蟲 (Web scraping operations)",
		Long: `當 API endpoints 無法使用時，改走 Yahoo Finance 網頁爬蟲取得資料。
提供三種模式：--check（連線測試）/ --preview-json（extractor 乾跑）/ --preview-news（news parser 乾跑）。
(Web scraping operations for Yahoo Finance data.
This command provides access to scraping functionality when API endpoints are unavailable.)

範例 (Examples):
  yfin scrape --check --ticker AAPL --endpoint profile --preview
  yfin scrape --check --ticker MSFT --endpoint key-statistics --preview
  yfin scrape --preview-json --ticker AAPL --endpoints key-statistics,financials,analysis,profile
  yfin scrape --preview-news --ticker AAPL`,
		RunE: func(c *cobra.Command, args []string) error { return runScrape(c, cfg) },
	}
	c.Flags().BoolVar(&cfg.Check, "check", false, "Check scraping connectivity (no parsing)")
	c.Flags().StringVar(&cfg.Ticker, "ticker", "", "Stock symbol to scrape (e.g., AAPL)")
	c.Flags().StringVar(&cfg.Endpoint, "endpoint", "", "Endpoint to scrape (profile, key-statistics, financials, balance-sheet, cash-flow, analysis, analyst-insights, news)")
	c.Flags().StringVar(&cfg.Endpoints, "endpoints", "", "Comma-separated list of endpoints for preview-json (e.g., key-statistics,financials,analysis,profile)")
	c.Flags().BoolVar(&cfg.Preview, "preview", false, "Show preview without parsing")
	c.Flags().BoolVar(&cfg.PreviewJSON, "preview-json", false, "Preview JSON extraction")
	c.Flags().BoolVar(&cfg.PreviewNews, "preview-news", false, "Preview news articles")
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

	// Build a facade.Client that wraps both the yahoo and scrape HTTP
	// clients. facade is the single handler between cmd → svc; the CLI flag
	// overrides (--qps / --retry-max / --timeout) are applied in
	// cmd.CreateClient() on top of the ampy-config scrape settings.
	client, err := cmd.CreateClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(cmd.ExitGeneral)
	}
	_ = scrapeCfg // scrapeCfg.Enabled already validated above; facade handles client construction.

	if cfg.Check {
		return runScrapeCheck(ctx, client, cfg.Ticker, cfg.Endpoint, runID)
	}
	if cfg.PreviewJSON {
		return runScrapePreviewJSON(ctx, client, cfg.Ticker, cfg.Endpoints, runID)
	}
	if cfg.PreviewNews {
		return runScrapePreviewNews(ctx, client, cfg.Ticker, runID)
	}

	fmt.Fprintf(os.Stderr, "ERROR: Either --check, --preview-json, or --preview-news mode is required\n")
	os.Exit(cmd.ExitGeneral)
	return nil
}

// validateScrapeFlags validates scrape command flags
func validateScrapeFlags(cfg *scrapeConfig) error {
	if !cfg.Check && !cfg.PreviewJSON && !cfg.PreviewNews {
		return fmt.Errorf("either --check, --preview-json, or --preview-news flag is required")
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

	return nil
}

// runScrapeCheck runs a scrape connectivity check
func runScrapeCheck(ctx context.Context, client *facade.Client, ticker, endpoint, runID string) error {
	body, meta, err := client.ScrapeFetch(ctx, ticker, endpoint)
	if err != nil {
		return err
	}

	fmt.Printf("SCRAPE CHECK host=%s url=%s status=%d bytes=%d gzip=%t redirects=%d latency_p50≈%dms\n",
		meta.Host, meta.URL, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())
	fmt.Printf("CONTENT PREVIEW: %s\n", string(body))
	return nil
}

// runScrapePreviewNews executes the preview-news mode for testing news parser
func runScrapePreviewNews(ctx context.Context, client *facade.Client, ticker, runID string) error {
	if ticker == "" {
		return fmt.Errorf("ticker is required for preview-news mode")
	}

	fmt.Printf("PREVIEW NEWS ticker=%s\n", ticker)
	newsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	body, meta, err := client.ScrapeFetch(newsCtx, ticker, "news")
	if err != nil {
		return fmt.Errorf("failed to fetch news for %s: %v", ticker, err)
	}

	fmt.Printf("FETCH META: host=%s status=%d bytes=%d gzip=%t redirects=%d latency=%dms\n",
		meta.Host, meta.Status, meta.Bytes, meta.Gzip, meta.Redirects, meta.Duration.Milliseconds())

	now := time.Now()
	baseURL := fmt.Sprintf("https://%s", meta.Host)
	articles, stats, err := facade.ParseNews(body, baseURL, now)
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
func runScrapePreviewJSON(ctx context.Context, client *facade.Client, ticker, endpoints, runID string) error {
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
		body, meta, err := client.ScrapeFetch(endpointCtx, ticker, endpoint)
		cancel()

		if err != nil {
			fmt.Printf("ERROR: Failed to fetch: %v\n", err)
			continue
		}
		fmt.Printf("FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
			meta.Host, meta.Status, meta.Bytes, meta.Gzip)

		switch endpoint {
		case "key-statistics":
			if dto, err := facade.ParseComprehensiveKeyStatistics(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveStatisticsSummary(dto)
			}
		case "profile":
			if dto, err := facade.ParseComprehensiveProfile(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveProfileSummary(dto)
			}
		case "financials":
			if dto, err := facade.ParseComprehensiveFinancials(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printComprehensiveFinancialsSummary(dto)
			}
		case "balance-sheet", "cash-flow":
			// For balance sheet and cash flow, we need to fetch financials page to get currency
			financialsBody, financialsMeta, err := client.ScrapeFetch(ctx, ticker, "financials")
			if err != nil {
				fmt.Printf("CURRENCY FETCH ERROR: %v\n", err)
				if dto, err := facade.ParseComprehensiveFinancials(body, ticker, "NMS"); err != nil {
					fmt.Printf("PARSE ERROR: %v\n", err)
				} else {
					printComprehensiveFinancialsSummary(dto)
				}
			} else {
				fmt.Printf("CURRENCY FETCHED: host=%s status=%d bytes=%d gzip=%t\n",
					financialsMeta.Host, financialsMeta.Status, financialsMeta.Bytes, financialsMeta.Gzip)
				if dto, err := facade.ParseComprehensiveFinancialsWithCurrency(body, financialsBody, ticker, "NMS"); err != nil {
					fmt.Printf("PARSE ERROR: %v\n", err)
				} else {
					printComprehensiveFinancialsSummary(dto)
				}
			}
		case "analysis":
			if dto, err := facade.ParseAnalysis(body, ticker, "NMS"); err != nil {
				fmt.Printf("PARSE ERROR: %v\n", err)
			} else {
				printAnalysisSummary(dto)
			}
		case "analyst-insights":
			if dto, err := facade.ParseAnalystInsights(body, ticker, "NMS"); err != nil {
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
