// mi_qfiis.go 對應 `/fund/MI_QFIIS` 端點。
// 用途:外資及陸資投資持股統計(持有股數、佔發行股數%)。
// 對應 README.tsme.md「外資及陸資」章節。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/fund/MI_QFIIS?date=20221230&selectType=ALL&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
)

// MI_QFIISResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type MI_QFIISResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_QFIISResponse) GetStat() string { return r.Response.Stat }

// MI_QFIISRow is a typed representation of one MI_QFIIS data row.
type MI_QFIISRow struct {
	Code      string  // 證券代號
	Name      string  // 證券名稱
	SharesHeld int64  // 持有股數
	IssuePct   float64 // 佔發行股數%
}

// FetchMI_QFIIS retrieves the foreign+mainland investor holdings for `date`.
// selectType=ALL is always added by this fetcher.
func FetchMI_QFIIS(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_QFIIS: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_QFIISResponse](ctx, c, "/fund/MI_QFIIS", q)
}

// ParseMI_QFIISRow converts one raw `data` row into a typed MI_QFIISRow.
func ParseMI_QFIISRow(row []string) (MI_QFIISRow, error) {
	if len(row) < 4 {
		return MI_QFIISRow{}, fmt.Errorf("MI_QFIIS: row too short: %d cols", len(row))
	}
	return MI_QFIISRow{
		Code:       row[0],
		Name:       row[1],
		SharesHeld: ParseInt(row[2]),
		IssuePct:   ParsePercent(row[3]),
	}, nil
}