// bfiauu_year.go 對應 `/block/BFIAUU_YEAR` 端點。
// 用途:鉅額交易年成交資訊(年度、成交筆數、股數、金額)。
// 對應 README.tsme.md「鉅額交易」章節。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/block/BFIAUU_YEAR?date=20220101&response=json"

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
func FetchBFIAUUYEAR(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
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
	return FetchJSON[BFIAUUYEARResponse](ctx, c, "/block/BFIAUU_YEAR", q)
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
