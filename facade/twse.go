// twse.go — facade-level wrapper over `svc/twse` so the yfin `twse`
// subcommand can dispatch to any of the 23 TWSE endpoints without
// importing svc/twse at runtime. svc/twse types may still appear in
// signatures (acceptable type import) — the rule is no direct
// svc/twse runtime calls from cmd/.
//
// The 23 `svc/twse.Fetch*` functions have uniform signatures
// `(ctx, *Client, date, url.Values) → (any, error)`; `twseFetchers`
// captures them in a name → fetcher map. The lone exception is FMSRFK,
// which takes `(ctx, *Client, stockNo, date, opts)` — wrapped inline so
// the public dispatch shape stays uniform.
package facade

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bizshuk/yfin/svc/twse"
	"github.com/bizshuk/yfin/utils/httpx"
)

// TwseFetcher is the uniform signature: `ctx, *twse.Client, date, opts →
// (any, error)`. All entries in twseFetchers satisfy this contract.
type TwseFetcher func(ctx context.Context, client *twse.Client, date string, opts url.Values) (any, error)

// twseFetchers maps endpoint name → fetcher.
var twseFetchers = map[string]TwseFetcher{
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
	"BFIAUU":        twse.FetchBlockBFIAUU, //nolint:misspell // upstream naming
	"BFIAUU_STOCK":  twse.FetchBFIAUUSTOCK,
	"BFIMUU":        twse.FetchBFIMUU,
	"BFIAUU_YEAR":   twse.FetchBFIAUUYEAR,
	"FMTQIK":        twse.FetchFMTQIK,
	"STOCK_DAY_AVG": twse.FetchStockDayAvg,
	// FMSRFK is the lone exception — it needs stockNo between client and date.
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

// TwseEndpoint is a re-export of svc/twse.Endpoint so cmd callers don't
// need to import svc/twse for the (NeedsStock, NeedsMonth, Path) fields
// they read in flag validation.
type TwseEndpoint = twse.Endpoint

// TwseRegistry is a re-export of svc/twse.Registry so cmd's flag
// validation can iterate it without importing svc/twse.
var TwseRegistry = twse.Registry

// TwseErrNoData is a re-export of svc/twse.ErrNoData so callers can
// detect empty-result responses with errors.Is().
var TwseErrNoData = twse.ErrNoData

// TwseDispatch is the single entry point for `yfin twse --endpoint X`.
// It validates the endpoint name, then dispatches to the matching fetcher.
// Returns the raw envelope (any) — JSON encoding is the caller's job.
func TwseDispatch(ctx context.Context, client *twse.Client, endpoint, date string, opts url.Values) (any, error) {
	ep, ok := TwseRegistry[endpoint]
	if !ok {
		return nil, fmt.Errorf("unknown endpoint %q", endpoint)
	}
	if ep.NeedsStock && opts.Get("stockNo") == "" {
		return nil, fmt.Errorf("endpoint %q requires --stock", endpoint)
	}
	if ep.NeedsMonth && opts.Get("month") == "" {
		return nil, fmt.Errorf("endpoint %q requires --month", endpoint)
	}

	fetcher, ok := twseFetchers[endpoint]
	if !ok {
		return nil, fmt.Errorf("endpoint %q has no fetcher wired in facade/twse.go", endpoint)
	}
	return fetcher(ctx, client, date, opts)
}

// TwseIsNoData reports whether err is a TWSE no-data error (sentinel
// ErrNoData or message contains "no data"). Replaces the inline check
// cmd/twse previously did.
func TwseIsNoData(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, TwseErrNoData) || strings.Contains(err.Error(), "no data")
}

// NewTwseClient builds the default process-wide `*twse.Client` tuned for
// TWSE's public REST API. The User-Agent is TWSE-specific (the default
// Go UA is rejected by TWSE). Tests can build their own via
// `svc/twse.NewClientWithURL` and inject through `SetTwseClientProvider`.
func NewTwseClient() *twse.Client {
	hc := httpx.NewClient(&httpx.Config{
		BaseURL:          "",
		Timeout:          30 * time.Second,
		IdleTimeout:      90 * time.Second,
		MaxConnsPerHost:  10,
		MaxAttempts:      3,
		BackoffBaseMs:    500,
		BackoffJitterMs:  250,
		MaxDelayMs:       8000,
		QPS:              2.0,
		Burst:            4,
		CircuitWindow:    60 * time.Second,
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		UserAgent:        twseUserAgent,
		MaxBodyBytes:     0,
	})
	hc.Use(httpx.TWSEMiddleware(twseUserAgent))
	return twse.NewClient(hc)
}

// twseUserAgent is the browser-like UA TWSE rejects the default Go
// `User-Agent` for.
const twseUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
