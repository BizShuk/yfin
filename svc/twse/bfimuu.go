// bfimuu.go — `BFIMUU` (/block/BFIMUU) monthly block-trade aggregate (period/trades/volume/amount, date = YYYYMM01). Capacity: multi-year monthly rows per response.
package twse

import (
	"context"
	"fmt"
	"net/url"
)

// BFIMUResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type BFIMUResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *BFIMUResponse) GetStat() string { return r.Response.Stat }

// BFIMUURow is a typed representation of one BFIMUU data row.
// Fields: 年月份, 成交筆數, 成交股數, 成交金額.
type BFIMUURow struct {
	Period       string // 年月份
	Transactions int64  // 成交筆數
	Volume       int64  // 成交股數
	Amount       int64  // 成交金額
}

// FetchBFIMUU retrieves the monthly block-trade report for `date` (YYYYMM01).
func FetchBFIMUU(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFIMUU: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[BFIMUResponse](ctx, client, "/block/BFIMUU", q)
}

// ParseBFIMUURow converts one raw `data` row into a typed BFIMUURow.
func ParseBFIMUURow(row []string) (BFIMUURow, error) {
	if len(row) < 4 {
		return BFIMUURow{}, fmt.Errorf("BFIMUU: row too short: %d cols", len(row))
	}
	return BFIMUURow{
		Period:       row[0],
		Transactions: ParseInt(row[1]),
		Volume:       ParseInt(row[2]),
		Amount:       ParseInt(row[3]),
	}, nil
}
