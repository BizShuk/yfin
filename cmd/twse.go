// twse.go — `twse` cobra subcommand + `twseNameToFetcher` dispatch map wiring all 21 svc/twse endpoints (`MI_INDEX`/`STOCK_DAY`/`FMSRFK`/`MI_WEEK`/`FMTQIK`/...) behind the `twseFetcher(ctx, client, date, opts)` uniform contract. Capacity: 1 `twseConfig` + 1 `twseFetcher` type + 1 dispatch map (23 entries) + `runTwseEndpoint` RunE.
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bizshuk/yfin/svc/twse"
	"github.com/spf13/cobra"
)

// twseConfig holds CLI flags for the twse subcommand.
type twseConfig struct {
	endpoint string
	date     string
	stockNo  string
	month    string
	timeout  time.Duration
	pretty   bool
}

var twseCfg twseConfig

// twseCmd represents the `yfin twse` subcommand.
var twseCmd = &cobra.Command{
	Use:   "twse",
	Short: "Query any of the 21 TWSE endpoints (Taiwan Stock Exchange)",
	Long: `Query a Taiwan Stock Exchange (TWSE) statistical/quote endpoint via the
svc/twse package and print the raw JSON envelope to stdout.

Examples:
  yfin twse --endpoint MI_INDEX --date 20221230
  yfin twse --endpoint STOCK_DAY --date 20221230 --stock 2330
  yfin twse --endpoint FMSRFK --stock 2330 --date 2022
  yfin twse --endpoint MI_WEEK --date 20221230 --pretty`,
	RunE: runTwseEndpoint,
}

// twseFetcher is the uniform function signature used by nameToFetcher.
// All entries in twseNameToFetcher satisfy this contract: `client` is
// the injected `*twse.Client` (see buildTWSEClient), `date` is the
// primary date/period key, `opts` carries extra params (e.g. stockNo).
type twseFetcher func(ctx context.Context, client *twse.Client, date string, opts url.Values) (any, error)

// twseNameToFetcher maps an endpoint name (Registry key) to its fetcher.
// The single uniform Fetcher signature is `(ctx, client, date, opts)`;
// FMSRFK is the lone exception that needs `stockNo` between client and
// date — wrapped inline.
var twseNameToFetcher = map[string]twseFetcher{
	"MI_INDEX":      twse.FetchMI_INDEX,
	"STOCK_DAY":     twse.FetchSTOCK_DAY,
	"BWIBBU_d":      twse.FetchBWIBBU_d,
	"MI_INDEX_PLUS": twse.FetchMI_INDEX_PLUS,
	"MI_INDEX_ODD":  twse.FetchMI_INDEX_ODD,
	"MI_5MINS":      twse.FetchMI_5MINS,
	"TWTB4U":        twse.FetchTWTB4U,
	"MI_MARGN":      twse.FetchMI_MARGN,
	"T86":           twse.FetchT86,
	"MI_QFIIS":      twse.FetchMI_QFIIS,
	"BFI82U":        twse.FetchBFI82U,
	"TWT38U":        twse.FetchTWT38U,
	"TWT43U":        twse.FetchTWT43U,
	"TWT44U":        twse.FetchTWT44U,
	"BFIAUU":        twse.FetchBlockBFIAUU, //nolint:misspell // upstream naming; 4-col & 10-col were consolidated to 10-col
	"BFIAUU_STOCK":  twse.FetchBFIAUUSTOCK,
	"BFIMUU":        twse.FetchBFIMUU,
	"BFIAUU_YEAR":   twse.FetchBFIAUUYEAR,
	"FMTQIK":        twse.FetchFMTQIK,
	"STOCK_DAY_AVG": twse.FetchStockDayAvg,
	// FMSRFK has the signature FetchFMSRFK(ctx, client, stockNo, date, opts);
	// wrap it so the adapter sees a uniform (date, opts) shape.
	"FMSRFK": func(ctx context.Context, client *twse.Client, date string, opts url.Values) (any, error) {
		stockNo := opts.Get("stockNo")
		if stockNo == "" {
			return nil, fmt.Errorf("FMSRFK: --stock is required")
		}
		return twse.FetchFMSRFK(ctx, client, stockNo, date, opts)
	},
	"BFIAMU":  twse.FetchBFIAMU,
	"MI_WEEK": twse.FetchMI_WEEK,
}

func init() {
	twseCmd.Flags().StringVar(&twseCfg.endpoint, "endpoint", "", "TWSE endpoint name (e.g. MI_INDEX, STOCK_DAY, FMSRFK)")
	twseCmd.Flags().StringVar(&twseCfg.date, "date", "", "Date for the query (YYYYMMDD, or year for FMSRFK/STOCK_DAY_AVG)")
	twseCmd.Flags().StringVar(&twseCfg.stockNo, "stock", "", "Stock code (required for STOCK_DAY, STOCK_DAY_AVG, BFIAUU_STOCK, FMSRFK)")
	twseCmd.Flags().StringVar(&twseCfg.month, "month", "", "Month (YYYYMM) for monthly endpoints (BFIMUU, FMTQIK)")
	twseCmd.Flags().DurationVar(&twseCfg.timeout, "timeout", 30*time.Second, "HTTP timeout")
	twseCmd.Flags().BoolVar(&twseCfg.pretty, "pretty", false, "Pretty-print JSON output")
	_ = twseCmd.MarkFlagRequired("endpoint")
	_ = twseCmd.MarkFlagRequired("date")
}

// runTwseEndpoint is the RunE for `yfin twse`. It validates flags, builds
// query opts from the endpoint's NeedsStock/NeedsMonth markers, builds a
// shared `*twse.Client` (real or test), and dispatches the fetcher.
func runTwseEndpoint(cmd *cobra.Command, args []string) error {
	ep, ok := twse.Registry[twseCfg.endpoint]
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: unknown endpoint %q (use --endpoint MI_INDEX, STOCK_DAY, ...)\n", twseCfg.endpoint)
		return fmt.Errorf("unknown endpoint")
	}

	// Validate required inputs based on Registry metadata.
	if ep.NeedsStock && twseCfg.stockNo == "" {
		fmt.Fprintf(os.Stderr, "ERROR: endpoint %q requires --stock <code>\n", twseCfg.endpoint)
		return fmt.Errorf("missing --stock")
	}
	if ep.NeedsMonth && twseCfg.month == "" {
		fmt.Fprintf(os.Stderr, "ERROR: endpoint %q requires --month YYYYMM\n", twseCfg.endpoint)
		return fmt.Errorf("missing --month")
	}

	fetcher, ok := twseNameToFetcher[twseCfg.endpoint]
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: endpoint %q has no fetcher wired in cmd/twse.go\n", twseCfg.endpoint)
		return fmt.Errorf("no fetcher")
	}

	// Build opts from CLI flags (other than --endpoint/--date).
	opts := url.Values{}
	if twseCfg.stockNo != "" {
		opts.Set("stockNo", twseCfg.stockNo)
	}
	if twseCfg.month != "" {
		opts.Set("month", twseCfg.month)
	}

	// The HTTP transport is the injected `*twse.Client`. In production
	// it's built by `buildTWSEClient` in root.go; tests override by
	// assigning to `twseClientOverride` before invoking runTwseEndpoint.
	ctx, cancel := context.WithTimeout(context.Background(), twseCfg.timeout+5*time.Second)
	defer cancel()

	client := twseClientProvider()
	raw, err := fetcher(ctx, client, twseCfg.date, opts)
	if err != nil {
		if errors.Is(err, twse.ErrNoData) || strings.Contains(err.Error(), "no data") {
			fmt.Fprintf(os.Stderr, "INFO: TWSE returned no data for %s on %s\n", twseCfg.endpoint, twseCfg.date)
			return nil
		}
		fmt.Fprintf(os.Stderr, "ERROR: fetch failed: %v\n", err)
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	if twseCfg.pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(raw); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: encode json: %v\n", err)
		return err
	}
	return nil
}

// twseClientProvider returns the `*twse.Client` used by every twse
// command invocation. Production wiring replaces the default with
// `buildTWSEClient` (see root.go); tests use `setTwseClientForTest`.
var twseClientProvider = func() *twse.Client {
	return buildTWSEClient()
}

// setTwseClientForTest swaps the provider for the lifetime of one test.
// Returns a restore function suitable for t.Cleanup.
func setTwseClientForTest(t *testing.T, client *twse.Client) func() {
	t.Helper()
	prev := twseClientProvider
	twseClientProvider = func() *twse.Client { return client }
	return func() { twseClientProvider = prev }
}
