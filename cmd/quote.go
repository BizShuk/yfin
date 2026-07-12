// quote.go — `quote` cobra subcommand that fetches a snapshot quote per
// ticker (no time-series), prints a one-line preview, optionally publishes
// to the bus, and optionally writes a local JSON file.
// Capacity: 1 `QuoteConfig` + 1 var + 1 `quoteCmd` + 1 `init()` (7 flags) +
// runQuote / validateQuoteFlags / processQuote / printQuotePreview /
// handleQuoteBusPublishing / handleQuoteLocalExport / estimateQuoteSize.
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bizshuk/yfin/svc/emit"
	"github.com/bizshuk/yfin/svc/norm"
	"github.com/bizshuk/yfin/utils/bus"
	"github.com/spf13/cobra"
)

// QuoteConfig holds configuration for the quote command
type QuoteConfig struct {
	Tickers     string
	Preview     bool
	Publish     bool
	Env         string
	TopicPrefix string
	Out         string
	OutDir      string
}

var quoteConfig QuoteConfig

// quoteCmd represents the quote command
var quoteCmd = &cobra.Command{
	Use:   "quote",
	Short: "Fetch snapshot quotes",
	Long: `Fetch snapshot quotes for one or more symbols.

Examples:
  yfin quote --tickers AAPL,MSFT,TSLA --preview
  yfin quote --tickers AAPL --publish --env prod --topic-prefix ampy`,
	RunE: runQuote,
}

func init() {
	// Quote command flags
	quoteCmd.Flags().StringVar(&quoteConfig.Tickers, "tickers", "", "Comma-separated list of symbols (e.g., AAPL,MSFT,TSLA)")
	quoteCmd.Flags().BoolVar(&quoteConfig.Preview, "preview", false, "Show preview without publishing")
	quoteCmd.Flags().BoolVar(&quoteConfig.Publish, "publish", false, "Enable bus publishing")
	quoteCmd.Flags().StringVar(&quoteConfig.Env, "env", "dev", "Environment (dev, staging, prod)")
	quoteCmd.Flags().StringVar(&quoteConfig.TopicPrefix, "topic-prefix", "ampy", "Topic prefix for bus publishing")
	quoteCmd.Flags().StringVar(&quoteConfig.Out, "out", "", "Output format (json)")
	quoteCmd.Flags().StringVar(&quoteConfig.OutDir, "out-dir", "", "Output directory")
	rootCmd.AddCommand(quoteCmd)
}

// runQuote executes the quote command
func runQuote(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateQuoteFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	// Parse tickers
	tickers := strings.Split(quoteConfig.Tickers, ",")
	for i, ticker := range tickers {
		tickers[i] = strings.TrimSpace(ticker)
	}

	// Create client
	client, err := createClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Create bus if publishing
	var busInstance *bus.Bus
	var busConfig *bus.Config
	if quoteConfig.Publish || quoteConfig.Preview {
		busConfig = createBusConfig(quoteConfig.Env, quoteConfig.TopicPrefix)
		busInstance, err = bus.NewBus(busConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to create bus: %v\n", err)
			os.Exit(ExitGeneral)
		}
		defer busInstance.Close(context.Background())
	}

	// Process quotes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	successCount := 0
	for _, ticker := range tickers {
		if err := processQuote(ctx, client, ticker, runID, busInstance, busConfig); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to process quote for %s: %v\n", ticker, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: No quotes processed successfully\n")
		os.Exit(ExitGeneral)
	}

	fmt.Printf("Successfully processed %d/%d quotes\n", successCount, len(tickers))
	return nil
}

// validateQuoteFlags validates quote command flags
func validateQuoteFlags() error {
	if quoteConfig.Tickers == "" {
		return fmt.Errorf("--tickers is required")
	}
	if quoteConfig.Out != "" && quoteConfig.Out != "json" {
		return fmt.Errorf("--out must be 'json' for quotes")
	}
	return nil
}

// processQuote processes a single quote
func processQuote(ctx context.Context, client *cliClient, ticker string, runID string, busInstance *bus.Bus, busConfig *bus.Config) error {
	// Fetch quote via the CLI helper (svc/yahoo + internal/norm).
	quote, err := fetchQuoteNorm(ctx, client.Yahoo, ticker, runID)
	if err != nil {
		return err
	}

	// Print preview
	printQuotePreview(quote)

	// Handle bus publishing
	if busInstance != nil {
		if err := handleQuoteBusPublishing(ctx, quote, busInstance, busConfig, runID, quoteConfig.Preview); err != nil {
			return fmt.Errorf("bus publishing failed: %v", err)
		}
	}

	// Handle local export
	if quoteConfig.Out != "" && quoteConfig.OutDir != "" {
		if err := handleQuoteLocalExport(quote, ticker, quoteConfig.Out, quoteConfig.OutDir); err != nil {
			return fmt.Errorf("local export failed: %v", err)
		}
	}

	return nil
}

// printQuotePreview prints the quote preview according to specification
func printQuotePreview(quote *norm.NormalizedQuote) {
	price := "N/A"
	if quote.RegularMarketPrice != nil {
		price = fmt.Sprintf("%.4f", norm.FromScaledDecimal(*quote.RegularMarketPrice))
	}

	high := "N/A"
	if quote.RegularMarketHigh != nil {
		high = fmt.Sprintf("%.4f", norm.FromScaledDecimal(*quote.RegularMarketHigh))
	}

	low := "N/A"
	if quote.RegularMarketLow != nil {
		low = fmt.Sprintf("%.4f", norm.FromScaledDecimal(*quote.RegularMarketLow))
	}

	fmt.Printf("SYMBOL %s quote  price=%s %s  high=%s  low=%s  venue=%s\n",
		quote.Security.Symbol, price, quote.CurrencyCode, high, low, quote.Venue)
}

// handleQuoteBusPublishing handles bus publishing for quotes
func handleQuoteBusPublishing(ctx context.Context, quote *norm.NormalizedQuote, busInstance *bus.Bus, busConfig *bus.Config, runID string, preview bool) error {
	// Emit to ampy-proto format
	ampyQuote, err := emit.EmitQuote(quote)
	if err != nil {
		return fmt.Errorf("failed to emit quote: %v", err)
	}

	// Create bus message
	busMessage := &bus.QuoteMessage{
		Quote: ampyQuote,
		Key: &bus.Key{
			Symbol: quote.Security.Symbol,
			MIC:    quote.Security.MIC,
		},
		RunID: runID,
		Env:   busConfig.Env,
	}

	if preview {
		// Estimate payload size
		payloadSize := estimateQuoteSize(ampyQuote)
		previewSummary, err := busInstance.PreviewQuote(busMessage, payloadSize)
		if err != nil {
			return fmt.Errorf("failed to generate preview: %v", err)
		}
		bus.PrintPreview(previewSummary)
	} else {
		// Actually publish
		if err := busInstance.PublishQuote(ctx, busMessage); err != nil {
			return fmt.Errorf("failed to publish quote: %v", err)
		}
		fmt.Printf("Published quote to bus\n")
	}

	return nil
}

// handleQuoteLocalExport handles local export for quotes
func handleQuoteLocalExport(quote *norm.NormalizedQuote, ticker, outFormat, outDir string) error {
	// Create output directory
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s_snapshot_quote.%s", ticker, outFormat)
	filePath := filepath.Join(outDir, "quotes", filename)

	// Create quotes subdirectory
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to create quotes directory: %v", err)
	}

	// Write file
	switch outFormat {
	case "json":
		return writeJSONFile(filePath, quote)
	default:
		return fmt.Errorf("unsupported output format: %s", outFormat)
	}
}

// estimateQuoteSize estimates the size of a quote payload
func estimateQuoteSize(quote interface{}) int {
	// Quote messages are typically small
	return 150
}
