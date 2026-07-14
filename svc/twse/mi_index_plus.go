package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)


// FetchMI_INDEX_PLUS retrieves the after-hours (盤後定價) index data
// for `date`.
func FetchMI_INDEX_PLUS(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_INDEX_PLUS: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.MI_INDEX_PLUSResponse](ctx, client, "/afterTrading/MI_INDEX_PLUS", q)
}

// ParseMIIndexPlusRow converts one raw `data` row into a typed
// MIIndexPlusRow.
func ParseMIIndexPlusRow(row []string) (model.MIIndexPlusRow, error) {
	if len(row) < 4 {
		return model.MIIndexPlusRow{}, fmt.Errorf("MI_INDEX_PLUS: row too short: %d cols", len(row))
	}
	return model.MIIndexPlusRow{
		IndexName: strings.TrimSpace(row[0]),
		Close:     ParseFloat(row[1]),
		Change:    ParseFloat(row[2]),
		ChangePct: ParsePercent(row[3]),
	}, nil
}
