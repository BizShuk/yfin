// fmsrfk.go — `FMSRFK` (/exchangeReport/FMSRFK) per-stock monthly aggregate (year-month/high/low/weighted avg + volume + turnover %). Capacity: 1 stock × 12 monthly rows per year-long request.
package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// FMSRFKResponse embeds the common Response envelope and adds stockNo/date
// (year) fields that TWSE returns for /exchangeReport/FMSRFK.
type FMSRFKResponse struct {
	Response
	StockNo string `json:"stockNo"`
	Date    string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *FMSRFKResponse) GetStat() string { return r.Response.Stat }

// FMSRFKRow is a typed representation of one FMSRFK data row.
// Columns: 年度, 月份, 最高, 最低, 加權平均價, 成交股數, 成交金額, 週轉率%.
type FMSRFKRow struct {
	Year        string  // 年度
	Month       string  // 月份
	High        float64 // 最高
	Low         float64 // 最低
	WAvgPrice   float64 // 加權平均價
	TradeVolume int64   // 成交股數
	TradeValue  int64   // 成交金額
	TurnoverPct float64 // 週轉率%
}

// FetchFMSRFK retrieves per-stock monthly trading info for the year `date`.
// `stockNo` is required (e.g. "2330"); `date` is the year (e.g. "2022").
func FetchFMSRFK(ctx context.Context, client *Client, stockNo, date string, opts url.Values) (any, error) {
	if stockNo == "" {
		return nil, fmt.Errorf("twse/FMSRFK: stockNo is required")
	}
	if date == "" {
		return nil, fmt.Errorf("twse/FMSRFK: date is required")
	}
	q := url.Values{}
	q.Set("stockNo", stockNo)
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[FMSRFKResponse](ctx, client, "/exchangeReport/FMSRFK", q)
}

// ParseFMSRFKRow converts one raw `data` row into a typed FMSRFKRow.
func ParseFMSRFKRow(row []string) (FMSRFKRow, error) {
	if len(row) < 8 {
		return FMSRFKRow{}, fmt.Errorf("FMSRFK: row too short: %d cols", len(row))
	}
	return FMSRFKRow{
		Year:        strings.TrimSpace(row[0]),
		Month:       strings.TrimSpace(row[1]),
		High:        ParseFloat(row[2]),
		Low:         ParseFloat(row[3]),
		WAvgPrice:   ParseFloat(row[4]),
		TradeVolume: ParseInt(row[5]),
		TradeValue:  ParseInt(row[6]),
		TurnoverPct: ParseFloat(row[7]),
	}, nil
}
