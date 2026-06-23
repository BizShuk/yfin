package twse

import (
	"context"
	"fmt"
	"net/url"

	"github.com/AmpyFin/yfinance-go/internal/httpx"
)

// TWTB4UResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type TWTB4UResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *TWTB4UResponse) GetStat() string { return r.Response.Stat }

// TWTB4URow is a typed representation of one TWTB4U data row.
type TWTB4URow struct {
	Code        string // 證券代號
	Name        string // 證券名稱
	TradeShares int64  // 當日沖銷交易成交股數
	TradeAmount int64  // 當日沖銷交易成交金額
	BuyAmount   int64  // 當日沖銷交易買進成交金額
	SellAmount  int64  // 當日沖銷交易賣出成交金額
}

// FetchTWTB4U retrieves the daily day-trade targets and statistics for `date`.
func FetchTWTB4U(ctx context.Context, c *httpx.Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/TWTB4U: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[TWTB4UResponse](ctx, c, "/afterTrading/TWTB4U", q)
}

// ParseTWTB4URow converts one raw `data` row into a typed TWTB4URow.
func ParseTWTB4URow(row []string) (TWTB4URow, error) {
	if len(row) < 6 {
		return TWTB4URow{}, fmt.Errorf("TWTB4U: row too short: %d cols", len(row))
	}
	return TWTB4URow{
		Code:        row[0],
		Name:        row[1],
		TradeShares: ParseInt(row[2]),
		TradeAmount: ParseInt(row[3]),
		BuyAmount:   ParseInt(row[4]),
		SellAmount:  ParseInt(row[5]),
	}, nil
}