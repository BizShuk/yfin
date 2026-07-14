package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)



// FetchBWIBBU_d retrieves the per-stock P/E, dividend yield, and P/B
// ratio snapshot for `date`. `opts` may include `selectType=ALL`
// (TWSE expects this).
func FetchBWIBBU_d(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BWIBBU_d: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("selectType", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.BWIBBU_dResponse](ctx, client, "/afterTrading/BWIBBU_d", q)
}

// ParseBWIBBUdRow converts one raw `data` row into a typed BWIBBUdRow.
func ParseBWIBBUdRow(row []string) (model.BWIBBUdRow, error) {
	if len(row) < 5 {
		return model.BWIBBUdRow{}, fmt.Errorf("BWIBBU_d: row too short: %d cols", len(row))
	}
	return model.BWIBBUdRow{
		Code:     strings.TrimSpace(row[0]),
		Name:     strings.TrimSpace(row[1]),
		PE:       ParseFloat(row[2]),
		YieldPct: ParsePercent(row[3]),
		PBR:      ParseFloat(row[4]),
	}, nil
}
