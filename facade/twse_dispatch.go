// twse_dispatch.go — endpoint registry re-exports + the name → fetcher
// map + the dispatch entry point behind `(*TwseClient).Dispatch`.
//
// The 23 `svc/twse.Fetch*` functions share the signature
// `(ctx, *Client, date, url.Values) → (any, error)`; `twseFetchers`
// captures them in a name → fetcher map. The lone exception is FMSRFK,
// which takes `(ctx, *Client, stockNo, date, opts)` — wrapped inline so
// the dispatch shape stays uniform.
//
// Capacity: `TwseEndpoint`/`TwseRegistry`/`TwseErrNoData` re-exports +
// `TwseFetcher` + `twseFetchers` (23 entries) + `twseDispatch` +
// `TwseHasFetcher`.
package facade

import (
	"context"
	"fmt"
	"net/url"

	"github.com/bizshuk/yfin/svc/twse"
)

// TwseEndpoint is a re-export of svc/twse.Endpoint so cmd callers can
// read the (NeedsStock, NeedsMonth, Path) fields during flag validation
// without importing svc/twse.
type TwseEndpoint = twse.Endpoint

// TwseRegistry is a re-export of svc/twse.Registry so cmd's flag
// validation can iterate it without importing svc/twse.
var TwseRegistry = twse.Registry

// TwseErrNoData is a re-export of svc/twse.ErrNoData so callers can
// detect empty-result responses with errors.Is().
var TwseErrNoData = twse.ErrNoData

// TwseFetcher is the uniform fetcher signature every entry in
// twseFetchers satisfies.
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

// TwseHasFetcher reports whether endpoint has a fetcher wired up. It lets
// tests assert registry ↔ fetcher-map coverage without dispatching.
func TwseHasFetcher(endpoint string) bool {
	_, ok := twseFetchers[endpoint]
	return ok
}

// twseDispatch validates the endpoint name and its required flags, then
// dispatches to the matching fetcher. Reached via (*TwseClient).Dispatch.
func twseDispatch(ctx context.Context, client *twse.Client, endpoint, date string, opts url.Values) (any, error) {
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
		return nil, fmt.Errorf("endpoint %q has no fetcher wired in facade/twse_dispatch.go", endpoint)
	}
	return fetcher(ctx, client, date, opts)
}
