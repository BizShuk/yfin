// yahoo.go — facade-level dispatchers for the svc/yahoo chart / quoteSummary
// endpoints that cmd/dispatch's skill-parity commandRegistry uses. Each
// `YahooDispatch` and `YahooFetchHistory` call wraps a single
// svc/yahoo.Client.Fetch* method so cmd/dispatch can drop its direct
// svc/yahoo import.
//
// All dispatchers return `(any, error)` — the dispatcher's job is to
// identify which svc method to invoke and hand back the raw envelope;
// JSON / struct decoding is the caller's responsibility.
package facade

import (
	"context"
	"fmt"
	"time"
)

// YahooDispatch maps a yfinance-style command name (info / actions /
// metadata / major-holders / insider-transactions / upgrades / calendar /
// earnings-dates / sec-filings / sustainability / recommendations /
// options / isin) to its corresponding svc/yahoo fetch. Returns the raw
// envelope (any) — callers must JSON-encode or assert the concrete type.
func (c *Client) YahooDispatch(ctx context.Context, command, symbol string) (any, error) {
	switch command {
	case "info":
		return c.yahooClient.FetchInfo(ctx, symbol)
	case "actions":
		return c.yahooClient.FetchActions(ctx, symbol)
	case "metadata":
		return c.yahooClient.FetchMetadata(ctx, symbol)
	// FetchHolders serves 3 Python yfinance aliases (major / institutional /
	// mutualfund) — they're the same Yahoo endpoint.
	case "major-holders", "institutional-holders", "mutualfund-holders":
		return c.yahooClient.FetchHolders(ctx, symbol)
	// FetchInsider serves 3 aliases (transactions / purchases / roster).
	case "insider-transactions", "insider-purchases", "insider-roster":
		return c.yahooClient.FetchInsider(ctx, symbol)
	case "upgrades":
		return c.yahooClient.FetchUpgrades(ctx, symbol)
	case "calendar":
		return c.yahooClient.FetchCalendar(ctx, symbol)
	case "earnings-dates":
		return c.yahooClient.FetchEarningsDates(ctx, symbol)
	case "sec-filings":
		return c.yahooClient.FetchSecFilings(ctx, symbol)
	case "sustainability":
		return c.yahooClient.FetchESG(ctx, symbol)
	// FetchRecommendationTrend serves both aliases.
	case "recommendations", "recommendations-summary":
		return c.yahooClient.FetchRecommendationTrend(ctx, symbol)
	case "options":
		return c.yahooClient.FetchOptions(ctx, symbol)
	case "isin":
		return c.yahooClient.FetchISIN(ctx, symbol)
	default:
		return nil, fmt.Errorf("unknown yahoo command %q", command)
	}
}

// YahooFetchHistory is the yfinance "history" command — fetches the last
// 30 days of daily bars and returns facade.BarBatch (plain SDK). Mirrors
// the previous cmd/dispatch/ behaviour (30-day lookback window).
func (c *Client) YahooFetchHistory(ctx context.Context, symbol string) (any, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	return c.FetchDailyBars(ctx, symbol, start, end, true, "yfin-batch")
}
