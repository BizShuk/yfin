package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)


// FetchBFIMUU retrieves the monthly block-trade report for `date` (YYYYMM01).
func FetchBFIMUU(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFIMUU: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.BFIMUResponse](ctx, client, "/block/BFIMUU", q)
}

// ParseBFIMUURow converts one raw `data` row into a typed BFIMUURow.
func ParseBFIMUURow(row []string) (model.BFIMUURow, error) {
	if len(row) < 4 {
		return model.BFIMUURow{}, fmt.Errorf("BFIMUU: row too short: %d cols", len(row))
	}
	return model.BFIMUURow{
		Period:       row[0],
		Transactions: ParseInt(row[1]),
		Volume:       ParseInt(row[2]),
		Amount:       ParseInt(row[3]),
	}, nil
}
