package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)


// FetchFMSRFK retrieves per-stock monthly trading info for the year `date`.
// `stockNo` is required (e.g. "2330"); `date` is the year (e.g. "2022").
func FetchFMSRFK(ctx context.Context, client *Client, stockNo, date string, opts url.Values) (any, error) {
	if stockNo == "" {
		return nil, fmt.Errorf("twse/FMSRFK: stockNo is required")
	}
	if date == "" {
		return nil, fmt.Errorf("twse/FMSRFK: date is required")
	}
	q := url.Values{}
	q.Set("stockNo", stockNo)
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.FMSRFKResponse](ctx, client, "/exchangeReport/FMSRFK", q)
}

// ParseFMSRFKRow converts one raw `data` row into a typed FMSRFKRow.
func ParseFMSRFKRow(row []string) (model.FMSRFKRow, error) {
	if len(row) < 8 {
		return model.FMSRFKRow{}, fmt.Errorf("FMSRFK: row too short: %d cols", len(row))
	}
	return model.FMSRFKRow{
		Year:        strings.TrimSpace(row[0]),
		Month:       strings.TrimSpace(row[1]),
		High:        ParseFloat(row[2]),
		Low:         ParseFloat(row[3]),
		WAvgPrice:   ParseFloat(row[4]),
		TradeVolume: ParseInt(row[5]),
		TradeValue:  ParseInt(row[6]),
		TurnoverPct: ParseFloat(row[7]),
	}, nil
}
