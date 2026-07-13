package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)

// Type aliases — structs now live in model/twse.go.
type (
	MI_WEEKResponse = model.MI_WEEKResponse
	MIWeekRow = model.MIWeekRow
)

// MI_WEEKResponse embeds the common Response envelope and adds the `date`
// field that TWSE returns for /statistics/MI_WEEK.

// MIWeekRow is a typed representation of one MI_WEEK data row.
// Columns: 股票代號, 股票名稱, 發行股數, 市值.

// FetchMI_WEEK retrieves the weekly stock market-cap report for `date`.
// `date` is required (YYYYMMDD).
func FetchMI_WEEK(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_WEEK: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_WEEKResponse](ctx, client, "/statistics/MI_WEEK", q)
}

// ParseMIWeekRow converts one raw `data` row into a typed MIWeekRow.
func ParseMIWeekRow(row []string) (MIWeekRow, error) {
	if len(row) < 4 {
		return MIWeekRow{}, fmt.Errorf("MI_WEEK: row too short: %d cols", len(row))
	}
	return MIWeekRow{
		StockCode:    strings.TrimSpace(row[0]),
		StockName:    strings.TrimSpace(row[1]),
		SharesIssued: ParseInt(row[2]),
		MarketCap:    ParseInt(row[3]),
	}, nil
}
