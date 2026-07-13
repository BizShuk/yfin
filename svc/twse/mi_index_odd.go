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
	MI_INDEX_ODDResponse = model.MI_INDEX_ODDResponse
	MIIndexOddRow = model.MIIndexOddRow
)

// MI_INDEX_ODDResponse embeds the common Response envelope and adds
// the `date` field that TWSE returns on this endpoint.

// MIIndexOddRow is a typed representation of one MI_INDEX_ODD data row.
// Fields: 證券代號, 證券名稱, 成交股數, 成交金額, 開盤, 最高, 最低, 收盤.

// FetchMI_INDEX_ODD retrieves the odd-lot (零股) trading snapshot for
// `date`.
func FetchMI_INDEX_ODD(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_INDEX_ODD: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MI_INDEX_ODDResponse](ctx, client, "/afterTrading/MI_INDEX_ODD", q)
}

// ParseMIIndexOddRow converts one raw `data` row into a typed
// MIIndexOddRow.
func ParseMIIndexOddRow(row []string) (MIIndexOddRow, error) {
	if len(row) < 8 {
		return MIIndexOddRow{}, fmt.Errorf("MI_INDEX_ODD: row too short: %d cols", len(row))
	}
	return MIIndexOddRow{
		Code:   strings.TrimSpace(row[0]),
		Name:   strings.TrimSpace(row[1]),
		Volume: ParseInt(row[2]),
		Amount: ParseInt(row[3]),
		Open:   ParseFloat(row[4]),
		High:   ParseFloat(row[5]),
		Low:    ParseFloat(row[6]),
		Close:  ParseFloat(row[7]),
	}, nil
}
