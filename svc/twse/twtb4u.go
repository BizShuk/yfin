package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)



// FetchTWTB4U retrieves the daily day-trade targets and statistics for `date`.
func FetchTWTB4U(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
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
	return FetchJSON[model.TWTB4UResponse](ctx, client, "/afterTrading/TWTB4U", q)
}

// ParseTWTB4URow converts one raw `data` row into a typed TWTB4URow.
func ParseTWTB4URow(row []string) (model.TWTB4URow, error) {
	if len(row) < 6 {
		return model.TWTB4URow{}, fmt.Errorf("TWTB4U: row too short: %d cols", len(row))
	}
	return model.TWTB4URow{
		Code:        row[0],
		Name:        row[1],
		TradeShares: ParseInt(row[2]),
		TradeAmount: ParseInt(row[3]),
		BuyAmount:   ParseInt(row[4]),
		SellAmount:  ParseInt(row[5]),
	}, nil
}
