// mi_index.go 對應 `/afterTrading/MI_INDEX` 端點。
// 用途:全市場每日收盤行情(價格指數、漲跌點數、漲跌百分比)。
// 對應 README.tsme.md「盤後交易資訊」第 1 個端點。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/afterTrading/MI_INDEX?date=20221230&type=ALL&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// MI_INDEXResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type MI_INDEXResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_INDEXResponse) GetStat() string { return r.Response.Stat }

// MIIndexRow is a typed representation of one MI_INDEX data row.
// Fields: 指數, 收盤指數, 漲跌點數, 漲跌百分比.
type MIIndexRow struct {
	IndexName string  // 指數
	Close     float64 // 收盤指數
	Change    float64 // 漲跌點數
	ChangePct float64 // 漲跌百分比
}

// FetchMI_INDEX retrieves the daily market index close for `date`.
// `opts` may include a `type=ALL` parameter (TWSE expects this).
func FetchMI_INDEX(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_INDEX: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("type", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_INDEXResponse](ctx, c, "/afterTrading/MI_INDEX", q)
}

// ParseMIIndexRow converts one raw `data` row into a typed MIIndexRow.
func ParseMIIndexRow(row []string) (MIIndexRow, error) {
	if len(row) < 4 {
		return MIIndexRow{}, fmt.Errorf("MI_INDEX: row too short: %d cols", len(row))
	}
	return MIIndexRow{
		IndexName: strings.TrimSpace(row[0]),
		Close:     ParseFloat(row[1]),
		Change:    ParseFloat(row[2]),
		ChangePct: ParsePercent(row[3]),
	}, nil
}
