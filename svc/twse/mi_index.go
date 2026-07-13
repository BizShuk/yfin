package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)

// Type aliases — structs now live in model/twse.go.
type (
	MI_INDEXResponse = model.MI_INDEXResponse
	MIIndexRow = model.MIIndexRow
)

// MI_INDEXResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.

// MIIndexRow is a typed representation of one MI_INDEX data row.
// Fields: 指數, 收盤指數, 漲跌點數, 漲跌百分比.

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
	return FetchJSON[MI_INDEXResponse](ctx, client, "/afterTrading/MI_INDEX", q)
}

// ParseMIIndexRow converts one raw `data` row into a typed MIIndexRow.
func ParseMIIndexRow(row []string) (MIIndexRow, error) {
	if len(row) < 4 {
		return MIIndexRow{}, fmt.Errorf("MI_INDEX: row too short: %d cols", len(row))
	}
	return MIIndexRow{
		IndexName: strings.TrimSpace(row[0]),
		Close:     ParseFloat(row[1]),
		Change:    ParseFloat(row[2]),
		ChangePct: ParsePercent(row[3]),
	}, nil
}
