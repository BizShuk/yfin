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
	BFI82UResponse = model.BFI82UResponse
	BFI82URow = model.BFI82URow
)

// BFI82UResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.


// BFI82URow is a typed representation of one BFI82U data row.
// Fields: 單位名稱, 買進金額, 賣出金額, 買賣差額.

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
	return FetchJSON[BFI82UResponse](ctx, client, "/fund/BFI82U", q)
}

// ParseBFI82URow converts one raw `data` row into a typed BFI82URow.
func ParseBFI82URow(row []string) (BFI82URow, error) {
	if len(row) < 4 {
		return BFI82URow{}, fmt.Errorf("BFI82U: row too short: %d cols", len(row))
	}
	return BFI82URow{
		UnitName: strings.TrimSpace(row[0]),
		Buy:      ParseFloat(row[1]),
		Sell:     ParseFloat(row[2]),
		Net:      ParseFloat(row[3]),
	}, nil
}
