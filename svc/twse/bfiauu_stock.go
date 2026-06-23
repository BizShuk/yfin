// bfiauu_stock.go 對應 `BFIAUU_STOCK` 端點。
// 用途:單一證券日成交資訊(透過 stockNo 過濾鉅額交易)。
// 對應 README.tsme.md「鉅額交易」章節;需指定 stockNo。
// 範例:
//   curl "https://www.twse.com.tw/rwd/zh/block/BFIAUU?date=20221230&stockNo=2330&response=json"

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
func FetchBFIAUUSTOCK(ctx context.Context, c Caller, date string, opts url.Values) (any, error) {
	if opts.Get("stockNo") == "" {
		return nil, fmt.Errorf("twse/BFIAUU_STOCK: stockNo is required")
	}
	return FetchBlockBFIAUU(ctx, c, date, opts)
}

// ParseBFIAUUSTOCKRow parses a row from the BFIAUU_STOCK endpoint. The
// data shape is identical to BFIAUU, so it delegates to ParseBlockBFIAUURow.
func ParseBFIAUUSTOCKRow(row []string) (BlockBFIAUURow, error) {
	return ParseBlockBFIAUURow(row)
}
