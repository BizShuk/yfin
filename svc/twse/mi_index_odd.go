package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)


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
	return FetchJSON[model.MI_INDEX_ODDResponse](ctx, client, "/afterTrading/MI_INDEX_ODD", q)
}

// ParseMIIndexOddRow converts one raw `data` row into a typed
// MIIndexOddRow.
func ParseMIIndexOddRow(row []string) (model.MIIndexOddRow, error) {
	if len(row) < 8 {
		return model.MIIndexOddRow{}, fmt.Errorf("MI_INDEX_ODD: row too short: %d cols", len(row))
	}
	return model.MIIndexOddRow{
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
