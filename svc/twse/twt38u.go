package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)

// Type aliases — structs now live in model/twse.go.
type (
	TWT38UResponse = model.TWT38UResponse
	TWT38URow = model.TWT38URow
)

// TWT38UResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.


// TWT38URow is a typed representation of one TWT38U data row.
// Fields: 單位名稱, 買進股數, 賣出股數, 買賣差額股數.

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
	return FetchJSON[TWT38UResponse](ctx, client, "/fund/TWT38U", q)
}

// ParseTWT38URow converts one raw `data` row into a typed TWT38URow.
func ParseTWT38URow(row []string) (TWT38URow, error) {
	if len(row) < 4 {
		return TWT38URow{}, fmt.Errorf("TWT38U: row too short: %d cols", len(row))
	}
	return TWT38URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseInt(row[1]),
		Sell:     ParseInt(row[2]),
		Net:      ParseInt(row[3]),
	}, nil
}
