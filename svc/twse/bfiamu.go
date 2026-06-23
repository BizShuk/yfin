// bfiamu.go 對應 `/afterTrading/BFIAMU` 端點。
// 用途:每日各類指數成交量值(收盤指數、漲跌、百分比)。
// 對應 README.tsme.md「盤後交易資訊」附屬指數章節。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/afterTrading/BFIAMU?date=20221230&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// BFIAMUResponse embeds the common Response envelope and adds the `date`
// field that TWSE returns for /afterTrading/BFIAMU.
type BFIAMUResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *BFIAMUResponse) GetStat() string { return r.Response.Stat }

// BFIAMURow is a typed representation of one BFIAMU data row.
// Columns: 指數, 收盤指數, 漲跌, 百分比.
type BFIAMURow struct {
	IndexName string  // 指數
	Close     float64 // 收盤指數
	Change    float64 // 漲跌
	ChangePct float64 // 百分比
}

// FetchBFIAMU retrieves per-day index close & change values for `date`.
// `date` is required (YYYYMMDD).
func FetchBFIAMU(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFIAMU: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[BFIAMUResponse](ctx, c, "/afterTrading/BFIAMU", q)
}

// ParseBFIAMURow converts one raw `data` row into a typed BFIAMURow.
func ParseBFIAMURow(row []string) (BFIAMURow, error) {
	if len(row) < 4 {
		return BFIAMURow{}, fmt.Errorf("BFIAMU: row too short: %d cols", len(row))
	}
	return BFIAMURow{
		IndexName: strings.TrimSpace(row[0]),
		Close:     ParseFloat(row[1]),
		Change:    ParseFloat(row[2]),
		ChangePct: ParsePercent(row[3]),
	}, nil
}