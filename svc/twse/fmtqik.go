package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)


// FetchFMTQIK retrieves the TAIEX index and trading volume for `date`.
// `date` should be YYYYMMDD (month-start or month-end).
func FetchFMTQIK(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/FMTQIK: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.FMTQIKResponse](ctx, client, "/exchangeReport/FMTQIK", q)
}

// ParseFMTQIKRow converts one raw `data` row into a typed FMTQIKRow.
func ParseFMTQIKRow(row []string) (model.FMTQIKRow, error) {
	if len(row) < 5 {
		return model.FMTQIKRow{}, fmt.Errorf("FMTQIK: row too short: %d cols", len(row))
	}
	return model.FMTQIKRow{
		Date:         row[0],
		Volume:       ParseInt(row[1]),
		Amount:       ParseInt(row[2]),
		Transactions: ParseInt(row[3]),
		Index:        ParseFloat(row[4]),
	}, nil
}
