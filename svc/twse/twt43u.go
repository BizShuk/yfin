package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)



// FetchTWT43U retrieves the daily aggregated buy/sell volume of
// investment trust companies (投信) for `date`.
func FetchTWT43U(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/TWT43U: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.TWT43UResponse](ctx, client, "/fund/TWT43U", q)
}

// ParseTWT43URow converts one raw `data` row into a typed TWT43URow.
func ParseTWT43URow(row []string) (model.TWT43URow, error) {
	if len(row) < 4 {
		return model.TWT43URow{}, fmt.Errorf("TWT43U: row too short: %d cols", len(row))
	}
	return model.TWT43URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseInt(row[1]),
		Sell:     ParseInt(row[2]),
		Net:      ParseInt(row[3]),
	}, nil
}
