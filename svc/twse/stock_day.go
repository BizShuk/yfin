package twse

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"github.com/bizshuk/yfin/model"
)


// FetchSTOCK_DAY retrieves per-stock daily trade info for `date` and
// `stockNo` (must be supplied via opts).
func FetchSTOCK_DAY(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/STOCK_DAY: date is required")
	}
	stockNo := opts.Get("stockNo")
	if stockNo == "" {
		return nil, fmt.Errorf("twse/STOCK_DAY: stockNo is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("stockNo", stockNo)
	for k, vs := range opts {
		if k == "stockNo" {
			continue
		}
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[model.STOCK_DAYResponse](ctx, client, "/afterTrading/STOCK_DAY", q)
}

// ParseStockDayRow converts one raw `data` row into a typed StockDayRow.
func ParseStockDayRow(row []string) (model.StockDayRow, error) {
	if len(row) < 9 {
		return model.StockDayRow{}, fmt.Errorf("STOCK_DAY: row too short: %d cols", len(row))
	}
	return model.StockDayRow{
		Date:         strings.TrimSpace(row[0]),
		Volume:       ParseInt(row[1]),
		Amount:       ParseInt(row[2]),
		Open:         ParseFloat(row[3]),
		High:         ParseFloat(row[4]),
		Low:          ParseFloat(row[5]),
		Close:        ParseFloat(row[6]),
		Change:       ParseFloat(row[7]),
		Transactions: ParseInt(row[8]),
	}, nil
}
