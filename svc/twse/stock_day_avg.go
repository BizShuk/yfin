// stock_day_avg.go 對應 `/exchangeReport/STOCK_DAY_AVG` 端點。
// 用途:個股月均價(最高、最低、加權平均、成交筆數/股數/金額)。
// 對應 README.tsme.md「大盤統計」章節;月份格式 YYYYMM01。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/exchangeReport/STOCK_DAY_AVG?date=20221201&stockNo=2330&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
)

// StockDayAvgResponse embeds the common Response envelope and adds the
// `date` and `stockNo` fields that TWSE returns on this endpoint.
type StockDayAvgResponse struct {
	Response
	Date    string `json:"date"`
	StockNo string `json:"stockNo"`
}

// GetStat returns the embedded stat field.
func (r *StockDayAvgResponse) GetStat() string { return r.Response.Stat }

// StockDayAvgRow is a typed representation of one STOCK_DAY_AVG data row.
// Fields: 年度, 月份, 最高, 最低, 加權平均價, 成交筆數, 成交股數, 成交金額.
type StockDayAvgRow struct {
	Year         string  // 年度
	Month        string  // 月份
	High         float64 // 最高
	Low          float64 // 最低
	WeightedAvg  float64 // 加權平均價
	Transactions int64   // 成交筆數
	Volume       int64   // 成交股數
	Amount       int64   // 成交金額
}

// FetchStockDayAvg retrieves the per-stock monthly average price for `date`
// (YYYYMM01). `stockNo` must be supplied via opts.
func FetchStockDayAvg(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/STOCK_DAY_AVG: date is required")
	}
	stockNo := opts.Get("stockNo")
	if stockNo == "" {
		return nil, fmt.Errorf("twse/STOCK_DAY_AVG: stockNo is required")
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
	return FetchJSON[StockDayAvgResponse](ctx, c, "/exchangeReport/STOCK_DAY_AVG", q)
}

// ParseStockDayAvgRow converts one raw `data` row into a typed StockDayAvgRow.
func ParseStockDayAvgRow(row []string) (StockDayAvgRow, error) {
	if len(row) < 8 {
		return StockDayAvgRow{}, fmt.Errorf("STOCK_DAY_AVG: row too short: %d cols", len(row))
	}
	return StockDayAvgRow{
		Year:         row[0],
		Month:        row[1],
		High:         ParseFloat(row[2]),
		Low:          ParseFloat(row[3]),
		WeightedAvg:  ParseFloat(row[4]),
		Transactions: ParseInt(row[5]),
		Volume:       ParseInt(row[6]),
		Amount:       ParseInt(row[7]),
	}, nil
}
