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

	"github.com/bizshuk/yfin/svc/yahoo"
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
	case "major-holders", "institutional-holders", "mutualfund-holders":
		dto, err := c.yahooClient.FetchHolders(ctx, symbol)
		if err != nil {
			return nil, err
		}
		return projectHolders(command, dto)
	case "insider-transactions", "insider-purchases", "insider-roster":
		dto, err := c.yahooClient.FetchInsider(ctx, symbol)
		if err != nil {
			return nil, err
		}
		return projectInsider(command, dto)
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

func projectHolders(command string, dto *yahoo.HoldersDTO) (any, error) {
	switch command {
	case "major-holders":
		return dto.MajorDirectHolders, nil
	case "institutional-holders":
		return dto.InstitutionOwnership, nil
	case "mutualfund-holders":
		return dto.FundOwnership, nil
	default:
		return nil, fmt.Errorf("unsupported holders command %q", command)
	}
}

func projectInsider(command string, dto *yahoo.InsiderDTO) (any, error) {
	switch command {
	case "insider-transactions":
		return dto.Transactions, nil
	case "insider-purchases":
		return yahoo.InsiderPurchaseSummaryTable(&dto.PurchaseActivity), nil
	case "insider-roster":
		return dto.Roster, nil
	default:
		return nil, fmt.Errorf("unsupported insider command %q", command)
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
