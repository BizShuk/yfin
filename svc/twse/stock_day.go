// stock_day.go 對應 `/afterTrading/STOCK_DAY` 端點。
// 用途:個股日成交資訊(開高低收、成交量、成交金額、成交筆數)。
// 對應 README.tsme.md「盤後交易資訊」第 2 個端點;需指定 stockNo。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/afterTrading/STOCK_DAY?date=20221230&stockNo=2330&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// STOCK_DAYResponse embeds the common Response envelope and adds the
// `date` and `stockNo` fields that TWSE returns on this endpoint.
type STOCK_DAYResponse struct {
	Response
	Date    string `json:"date"`
	StockNo string `json:"stockNo"`
}

// GetStat returns the embedded stat field.
func (r *STOCK_DAYResponse) GetStat() string { return r.Response.Stat }

// StockDayRow is a typed representation of one STOCK_DAY data row.
// Fields: 日期, 成交股數, 成交金額, 開盤, 最高, 最低, 收盤, 漲跌價差, 成交筆數.
type StockDayRow struct {
	Date         string  // 日期
	Volume       int64   // 成交股數
	Amount       int64   // 成交金額
	Open         float64 // 開盤
	High         float64 // 最高
	Low          float64 // 最低
	Close        float64 // 收盤
	Change       float64 // 漲跌價差
	Transactions int64   // 成交筆數
}

// FetchSTOCK_DAY retrieves per-stock daily trade info for `date` and
// `stockNo` (must be supplied via opts).
func FetchSTOCK_DAY(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/STOCK_DAY: date is required")
	}
	stockNo := opts.Get("stockNo")
	if stockNo == "" {
		return nil, fmt.Errorf("twse/STOCK_DAY: stockNo is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("stockNo", stockNo)
	for k, vs := range opts {
		if k == "stockNo" {
			continue
		}
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[STOCK_DAYResponse](ctx, c, "/afterTrading/STOCK_DAY", q)
}

// ParseStockDayRow converts one raw `data` row into a typed StockDayRow.
func ParseStockDayRow(row []string) (StockDayRow, error) {
	if len(row) < 9 {
		return StockDayRow{}, fmt.Errorf("STOCK_DAY: row too short: %d cols", len(row))
	}
	return StockDayRow{
		Date:         strings.TrimSpace(row[0]),
		Volume:       ParseInt(row[1]),
		Amount:       ParseInt(row[2]),
		Open:         ParseFloat(row[3]),
		High:         ParseFloat(row[4]),
		Low:          ParseFloat(row[5]),
		Close:        ParseFloat(row[6]),
		Change:       ParseFloat(row[7]),
		Transactions: ParseInt(row[8]),
	}, nil
}
