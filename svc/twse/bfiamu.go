package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)


// FetchBFIAMU retrieves per-day index close & change values for `date`.
// `date` is required (YYYYMMDD).
func FetchBFIAMU(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFIAMU: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.BFIAMUResponse](ctx, client, "/afterTrading/BFIAMU", q)
}

// ParseBFIAMURow converts one raw `data` row into a typed BFIAMURow.
func ParseBFIAMURow(row []string) (model.BFIAMURow, error) {
	if len(row) < 4 {
		return model.BFIAMURow{}, fmt.Errorf("BFIAMU: row too short: %d cols", len(row))
	}
	return model.BFIAMURow{
		IndexName: strings.TrimSpace(row[0]),
		Close:     ParseFloat(row[1]),
		Change:    ParseFloat(row[2]),
		ChangePct: ParsePercent(row[3]),
	}, nil
}
