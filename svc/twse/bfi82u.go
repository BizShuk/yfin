package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)



// FetchBFI82U retrieves the daily aggregated buy/sell amounts of the
// three main institutional investors (自營商, 投信, 外資及陸資) for `date`.
// TWSE requires `type=day` for this endpoint.
func FetchBFI82U(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFI82U: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("type", "day")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.BFI82UResponse](ctx, client, "/fund/BFI82U", q)
}

// ParseBFI82URow converts one raw `data` row into a typed BFI82URow.
func ParseBFI82URow(row []string) (model.BFI82URow, error) {
	if len(row) < 4 {
		return model.BFI82URow{}, fmt.Errorf("BFI82U: row too short: %d cols", len(row))
	}
	return model.BFI82URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseFloat(row[1]),
		Sell:     ParseFloat(row[2]),
		Net:      ParseFloat(row[3]),
	}, nil
}
