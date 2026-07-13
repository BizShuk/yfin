package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)

// Type aliases — structs now live in model/twse.go.
type (
	BFIAUUYEARResponse = model.BFIAUUYEARResponse
	BFIAUUYEARRow = model.BFIAUUYEARRow
)

// BFIAUUYEARResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.

// BFIAUUYEARRow is a typed representation of one BFIAUU_YEAR data row.
// Fields: 年度, 成交筆數, 成交股數, 成交金額.

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
	return FetchJSON[BFIAUUYEARResponse](ctx, client, "/block/BFIAUU_YEAR", q)
}

// ParseBFIAUUYEARRow converts one raw `data` row into a typed BFIAUUYEARRow.
func ParseBFIAUUYEARRow(row []string) (BFIAUUYEARRow, error) {
	if len(row) < 4 {
		return BFIAUUYEARRow{}, fmt.Errorf("BFIAUU_YEAR: row too short: %d cols", len(row))
	}
	return BFIAUUYEARRow{
		Year:         row[0],
		Transactions: ParseInt(row[1]),
		Volume:       ParseInt(row[2]),
		Amount:       ParseInt(row[3]),
	}, nil
}
