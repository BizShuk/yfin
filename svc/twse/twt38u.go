// twt38u.go 對應 `/fund/TWT38U` 端點。
// 用途:外資及陸資買賣超彙總表(買進、賣出、買賣差額股數)。
// 對應 README.tsme.md「外資及陸資」章節。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/fund/TWT38U?date=20221230&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// TWT38UResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type TWT38UResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *TWT38UResponse) GetStat() string { return r.Response.Stat }

// TWT38URow is a typed representation of one TWT38U data row.
// Fields: 單位名稱, 買進股數, 賣出股數, 買賣差額股數.
type TWT38URow struct {
	UnitName string  // 單位名稱
	Buy      int64   // 買進股數
	Sell     int64   // 賣出股數
	Net      int64   // 買賣差額股數
}

// FetchTWT38U retrieves the daily aggregated buy/sell volume of
// foreign investors (含陸資) for `date`.
func FetchTWT38U(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/TWT38U: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[TWT38UResponse](ctx, c, "/fund/TWT38U", q)
}

// ParseTWT38URow converts one raw `data` row into a typed TWT38URow.
func ParseTWT38URow(row []string) (TWT38URow, error) {
	if len(row) < 4 {
		return TWT38URow{}, fmt.Errorf("TWT38U: row too short: %d cols", len(row))
	}
	return TWT38URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseInt(row[1]),
		Sell:     ParseInt(row[2]),
		Net:      ParseInt(row[3]),
	}, nil
}