// mi_index.go — `MI_INDEX` (/afterTrading/MI_INDEX) daily market index close + change + change-pct (issuance-weight, OTC, sector breakdowns). Capacity: ~50 indices per response.
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
func FetchMI_INDEX(ctx context.Context, date string, opts url.Values) (any, error) {
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
	return FetchJSON[MI_INDEXResponse](ctx, "/afterTrading/MI_INDEX", q)
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
