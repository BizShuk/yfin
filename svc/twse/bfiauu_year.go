// bfiauu_year.go — `BFIAUU_YEAR` (/block/BFIAUU_YEAR) annual block-trade aggregate (year/trades/volume/amount). Capacity: 10+ years per response (date = YYYY0101).
package twse

import (
	"context"
	"fmt"
	"net/url"
)

// BFIAUUYEARResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type BFIAUUYEARResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *BFIAUUYEARResponse) GetStat() string { return r.Response.Stat }

// BFIAUUYEARRow is a typed representation of one BFIAUU_YEAR data row.
// Fields: 年度, 成交筆數, 成交股數, 成交金額.
type BFIAUUYEARRow struct {
	Year         string // 年度
	Transactions int64  // 成交筆數
	Volume       int64  // 成交股數
	Amount       int64  // 成交金額
}

// FetchBFIAUUYEAR retrieves the annual block-trade report for `date` (YYYY0101).
func FetchBFIAUUYEAR(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFIAUU_YEAR: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[BFIAUUYEARResponse](ctx, client, "/block/BFIAUU_YEAR", q)
}

// ParseBFIAUUYEARRow converts one raw `data` row into a typed BFIAUUYEARRow.
func ParseBFIAUUYEARRow(row []string) (BFIAUUYEARRow, error) {
	if len(row) < 4 {
		return BFIAUUYEARRow{}, fmt.Errorf("BFIAUU_YEAR: row too short: %d cols", len(row))
	}
	return BFIAUUYEARRow{
		Year:         row[0],
		Transactions: ParseInt(row[1]),
		Volume:       ParseInt(row[2]),
		Amount:       ParseInt(row[3]),
	}, nil
}
