package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)



// FetchTWT38U retrieves the daily aggregated buy/sell volume of
// foreign investors (含陸資) for `date`.
func FetchTWT38U(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/TWT38U: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.TWT38UResponse](ctx, client, "/fund/TWT38U", q)
}

// ParseTWT38URow converts one raw `data` row into a typed TWT38URow.
func ParseTWT38URow(row []string) (model.TWT38URow, error) {
	if len(row) < 4 {
		return model.TWT38URow{}, fmt.Errorf("TWT38U: row too short: %d cols", len(row))
	}
	return model.TWT38URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseInt(row[1]),
		Sell:     ParseInt(row[2]),
		Net:      ParseInt(row[3]),
	}, nil
}
