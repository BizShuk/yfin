package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)

// Type aliases — structs now live in model/twse.go.
type (
	MI_5MINSResponse = model.MI_5MINSResponse
	MI_5MINSRow = model.MI_5MINSRow
)

// MI_5MINSResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.


// MI_5MINSRow is a typed representation of one MI_5MINS data row.

// FetchMI_5MINS retrieves the every-5-seconds order/trade statistics for `date`.
func FetchMI_5MINS(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
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
	return FetchJSON[MI_5MINSResponse](ctx, client, "/afterTrading/MI_5MINS", q)
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
