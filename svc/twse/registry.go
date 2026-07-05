// registry.go — endpoint descriptor (`Endpoint` + `Fetcher` signature) and global `Registry` map (code → descriptor) for CLI/help/dispatch. Capacity: 23 entries across 5 boards (afterTrading/marginTrading/fund/block/statistics) + 1 `Dispatcher`.
package twse

import (
	"context"
	"fmt"
	"net/url"
)

// Fetcher is the per-endpoint contract: build a query, call FetchJSON,
// return the typed DTO. Each endpoint file provides its own concrete
// Fetcher. The HTTP transport is injected via the `*Client` parameter;
// the Fetcher never reaches for a package-level singleton.
type Fetcher func(ctx context.Context, client *Client, baseDate string, opts url.Values) (any, error)

// Endpoint describes a TWSE endpoint for CLI help/dispatch.
type Endpoint struct {
	Name        string
	Board       string // "afterTrading" | "fund" | "block" | "statistics" | "marginTrading"
	Path        string
	Description string
	NeedsStock  bool
	NeedsMonth  bool
	Fetch       Fetcher
}

// Registry maps endpoint code (e.g. "MI_INDEX") to its descriptor.
//
// Fetch is nil for every entry at this stage; per-endpoint Fetcher
// functions are wired in by the endpoint files (one per file under
// svc/twse/) and bound to the CLI dispatch map in cmd/twse.go.
var Registry = map[string]Endpoint{
	"MI_INDEX":      {Name: "MI_INDEX", Board: "afterTrading", Path: "/afterTrading/MI_INDEX", Description: "每日收盤行情"},
	"STOCK_DAY":     {Name: "STOCK_DAY", Board: "afterTrading", Path: "/afterTrading/STOCK_DAY", Description: "個股日成交資訊", NeedsStock: true},
	"BWIBBU_d":      {Name: "BWIBBU_d", Board: "afterTrading", Path: "/afterTrading/BWIBBU_d", Description: "個股日本益比、殖利率及股價淨值比"},
	"MI_INDEX_PLUS": {Name: "MI_INDEX_PLUS", Board: "afterTrading", Path: "/afterTrading/MI_INDEX_PLUS", Description: "盤後定價交易"},
	"MI_INDEX_ODD":  {Name: "MI_INDEX_ODD", Board: "afterTrading", Path: "/afterTrading/MI_INDEX_ODD", Description: "零股交易行情單"},
	"MI_5MINS":      {Name: "MI_5MINS", Board: "afterTrading", Path: "/afterTrading/MI_5MINS", Description: "每5秒委託成交統計"},
	"TWTB4U":        {Name: "TWTB4U", Board: "afterTrading", Path: "/afterTrading/TWTB4U", Description: "當日沖銷交易標的及統計"},
	"MI_MARGN":      {Name: "MI_MARGN", Board: "marginTrading", Path: "/marginTrading/MI_MARGN", Description: "融資融券餘額"},
	"T86":           {Name: "T86", Board: "fund", Path: "/fund/T86", Description: "三大法人買賣超日報"},
	"MI_QFIIS":      {Name: "MI_QFIIS", Board: "fund", Path: "/fund/MI_QFIIS", Description: "外資及陸資投資持股統計"},
	"BFI82U":        {Name: "BFI82U", Board: "fund", Path: "/fund/BFI82U", Description: "三大法人買賣金額統計表"},
	"TWT38U":        {Name: "TWT38U", Board: "fund", Path: "/fund/TWT38U", Description: "外資及陸資買賣超彙總表"},
	"TWT43U":        {Name: "TWT43U", Board: "fund", Path: "/fund/TWT43U", Description: "投信買賣超彙總表"},
	"TWT44U":        {Name: "TWT44U", Board: "fund", Path: "/fund/TWT44U", Description: "自營商買賣超彙總表"},
	"BFIAUU":        {Name: "BFIAUU", Board: "block", Path: "/block/BFIAUU", Description: "鉅額交易日成交資訊"},
	"BFIAUU_STOCK":  {Name: "BFIAUU_STOCK", Board: "block", Path: "/block/BFIAUU", Description: "單一證券日成交資訊", NeedsStock: true},
	"BFIMUU":        {Name: "BFIMUU", Board: "block", Path: "/block/BFIMUU", Description: "鉅額交易月成交資訊", NeedsMonth: true},
	"BFIAUU_YEAR":   {Name: "BFIAUU_YEAR", Board: "block", Path: "/block/BFIAUU_YEAR", Description: "鉅額交易年成交資訊"},
	"FMTQIK":        {Name: "FMTQIK", Board: "statistics", Path: "/exchangeReport/FMTQIK", Description: "臺股指數及交易量表", NeedsMonth: true},
	"STOCK_DAY_AVG": {Name: "STOCK_DAY_AVG", Board: "statistics", Path: "/exchangeReport/STOCK_DAY_AVG", Description: "個股月均價", NeedsStock: true, NeedsMonth: true},
	"FMSRFK":        {Name: "FMSRFK", Board: "statistics", Path: "/exchangeReport/FMSRFK", Description: "個股月成交資訊", NeedsStock: true},
	"BFIAMU":        {Name: "BFIAMU", Board: "statistics", Path: "/afterTrading/BFIAMU", Description: "每日各類指數成交量值"},
	"MI_WEEK":       {Name: "MI_WEEK", Board: "statistics", Path: "/statistics/MI_WEEK", Description: "股票市值週報"},
}

// Dispatcher resolves an endpoint code to its Fetcher and invokes it
// against a shared `*Client`. Use NewDispatcher at composition-root time
// (one dispatcher per process, holding the shared client) and call
// `Call(ctx, code, date, opts)` to dispatch.
type Dispatcher struct {
	client *Client
}

// NewDispatcher binds a `*Client` to a dispatcher. The dispatcher is
// safe to share across goroutines — fetchers themselves are pure
// functions.
func NewDispatcher(client *Client) *Dispatcher {
	return &Dispatcher{client: client}
}

// Client returns the `*Client` the dispatcher was bound to.
func (d *Dispatcher) Client() *Client { return d.client }

// Call looks up `code` in `Registry`, invokes its Fetcher with
// `d.client`, and returns whatever the Fetcher returns. An unknown
// code yields `fmt.Errorf("unknown endpoint: %s", code)`.
func (d *Dispatcher) Call(ctx context.Context, code, baseDate string, opts url.Values) (any, error) {
	ep, ok := Registry[code]
	if !ok {
		return nil, fmt.Errorf("unknown endpoint: %s", code)
	}
	return ep.Fetch(ctx, d.client, baseDate, opts)
}

// init wires the per-endpoint Fetcher functions into the Registry map
// at package-init time. The functions are declared in their own files
// under svc/twse/ (one per endpoint); we can't initialize them inline
// in the var literal because Go forbids referring to a function by
// name before it has been declared.
func init() {
	setFetcher := func(code string, f Fetcher) {
		ep := Registry[code]
		ep.Fetch = f
		Registry[code] = ep
	}
	setFetcher("MI_INDEX", FetchMI_INDEX)
	setFetcher("STOCK_DAY", FetchSTOCK_DAY)
	setFetcher("BWIBBU_d", FetchBWIBBU_d)
	setFetcher("MI_INDEX_PLUS", FetchMI_INDEX_PLUS)
	setFetcher("MI_INDEX_ODD", FetchMI_INDEX_ODD)
	setFetcher("MI_5MINS", FetchMI_5MINS)
	setFetcher("TWTB4U", FetchTWTB4U)
	setFetcher("MI_MARGN", FetchMI_MARGN)
	setFetcher("T86", FetchT86)
	setFetcher("MI_QFIIS", FetchMI_QFIIS)
	setFetcher("BFI82U", FetchBFI82U)
	setFetcher("TWT38U", FetchTWT38U)
	setFetcher("TWT43U", FetchTWT43U)
	setFetcher("TWT44U", FetchTWT44U)
	setFetcher("BFIAUU", FetchBlockBFIAUU)
	setFetcher("BFIAUU_STOCK", FetchBFIAUUSTOCK)
	setFetcher("BFIMUU", FetchBFIMUU)
	setFetcher("BFIAUU_YEAR", FetchBFIAUUYEAR)
	setFetcher("FMTQIK", FetchFMTQIK)
	setFetcher("STOCK_DAY_AVG", FetchStockDayAvg)
	// FMSRFK has a different signature — wrap it.
	setFetcher("FMSRFK", func(ctx context.Context, client *Client, baseDate string, opts url.Values) (any, error) {
		stockNo := opts.Get("stockNo")
		if stockNo == "" {
			return nil, fmt.Errorf("twse/FMSRFK: stockNo is required")
		}
		return FetchFMSRFK(ctx, client, stockNo, baseDate, opts)
	})
	setFetcher("BFIAMU", FetchBFIAMU)
	setFetcher("MI_WEEK", FetchMI_WEEK)
}
