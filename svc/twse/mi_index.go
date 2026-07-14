package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)

// FetchMI_INDEX retrieves the daily market index close for `date`.
// `opts` may include a `type=ALL` parameter (TWSE expects this).
func FetchMI_INDEX(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_INDEX: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("type", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.MI_INDEXResponse](ctx, client, "/afterTrading/MI_INDEX", q)
}

// ParseMIIndexRow converts one raw `data` row into a typed model.MIIndexRow.
func ParseMIIndexRow(row []string) (model.MIIndexRow, error) {
	if len(row) < 4 {
		return model.MIIndexRow{}, fmt.Errorf("MI_INDEX: row too short: %d cols", len(row))
	}
	return model.MIIndexRow{
		IndexName: strings.TrimSpace(row[0]),
		Close:     ParseFloat(row[1]),
		Change:    ParseFloat(row[2]),
		ChangePct: ParsePercent(row[3]),
	}, nil
}
