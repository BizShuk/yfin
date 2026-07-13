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
	BlockBFIAUUResponse = model.BlockBFIAUUResponse
	BlockBFIAUURow = model.BlockBFIAUURow
)

// BlockBFIAUUResponse embeds the common Response envelope and adds
// the `date` and optional `stockNo` fields that TWSE returns on the
// /block/BFIAUU block-trade endpoint.


// BlockBFIAUURow is a typed representation of one /block/BFIAUU data row.
// Fields: 序號, 證券代號, 證券名稱, 買進證券商, 賣出證券商, 成交數量, 成交金額, 成交價格, 成交時間, 買進成交價.

// FetchBlockBFIAUU retrieves block-trade (鉅額交易) data for `date`.
// If `stockNo` is set in `opts`, the response is filtered to that symbol.
func FetchBlockBFIAUU(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/BFIAUU: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[BlockBFIAUUResponse](ctx, client, "/block/BFIAUU", q)
}

// ParseBlockBFIAUURow converts one raw `data` row into a typed BlockBFIAUURow.
func ParseBlockBFIAUURow(row []string) (BlockBFIAUURow, error) {
	if len(row) < 10 {
		return BlockBFIAUURow{}, fmt.Errorf("BFIAUU: row too short: %d cols", len(row))
	}
	return BlockBFIAUURow{
		Seq:           strings.TrimSpace(row[0]),
		StockCode:     strings.TrimSpace(row[1]),
		StockName:     strings.TrimSpace(row[2]),
		BuyBroker:     strings.TrimSpace(row[3]),
		SellBroker:    strings.TrimSpace(row[4]),
		TradeVolume:   ParseInt(row[5]),
		TradeAmount:   ParseFloat(row[6]),
		TradePrice:    ParseFloat(row[7]),
		TradeTime:     strings.TrimSpace(row[8]),
		BuyTradePrice: ParseFloat(row[9]),
	}, nil
}
