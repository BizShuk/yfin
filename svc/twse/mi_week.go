package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)


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
	return FetchJSON[model.MI_WEEKResponse](ctx, client, "/statistics/MI_WEEK", q)
}

// ParseMIWeekRow converts one raw `data` row into a typed MIWeekRow.
func ParseMIWeekRow(row []string) (model.MIWeekRow, error) {
	if len(row) < 4 {
		return model.MIWeekRow{}, fmt.Errorf("MI_WEEK: row too short: %d cols", len(row))
	}
	return model.MIWeekRow{
		StockCode:    strings.TrimSpace(row[0]),
		StockName:    strings.TrimSpace(row[1]),
		SharesIssued: ParseInt(row[2]),
		MarketCap:    ParseInt(row[3]),
	}, nil
}
