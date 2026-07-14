package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)



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
	return FetchJSON[model.T86Response](ctx, client, "/fund/T86", q)
}

// ParseT86Row converts one raw `data` row into a typed T86Row.
func ParseT86Row(row []string) (model.T86Row, error) {
	if len(row) < 12 {
		return model.T86Row{}, fmt.Errorf("T86: row too short: %d cols", len(row))
	}
	return model.T86Row{
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
