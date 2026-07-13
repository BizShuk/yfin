package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)

// Type aliases — structs now live in model/twse.go.
type (
	MI_MARGNResponse = model.MI_MARGNResponse
	MI_MARGNRow = model.MI_MARGNRow
)

// MI_MARGNResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.

// MI_MARGNRow is a typed representation of one MI_MARGN data row.

// FetchMI_MARGN retrieves the margin trading balances for `date`.
// selectType=ALL is always added by this fetcher.
func FetchMI_MARGN(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_MARGN: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_MARGNResponse](ctx, client, "/marginTrading/MI_MARGN", q)
}

// ParseMI_MARGNRow converts one raw `data` row into a typed MI_MARGNRow.
func ParseMI_MARGNRow(row []string) (MI_MARGNRow, error) {
	if len(row) < 10 {
		return MI_MARGNRow{}, fmt.Errorf("MI_MARGN: row too short: %d cols", len(row))
	}
	return MI_MARGNRow{
		Code:          row[0],
		Name:          row[1],
		MarginBuy:     ParseInt(row[2]),
		MarginSell:    ParseInt(row[3]),
		MarginRepay:   ParseInt(row[4]),
		MarginBalance: ParseInt(row[5]),
		ShortBuy:      ParseInt(row[6]),
		ShortSell:     ParseInt(row[7]),
		ShortRepay:    ParseInt(row[8]),
		ShortBalance:  ParseInt(row[9]),
	}, nil
}
