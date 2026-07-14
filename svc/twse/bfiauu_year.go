package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)


// FetchBFIAUUYEAR retrieves the annual block-trade report for `date` (YYYY0101).
func FetchBFIAUUYEAR(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFIAUU_YEAR: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.BFIAUUYEARResponse](ctx, client, "/block/BFIAUU_YEAR", q)
}

// ParseBFIAUUYEARRow converts one raw `data` row into a typed BFIAUUYEARRow.
func ParseBFIAUUYEARRow(row []string) (model.BFIAUUYEARRow, error) {
	if len(row) < 4 {
		return model.BFIAUUYEARRow{}, fmt.Errorf("BFIAUU_YEAR: row too short: %d cols", len(row))
	}
	return model.BFIAUUYEARRow{
		Year:         row[0],
		Transactions: ParseInt(row[1]),
		Volume:       ParseInt(row[2]),
		Amount:       ParseInt(row[3]),
	}, nil
}
