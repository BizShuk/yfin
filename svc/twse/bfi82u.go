// bfi82u.go 對應 `/fund/BFI82U` 端點。
// 用途:三大法人買賣金額統計表(投信、自營商、外資及陸資)。
// 對應 README.tsme.md「三大法人」章節。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/fund/BFI82U?date=20221230&type=day&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// BFI82UResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type BFI82UResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *BFI82UResponse) GetStat() string { return r.Response.Stat }

// BFI82URow is a typed representation of one BFI82U data row.
// Fields: 單位名稱, 買進金額, 賣出金額, 買賣差額.
type BFI82URow struct {
	UnitName string  // 單位名稱
	Buy      float64 // 買進金額
	Sell     float64 // 賣出金額
	Net      float64 // 買賣差額
}

// FetchBFI82U retrieves the daily aggregated buy/sell amounts of the
// three main institutional investors (自營商, 投信, 外資及陸資) for `date`.
// TWSE requires `type=day` for this endpoint.
func FetchBFI82U(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFI82U: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("type", "day")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[BFI82UResponse](ctx, c, "/fund/BFI82U", q)
}

// ParseBFI82URow converts one raw `data` row into a typed BFI82URow.
func ParseBFI82URow(row []string) (BFI82URow, error) {
	if len(row) < 4 {
		return BFI82URow{}, fmt.Errorf("BFI82U: row too short: %d cols", len(row))
	}
	return BFI82URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseFloat(row[1]),
		Sell:     ParseFloat(row[2]),
		Net:      ParseFloat(row[3]),
	}, nil
}