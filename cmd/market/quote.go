// quote.go — `quote` cobra subcommand：擷取即時 quote 快照。`pull` 同樣是
// market-data 出口；兩者共用 client_json.go 的 writeJSONFile。本檔只負責
// quote 的 cobra command 構建與所有 helper。
package market

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/model"
	"github.com/spf13/cobra"
)

// quoteConfig holds configuration for the quote command
type quoteConfig struct {
	Tickers string
	Out     string
	OutDir  string
}

// newQuoteCmd returns the `quote` cobra command.
func newQuoteCmd() *cobra.Command {
	cfg := &quoteConfig{}
	c := &cobra.Command{
		Use:   "quote",
		Short: "擷取即時 quote 快照 (Fetch snapshot quotes)",
		Long: `擷取單一或多個 symbols 的即時 quote 快照。
(Fetch snapshot quotes for one or more symbols.)

範例 (Examples):
  yfin quote --tickers AAPL,MSFT,TSLA
  yfin quote --tickers AAPL --out json --out-dir ./out`,
		RunE: func(c *cobra.Command, args []string) error { return runQuote(cfg) },
	}
	// Quote command flags
	c.Flags().StringVar(&cfg.Tickers, "tickers", "", "Comma-separated list of symbols (e.g., AAPL,MSFT,TSLA)")
	c.Flags().StringVar(&cfg.Out, "out", "", "Output format (json)")
	c.Flags().StringVar(&cfg.OutDir, "out-dir", "", "Output directory")
	return c
}

// runQuote executes the quote command
func runQuote(cfg *quoteConfig) error {
	// Validate flags
	if err := validateQuoteFlags(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// Generate run ID if not provided
	runID := cmd.Global.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	// Parse tickers
	tickers := strings.Split(cfg.Tickers, ",")
	for i, ticker := range tickers {
		tickers[i] = strings.TrimSpace(ticker)
	}

	// Create client
	client, err := cmd.CreateClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(cmd.ExitGeneral)
	}

	// Process quotes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	successCount := 0
	for _, ticker := range tickers {
		if err := processQuote(ctx, client, ticker, runID, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to process quote for %s: %v\n", ticker, err)
			continue
		}
		successCount++
	}

	if successCount == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: No quotes processed successfully\n")
		os.Exit(cmd.ExitGeneral)
	}

	fmt.Printf("Successfully processed %d/%d quotes\n", successCount, len(tickers))
	return nil
}

// validateQuoteFlags validates quote command flags
func validateQuoteFlags(cfg *quoteConfig) error {
	if cfg.Tickers == "" {
		return fmt.Errorf("--tickers is required")
	}
	if cfg.Out != "" && cfg.Out != "json" {
		return fmt.Errorf("--out must be 'json' for quotes")
	}
	return nil
}

// processQuote processes a single quote
func processQuote(ctx context.Context, client *facade.Client, ticker string, runID string, cfg *quoteConfig) error {
	// Fetch quote via facade (Norm-returning variant for local JSON export).
	quote, err := client.FetchQuoteNorm(ctx, ticker, runID)
	if err != nil {
		return err
	}

	// Print preview
	printQuotePreview(quote)

	// Handle local export
	if cfg.Out != "" && cfg.OutDir != "" {
		if err := handleQuoteLocalExport(quote, ticker, cfg.Out, cfg.OutDir); err != nil {
			return fmt.Errorf("local export failed: %v", err)
		}
	}

	return nil
}

// printQuotePreview prints the quote preview according to specification
func printQuotePreview(quote *model.NormalizedQuote) {
	price := "N/A"
	if quote.RegularMarketPrice != nil {
		price = fmt.Sprintf("%.4f", model.FromScaledDecimal(*quote.RegularMarketPrice))
	}

	high := "N/A"
	if quote.RegularMarketHigh != nil {
		high = fmt.Sprintf("%.4f", model.FromScaledDecimal(*quote.RegularMarketHigh))
	}

	low := "N/A"
	if quote.RegularMarketLow != nil {
		low = fmt.Sprintf("%.4f", model.FromScaledDecimal(*quote.RegularMarketLow))
	}

	fmt.Printf("SYMBOL %s quote  price=%s %s  high=%s  low=%s  venue=%s\n",
		quote.Security.Symbol, price, quote.CurrencyCode, high, low, quote.Venue)
}

// handleQuoteLocalExport handles local export for quotes
func handleQuoteLocalExport(quote *model.NormalizedQuote, ticker, outFormat, outDir string) error {
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
