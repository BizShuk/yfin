package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)



// FetchTWT44U retrieves the daily aggregated buy/sell volume of
// dealers (自營商) for `date`.
func FetchTWT44U(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/TWT44U: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.TWT44UResponse](ctx, client, "/fund/TWT44U", q)
}

// ParseTWT44URow converts one raw `data` row into a typed TWT44URow.
func ParseTWT44URow(row []string) (model.TWT44URow, error) {
	if len(row) < 4 {
		return model.TWT44URow{}, fmt.Errorf("TWT44U: row too short: %d cols", len(row))
	}
	return model.TWT44URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseInt(row[1]),
		Sell:     ParseInt(row[2]),
		Net:      ParseInt(row[3]),
	}, nil
}
