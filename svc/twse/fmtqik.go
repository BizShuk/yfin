// fmtqik.go — `FMTQIK` (/exchangeReport/FMTQIK) TAIEX index and daily volume/amount/trade-count snapshot. Capacity: 1 day-row per response.
package twse

import (
	"context"
	"fmt"
	"net/url"
)

// FMTQIKResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type FMTQIKResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *FMTQIKResponse) GetStat() string { return r.Response.Stat }

// FMTQIKRow is a typed representation of one FMTQIK data row.
// Fields: 日期, 成交股數, 成交金額, 成交筆數, 發行量加權股價指數.
type FMTQIKRow struct {
	Date         string  // 日期
	Volume       int64   // 成交股數
	Amount       int64   // 成交金額
	Transactions int64   // 成交筆數
	Index        float64 // 發行量加權股價指數
}

// FetchFMTQIK retrieves the TAIEX index and trading volume for `date`.
// `date` should be YYYYMMDD (month-start or month-end).
func FetchFMTQIK(ctx context.Context, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/FMTQIK: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[FMTQIKResponse](ctx, "/exchangeReport/FMTQIK", q)
}

// ParseFMTQIKRow converts one raw `data` row into a typed FMTQIKRow.
func ParseFMTQIKRow(row []string) (FMTQIKRow, error) {
	if len(row) < 5 {
		return FMTQIKRow{}, fmt.Errorf("FMTQIK: row too short: %d cols", len(row))
	}
	return FMTQIKRow{
		Date:         row[0],
		Volume:       ParseInt(row[1]),
		Amount:       ParseInt(row[2]),
		Transactions: ParseInt(row[3]),
		Index:        ParseFloat(row[4]),
	}, nil
}
