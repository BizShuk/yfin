// fundamentals_run.go — `fundamentals` cobra subcommand：擷取 Yahoo Finance
// quarterly fundamentals snapshot。401-class error 時 surface ExitPaidFeature
// (code 2)。
package fundamentals

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/model"
	"github.com/spf13/cobra"
)

// fundamentalsConfig holds configuration for the fundamentals command
type fundamentalsConfig struct {
	Ticker  string
	Preview bool
}

// newFundamentalsCmd returns the `fundamentals` cobra command.
func newFundamentalsCmd() *cobra.Command {
	cfg := &fundamentalsConfig{}
	c := &cobra.Command{
		Use:   "fundamentals",
		Short: "擷取 fundamentals 季報（需 Yahoo Finance 付費訂閱）(Fetch fundamentals — requires paid subscription)",
		Long: `擷取單一 symbol 的 fundamentals 季報資料。
注意：本 endpoint 需 Yahoo Finance 付費訂閱。
(Fetch fundamentals data for a symbol.
Note: This endpoint requires Yahoo Finance paid subscription.)

範例 (Examples):
  yfin fundamentals --ticker AAPL --preview`,
		RunE: func(c *cobra.Command, args []string) error { return runFundamentals(cfg) },
	}
	c.Flags().StringVar(&cfg.Ticker, "ticker", "", "Stock symbol to fetch (e.g., AAPL)")
	c.Flags().BoolVar(&cfg.Preview, "preview", false, "Show preview")
	return c
}

// runFundamentals executes the fundamentals command
func runFundamentals(cfg *fundamentalsConfig) error {
	if err := validateFundamentalsFlags(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	runID := cmd.Global.RunID
	if runID == "" {
		runID = fmt.Sprintf("yfin_%d", time.Now().Unix())
	}

	client, err := cmd.CreateClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create client: %v\n", err)
		os.Exit(cmd.ExitGeneral)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := processFundamentals(ctx, client, cfg.Ticker, runID); err != nil {
		if isPaidFeatureError(err) {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(cmd.ExitPaidFeature)
		}
		fmt.Fprintf(os.Stderr, "ERROR: Failed to process fundamentals for %s: %v\n", cfg.Ticker, err)
		os.Exit(cmd.ExitGeneral)
	}
	return nil
}

// validateFundamentalsFlags validates fundamentals command flags
func validateFundamentalsFlags(cfg *fundamentalsConfig) error {
	if cfg.Ticker == "" {
		return fmt.Errorf("--ticker is required")
	}
	return nil
}

// processFundamentals processes fundamentals
func processFundamentals(ctx context.Context, client *facade.Client, ticker string, runID string) error {
	fundamentals, err := client.FetchFundamentalsNorm(ctx, ticker, runID)
	if err != nil {
		return err
	}
	printFundamentalsPreview(fundamentals)
	return nil
}

// printFundamentalsPreview prints the fundamentals preview
func printFundamentalsPreview(fundamentals *model.NormalizedFundamentalsSnapshot) {
	fmt.Printf("SYMBOL %s fundamentals  lines=%d  source=%s\n",
		fundamentals.Security.Symbol, len(fundamentals.Lines), fundamentals.Source)

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
