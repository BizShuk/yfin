// bfiauu_stock.go — `BFIAUU_STOCK` thin wrapper that adds `stockNo` requirement then delegates to `BFIAUU` (same 10-column row shape). Capacity: same variable trades as `BFIAUU`, filtered to one symbol.
package twse

import (
	"context"
	"fmt"
	"net/url"
)

// FetchBFIAUUSTOCK is a thin wrapper over FetchBFIAUU that requires `stockNo`
// to be supplied via opts. It delegates to the unified FetchBFIAUU once the
// parameter is validated. (After consolidating bfiauu_block.go into
// bfiauu.go, the BFIAUU endpoint uses the full 10-column block-trade
// shape, so BFIAUU_STOCK and BFIAUU share the same row parser.)
func FetchBFIAUUSTOCK(ctx context.Context, client *Client, date string, opts url.Values) (any, error) {
	if opts.Get("stockNo") == "" {
		return nil, fmt.Errorf("twse/BFIAUU_STOCK: stockNo is required")
	}
	return FetchBlockBFIAUU(ctx, client, date, opts)
}

// ParseBFIAUUSTOCKRow parses a row from the BFIAUU_STOCK endpoint. The
// data shape is identical to BFIAUU, so it delegates to ParseBlockBFIAUURow.
func ParseBFIAUUSTOCKRow(row []string) (BlockBFIAUURow, error) {
	return ParseBlockBFIAUURow(row)
}
