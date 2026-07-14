package twse

import (
	"context"
	"fmt"
	"net/url"
	"github.com/bizshuk/yfin/model"
)


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
	return FetchJSON[model.MI_QFIISResponse](ctx, client, "/fund/MI_QFIIS", q)
}

// ParseMI_QFIISRow converts one raw `data` row into a typed MI_QFIISRow.
func ParseMI_QFIISRow(row []string) (model.MI_QFIISRow, error) {
	if len(row) < 4 {
		return model.MI_QFIISRow{}, fmt.Errorf("MI_QFIIS: row too short: %d cols", len(row))
	}
	return model.MI_QFIISRow{
		Code:       row[0],
		Name:       row[1],
		SharesHeld: ParseInt(row[2]),
		IssuePct:   ParsePercent(row[3]),
	}, nil
}
