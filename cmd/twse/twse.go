// twse.go — `twse` cobra subcommand + `twseNameToFetcher` dispatch map wiring all 21 svc/twse endpoints (`MI_INDEX`/`STOCK_DAY`/`FMSRFK`/`MI_WEEK`/`FMTQIK`/...) behind the uniform `twseFetcher(ctx, client, date, opts)` contract. HTTP transport builder lives in client.go.
//
// Capacity: 1 `Register` + 1 `twseConfig` + 1 `twseCmd` var (kept exported-style for test access via `twseCfg` and `twseCmd`) + 1 `twseFetcher` type + 1 dispatch map (23 entries) + `runTwseEndpoint` RunE.
package twse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	svcTWSE "github.com/bizshuk/yfin/svc/twse"
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

// twseCmd is kept as a package-level var so tests can introspect and invoke
// `runTwseEndpoint(cmd, ...)` directly. Register attaches it to rootCmd.
var twseCmd = &cobra.Command{
	Use:   "twse",
	Short: "查詢任一 21 個 TWSE endpoints（臺灣證券交易所）(Query any of the 21 TWSE endpoints — Taiwan Stock Exchange)",
	Long: `走 svc/twse 查詢臺灣證券交易所（Taiwan Stock Exchange）統計/報價 endpoint，輸出原始 JSON envelope 到 stdout。
(Query a Taiwan Stock Exchange (TWSE) statistical/quote endpoint via the
svc/twse package and print the raw JSON envelope to stdout.)

範例 (Examples):
  yfin twse --endpoint MI_INDEX --date 20221230
  yfin twse --endpoint STOCK_DAY --date 20221230 --stock 2330
  yfin twse --endpoint FMSRFK --stock 2330 --date 2022
  yfin twse --endpoint MI_WEEK --date 20221230 --pretty`,
	RunE: runTwseEndpoint,
}

// twseFetcher is the uniform function signature used by nameToFetcher.
type twseFetcher func(ctx context.Context, client *svcTWSE.Client, date string, opts url.Values) (any, error)

// twseNameToFetcher maps an endpoint name to its fetcher.
var twseNameToFetcher = map[string]twseFetcher{
	"MI_INDEX":      svcTWSE.FetchMI_INDEX,
	"STOCK_DAY":     svcTWSE.FetchSTOCK_DAY,
	"BWIBBU_d":      svcTWSE.FetchBWIBBU_d,
	"MI_INDEX_PLUS": svcTWSE.FetchMI_INDEX_PLUS,
	"MI_INDEX_ODD":  svcTWSE.FetchMI_INDEX_ODD,
	"MI_5MINS":      svcTWSE.FetchMI_5MINS,
	"TWTB4U":        svcTWSE.FetchTWTB4U,
	"MI_MARGN":      svcTWSE.FetchMI_MARGN,
	"T86":           svcTWSE.FetchT86,
	"MI_QFIIS":      svcTWSE.FetchMI_QFIIS,
	"BFI82U":        svcTWSE.FetchBFI82U,
	"TWT38U":        svcTWSE.FetchTWT38U,
	"TWT43U":        svcTWSE.FetchTWT43U,
	"TWT44U":        svcTWSE.FetchTWT44U,
	"BFIAUU":        svcTWSE.FetchBlockBFIAUU, //nolint:misspell // upstream naming; 4-col & 10-col were consolidated to 10-col
	"BFIAUU_STOCK":  svcTWSE.FetchBFIAUUSTOCK,
	"BFIMUU":        svcTWSE.FetchBFIMUU,
	"BFIAUU_YEAR":   svcTWSE.FetchBFIAUUYEAR,
	"FMTQIK":        svcTWSE.FetchFMTQIK,
	"STOCK_DAY_AVG": svcTWSE.FetchStockDayAvg,
	"FMSRFK": func(ctx context.Context, client *svcTWSE.Client, date string, opts url.Values) (any, error) {
		stockNo := opts.Get("stockNo")
		if stockNo == "" {
			return nil, fmt.Errorf("FMSRFK: --stock is required")
		}
		return svcTWSE.FetchFMSRFK(ctx, client, stockNo, date, opts)
	},
	"BFIAMU":  svcTWSE.FetchBFIAMU,
	"MI_WEEK": svcTWSE.FetchMI_WEEK,
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

// Register attaches the `twse` subcommand onto rootCmd.
func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(twseCmd)
}

// runTwseEndpoint is the RunE for `yfin twse`.
func runTwseEndpoint(cmd *cobra.Command, args []string) error {
	ep, ok := svcTWSE.Registry[twseCfg.endpoint]
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: unknown endpoint %q (use --endpoint MI_INDEX, STOCK_DAY, ...)\n", twseCfg.endpoint)
		return fmt.Errorf("unknown endpoint")
	}
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
		fmt.Fprintf(os.Stderr, "ERROR: endpoint %q has no fetcher wired in cmd/twse/twse.go\n", twseCfg.endpoint)
		return fmt.Errorf("no fetcher")
	}

	opts := url.Values{}
	if twseCfg.stockNo != "" {
		opts.Set("stockNo", twseCfg.stockNo)
	}
	if twseCfg.month != "" {
		opts.Set("month", twseCfg.month)
	}

	ctx, cancel := context.WithTimeout(context.Background(), twseCfg.timeout+5*time.Second)
	defer cancel()

	client := twseClientProvider()
	raw, err := fetcher(ctx, client, twseCfg.date, opts)
	if err != nil {
		if errors.Is(err, svcTWSE.ErrNoData) || strings.Contains(err.Error(), "no data") {
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

// twseClientProvider returns the `*svcTWSE.Client` used by every twse
// command invocation. Production wiring points at `buildTWSEClient`; tests
// reassign this variable via `setTwseClientForTest` in cmd/twse/twse_test.go.
var twseClientProvider = func() *svcTWSE.Client {
	return buildTWSEClient()
}
