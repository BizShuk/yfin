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
	TWT44UResponse = model.TWT44UResponse
	TWT44URow = model.TWT44URow
)

// TWT44UResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.


// TWT44URow is a typed representation of one TWT44U data row.
// Fields: 單位名稱, 買進股數, 賣出股數, 買賣差額股數.

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
	return FetchJSON[TWT44UResponse](ctx, client, "/fund/TWT44U", q)
}

// ParseTWT44URow converts one raw `data` row into a typed TWT44URow.
func ParseTWT44URow(row []string) (TWT44URow, error) {
	if len(row) < 4 {
		return TWT44URow{}, fmt.Errorf("TWT44U: row too short: %d cols", len(row))
	}
	return TWT44URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseInt(row[1]),
		Sell:     ParseInt(row[2]),
		Net:      ParseInt(row[3]),
	}, nil
}
