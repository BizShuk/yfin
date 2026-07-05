// mi_index_odd.go — `MI_INDEX_ODD` (/afterTrading/MI_INDEX_ODD) odd-lot (零股) trading snapshot per stock (volume/amount/OHLC). Capacity: ~1900 stocks per response.
package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// MI_INDEX_ODDResponse embeds the common Response envelope and adds
// the `date` field that TWSE returns on this endpoint.
type MI_INDEX_ODDResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_INDEX_ODDResponse) GetStat() string { return r.Response.Stat }

// MIIndexOddRow is a typed representation of one MI_INDEX_ODD data row.
// Fields: 證券代號, 證券名稱, 成交股數, 成交金額, 開盤, 最高, 最低, 收盤.
type MIIndexOddRow struct {
	Code   string  // 證券代號
	Name   string  // 證券名稱
	Volume int64   // 成交股數
	Amount int64   // 成交金額
	Open   float64 // 開盤
	High   float64 // 最高
	Low    float64 // 最低
	Close  float64 // 收盤
}

// FetchMI_INDEX_ODD retrieves the odd-lot (零股) trading snapshot for
// `date`.
func FetchMI_INDEX_ODD(ctx context.Context, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_INDEX_ODD: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_INDEX_ODDResponse](ctx, "/afterTrading/MI_INDEX_ODD", q)
}

// ParseMIIndexOddRow converts one raw `data` row into a typed
// MIIndexOddRow.
func ParseMIIndexOddRow(row []string) (MIIndexOddRow, error) {
	if len(row) < 8 {
		return MIIndexOddRow{}, fmt.Errorf("MI_INDEX_ODD: row too short: %d cols", len(row))
	}
	return MIIndexOddRow{
		Code:   strings.TrimSpace(row[0]),
		Name:   strings.TrimSpace(row[1]),
		Volume: ParseInt(row[2]),
		Amount: ParseInt(row[3]),
		Open:   ParseFloat(row[4]),
		High:   ParseFloat(row[5]),
		Low:    ParseFloat(row[6]),
		Close:  ParseFloat(row[7]),
	}, nil
}
