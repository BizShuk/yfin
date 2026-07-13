package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)

// Type aliases — structs now live in model/twse.go.
type (
	T86Response = model.T86Response
	T86Row = model.T86Row
)

// T86Response embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.


// T86Row is a typed representation of one T86 data row.

// FetchT86 retrieves the three-institution daily buy/sell for `date`.
// selectType=ALL is always added by this fetcher.
func FetchT86(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/T86: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[T86Response](ctx, client, "/fund/T86", q)
}

// ParseT86Row converts one raw `data` row into a typed T86Row.
func ParseT86Row(row []string) (T86Row, error) {
	if len(row) < 12 {
		return T86Row{}, fmt.Errorf("T86: row too short: %d cols", len(row))
	}
	return T86Row{
		Code:        row[0],
		Name:        row[1],
		ForeignBuy:  ParseInt(row[2]),
		ForeignSell: ParseInt(row[3]),
		ForeignNet:  ParseInt(row[4]),
		TrustBuy:    ParseInt(row[5]),
		TrustSell:   ParseInt(row[6]),
		TrustNet:    ParseInt(row[7]),
		DealerBuy:   ParseInt(row[8]),
		DealerSell:  ParseInt(row[9]),
		DealerNet:   ParseInt(row[10]),
		TotalNet:    ParseInt(row[11]),
	}, nil
}
