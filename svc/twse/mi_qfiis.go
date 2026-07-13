package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)

// Type aliases — structs now live in model/twse.go.
type (
	MI_QFIISResponse = model.MI_QFIISResponse
	MI_QFIISRow = model.MI_QFIISRow
)

// MI_QFIISResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.

// MI_QFIISRow is a typed representation of one MI_QFIIS data row.

// FetchMI_QFIIS retrieves the foreign+mainland investor holdings for `date`.
// selectType=ALL is always added by this fetcher.
func FetchMI_QFIIS(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_QFIIS: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_QFIISResponse](ctx, client, "/fund/MI_QFIIS", q)
}

// ParseMI_QFIISRow converts one raw `data` row into a typed MI_QFIISRow.
func ParseMI_QFIISRow(row []string) (MI_QFIISRow, error) {
	if len(row) < 4 {
		return MI_QFIISRow{}, fmt.Errorf("MI_QFIIS: row too short: %d cols", len(row))
	}
	return MI_QFIISRow{
		Code:       row[0],
		Name:       row[1],
		SharesHeld: ParseInt(row[2]),
		IssuePct:   ParsePercent(row[3]),
	}, nil
}
