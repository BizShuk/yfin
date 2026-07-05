// mi_week.go — `MI_WEEK` (/statistics/MI_WEEK) weekly stock market-cap report (stock code/name + shares issued + market cap). Capacity: ~1900 stocks per response.
package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// MI_WEEKResponse embeds the common Response envelope and adds the `date`
// field that TWSE returns for /statistics/MI_WEEK.
type MI_WEEKResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_WEEKResponse) GetStat() string { return r.Response.Stat }

// MIWeekRow is a typed representation of one MI_WEEK data row.
// Columns: 股票代號, 股票名稱, 發行股數, 市值.
type MIWeekRow struct {
	StockCode    string // 股票代號
	StockName    string // 股票名稱
	SharesIssued int64  // 發行股數
	MarketCap    int64  // 市值
}

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
