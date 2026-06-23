// mi_5mins.go 對應 `/afterTrading/MI_5MINS` 端點。
// 用途:每 5 秒委託成交統計(累積委買賣筆數/張數、累計成交)。
// 對應 README.tsme.md「盤後交易資訊」第 6 個端點。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/afterTrading/MI_5MINS?date=20221230&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
)

// MI_5MINSResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type MI_5MINSResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_5MINSResponse) GetStat() string { return r.Response.Stat }

// MI_5MINSRow is a typed representation of one MI_5MINS data row.
type MI_5MINSRow struct {
	Time          string // 時間
	CumBuyOrders  int64  // 累積委買筆數
	CumBuyLots    int64  // 累積委買張數
	CumSellOrders int64  // 累積委賣筆數
	CumSellLots   int64  // 累積委賣張數
	CumTradeOrders int64 // 累計成交筆數
	CumTradeLots  int64  // 累計成交張數
}

// FetchMI_5MINS retrieves the every-5-seconds order/trade statistics for `date`.
func FetchMI_5MINS(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_5MINS: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_5MINSResponse](ctx, c, "/afterTrading/MI_5MINS", q)
}

// ParseMI_5MINSRow converts one raw `data` row into a typed MI_5MINSRow.
func ParseMI_5MINSRow(row []string) (MI_5MINSRow, error) {
	if len(row) < 7 {
		return MI_5MINSRow{}, fmt.Errorf("MI_5MINS: row too short: %d cols", len(row))
	}
	return MI_5MINSRow{
		Time:           row[0],
		CumBuyOrders:   ParseInt(row[1]),
		CumBuyLots:     ParseInt(row[2]),
		CumSellOrders:  ParseInt(row[3]),
		CumSellLots:    ParseInt(row[4]),
		CumTradeOrders: ParseInt(row[5]),
		CumTradeLots:   ParseInt(row[6]),
	}, nil
}