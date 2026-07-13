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
	MI_INDEX_PLUSResponse = model.MI_INDEX_PLUSResponse
	MIIndexPlusRow = model.MIIndexPlusRow
)

// MI_INDEX_PLUSResponse embeds the common Response envelope and adds
// the `date` field that TWSE returns on this endpoint.

// MIIndexPlusRow is a typed representation of one MI_INDEX_PLUS data row.
// Fields: 指數, 收盤指數, 漲跌點數, 漲跌百分比.

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
	return FetchJSON[MI_INDEX_PLUSResponse](ctx, client, "/afterTrading/MI_INDEX_PLUS", q)
}

// ParseMIIndexPlusRow converts one raw `data` row into a typed
// MIIndexPlusRow.
func ParseMIIndexPlusRow(row []string) (MIIndexPlusRow, error) {
	if len(row) < 4 {
		return MIIndexPlusRow{}, fmt.Errorf("MI_INDEX_PLUS: row too short: %d cols", len(row))
	}
	return MIIndexPlusRow{
		IndexName: strings.TrimSpace(row[0]),
		Close:     ParseFloat(row[1]),
		Change:    ParseFloat(row[2]),
		ChangePct: ParsePercent(row[3]),
	}, nil
}
