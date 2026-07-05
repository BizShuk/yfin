// twt43u.go — `TWT43U` (/fund/TWT43U) investment-trust (投信) daily aggregated buy/sell/net volume. Capacity: 1 summary row per response.
package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// TWT43UResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type TWT43UResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *TWT43UResponse) GetStat() string { return r.Response.Stat }

// TWT43URow is a typed representation of one TWT43U data row.
// Fields: 單位名稱, 買進股數, 賣出股數, 買賣差額股數.
type TWT43URow struct {
	UnitName string // 單位名稱
	Buy      int64  // 買進股數
	Sell     int64  // 賣出股數
	Net      int64  // 買賣差額股數
}

// FetchTWT43U retrieves the daily aggregated buy/sell volume of
// investment trust companies (投信) for `date`.
func FetchTWT43U(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/TWT43U: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[TWT43UResponse](ctx, client, "/fund/TWT43U", q)
}

// ParseTWT43URow converts one raw `data` row into a typed TWT43URow.
func ParseTWT43URow(row []string) (TWT43URow, error) {
	if len(row) < 4 {
		return TWT43URow{}, fmt.Errorf("TWT43U: row too short: %d cols", len(row))
	}
	return TWT43URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseInt(row[1]),
		Sell:     ParseInt(row[2]),
		Net:      ParseInt(row[3]),
	}, nil
}
