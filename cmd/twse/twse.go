// twse.go — `twse` cobra subcommand. All TWSE dispatch logic
// (23 fetcher map / endpoint registry / URL builder / no-data
// detection / process-wide client construction) lives in
// `facade/twse.go` so this file is purely a thin cobra wrapper.
//
// Capacity: 1 `Register` + 1 `twseConfig` + 1 `twseCmd` var (kept for
// test introspection) + `runTwseEndpoint` RunE.
package twse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/bizshuk/yfin/facade"
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

// twseCmd is kept as a package-level var so tests can introspect and
// invoke `runTwseEndpoint(cmd, ...)` directly.
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

// twseClientProvider returns the *svcTWSE.Client used by every twse
// command invocation. Production wiring points at facade.NewTwseClient;
// tests reassign this variable via `setTwseClientForTest` in
// cmd/twse/twse_test.go.
var twseClientProvider = func() *svcTWSE.Client {
	return facade.NewTwseClient()
}

// runTwseEndpoint is the RunE for `yfin twse`.
func runTwseEndpoint(cmd *cobra.Command, args []string) error {
	if _, ok := facade.TwseRegistry[twseCfg.endpoint]; !ok {
		fmt.Fprintf(os.Stderr, "ERROR: unknown endpoint %q (use --endpoint MI_INDEX, STOCK_DAY, ...)\n", twseCfg.endpoint)
		return fmt.Errorf("unknown endpoint")
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
	raw, err := facade.TwseDispatch(ctx, client, twseCfg.endpoint, twseCfg.date, opts)
	if err != nil {
		if facade.TwseIsNoData(err) {
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
