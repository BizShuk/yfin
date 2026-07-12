// fundamentals.go — `fundamentals` cobra subcommand (Yahoo Finance quarterly
// fundamentals snapshot). Surfaces ExitPaidFeature when Yahoo returns a
// 401-class error so callers see code 2 instead of a generic failure.
// Capacity: 1 `FundamentalsConfig` + 1 var + 1 `fundamentalsCmd` +
// 1 `init()` (2 flags) + `runFundamentals` / `validateFundamentalsFlags` /
// `processFundamentals` / `printFundamentalsPreview` / `isPaidFeatureError`.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bizshuk/yfin/svc/norm"
	"github.com/spf13/cobra"
)

// FundamentalsConfig holds configuration for the fundamentals command
type FundamentalsConfig struct {
	Ticker  string
	Preview bool
}

var fundConfig FundamentalsConfig

// fundamentalsCmd represents the fundamentals command
var fundamentalsCmd = &cobra.Command{
	Use:   "fundamentals",
	Short: "Fetch fundamentals (requires paid subscription)",
	Long: `Fetch fundamentals data for a symbol.
Note: This endpoint requires Yahoo Finance paid subscription.

Examples:
  yfin fundamentals --ticker AAPL --preview`,
	RunE: runFundamentals,
}

func init() {
	fundamentalsCmd.Flags().StringVar(&fundConfig.Ticker, "ticker", "", "Stock symbol to fetch (e.g., AAPL)")
	fundamentalsCmd.Flags().BoolVar(&fundConfig.Preview, "preview", false, "Show preview")
	rootCmd.AddCommand(fundamentalsCmd)
}

// runFundamentals executes the fundamentals command
func runFundamentals(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateFundamentalsFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Generate run ID if not provided
	runID := globalConfig.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	// Create client
	client, err := createClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(ExitGeneral)
	}

	// Process fundamentals
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := processFundamentals(ctx, client, fundConfig.Ticker, runID); err != nil {
		// Check if it's a paid feature error
		if isPaidFeatureError(err) {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(ExitPaidFeature)
		}
		fmt.Fprintf(os.Stderr, "ERROR: Failed to process fundamentals for %s: %v\n", fundConfig.Ticker, err)
		os.Exit(ExitGeneral)
	}

	return nil
}

// validateFundamentalsFlags validates fundamentals command flags
func validateFundamentalsFlags() error {
	if fundConfig.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}
	return nil
}

// processFundamentals processes fundamentals
func processFundamentals(ctx context.Context, client *cliClient, ticker string, runID string) error {
	// Fetch fundamentals via the CLI helper (svc/yahoo + internal/norm).
	// 401-class errors surface with "paid subscription" in the message —
	// isPaidFeatureError() up in runFundamentals matches on that.
	fundamentals, err := fetchFundamentalsNorm(ctx, client.Yahoo, ticker, runID)
	if err != nil {
		return err
	}

	// Print preview
	printFundamentalsPreview(fundamentals)

	return nil
}

// printFundamentalsPreview prints the fundamentals preview
func printFundamentalsPreview(fundamentals *norm.NormalizedFundamentalsSnapshot) {
	fmt.Printf("SYMBOL %s fundamentals  lines=%d  source=%s\n",
		fundamentals.Security.Symbol, len(fundamentals.Lines), fundamentals.Source)

	// Show first few lines
	for i, line := range fundamentals.Lines {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s: %.2f %s\n", line.Key, float64(line.Value.Scaled)/float64(line.Value.Scale), line.CurrencyCode)
	}
}

// isPaidFeatureError checks if an error indicates a paid feature is required
func isPaidFeatureError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "paid subscription") || strings.Contains(errStr, "401") || strings.Contains(errStr, "Unauthorized")
}
