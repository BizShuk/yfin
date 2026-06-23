// mi_margn.go 對應 `/marginTrading/MI_MARGN` 端點。
// 用途:融資融券餘額(融資/融券買賣、償還、餘額)。
// 對應 README.tsme.md「融資融券」章節。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/marginTrading/MI_MARGN?date=20221230&selectType=ALL&response=json"

package twse

import (
	"context"
	"fmt"
	"net/url"
)

// MI_MARGNResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type MI_MARGNResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_MARGNResponse) GetStat() string { return r.Response.Stat }

// MI_MARGNRow is a typed representation of one MI_MARGN data row.
type MI_MARGNRow struct {
	Code          string // 股票代號
	Name          string // 股票名稱
	MarginBuy     int64  // 融資買進
	MarginSell    int64  // 融資賣出
	MarginRepay   int64  // 融資現償
	MarginBalance int64  // 融資餘額
	ShortBuy      int64  // 融券買進
	ShortSell     int64  // 融券賣出
	ShortRepay    int64  // 融券現償
	ShortBalance  int64  // 融券餘額
}

// FetchMI_MARGN retrieves the margin trading balances for `date`.
// selectType=ALL is always added by this fetcher.
func FetchMI_MARGN(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_MARGN: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_MARGNResponse](ctx, c, "/marginTrading/MI_MARGN", q)
}

// ParseMI_MARGNRow converts one raw `data` row into a typed MI_MARGNRow.
func ParseMI_MARGNRow(row []string) (MI_MARGNRow, error) {
	if len(row) < 10 {
		return MI_MARGNRow{}, fmt.Errorf("MI_MARGN: row too short: %d cols", len(row))
	}
	return MI_MARGNRow{
		Code:          row[0],
		Name:          row[1],
		MarginBuy:     ParseInt(row[2]),
		MarginSell:    ParseInt(row[3]),
		MarginRepay:   ParseInt(row[4]),
		MarginBalance: ParseInt(row[5]),
		ShortBuy:      ParseInt(row[6]),
		ShortSell:     ParseInt(row[7]),
		ShortRepay:    ParseInt(row[8]),
		ShortBalance:  ParseInt(row[9]),
	}, nil
}