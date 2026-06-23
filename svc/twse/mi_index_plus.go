// mi_index_plus.go 對應 `/afterTrading/MI_INDEX_PLUS` 端點。
// 用途:盤後定價交易(收盤指數、漲跌)。
// 對應 README.tsme.md「盤後交易資訊」第 4 個端點。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/afterTrading/MI_INDEX_PLUS?date=20221230&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// MI_INDEX_PLUSResponse embeds the common Response envelope and adds
// the `date` field that TWSE returns on this endpoint.
type MI_INDEX_PLUSResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_INDEX_PLUSResponse) GetStat() string { return r.Response.Stat }

// MIIndexPlusRow is a typed representation of one MI_INDEX_PLUS data row.
// Fields: 指數, 收盤指數, 漲跌點數, 漲跌百分比.
type MIIndexPlusRow struct {
	IndexName string  // 指數
	Close     float64 // 收盤指數
	Change    float64 // 漲跌點數
	ChangePct float64 // 漲跌百分比
}

// FetchMI_INDEX_PLUS retrieves the after-hours (盤後定價) index data
// for `date`.
func FetchMI_INDEX_PLUS(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_INDEX_PLUS: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_INDEX_PLUSResponse](ctx, c, "/afterTrading/MI_INDEX_PLUS", q)
}

// ParseMIIndexPlusRow converts one raw `data` row into a typed
// MIIndexPlusRow.
func ParseMIIndexPlusRow(row []string) (MIIndexPlusRow, error) {
	if len(row) < 4 {
		return MIIndexPlusRow{}, fmt.Errorf("MI_INDEX_PLUS: row too short: %d cols", len(row))
	}
	return MIIndexPlusRow{
		IndexName: strings.TrimSpace(row[0]),
		Close:     ParseFloat(row[1]),
		Change:    ParseFloat(row[2]),
		ChangePct: ParsePercent(row[3]),
	}, nil
}
