// types.go — back-compat type alias for `model.Response` + scrape-internal
// constants (statNoData). The Response envelope + GetStat() accessor now
// live in model/twse.go; this file keeps the scrape-local noData sentinel
// used by FetchJSON's no-data detection.
package twse

import "github.com/bizshuk/yfin/model"

// statNoData is TWSE's traditional "no data" message (varies by endpoint, so we
// substring-check). The exact string from TWSE: "很抱歉，沒有符合條件的資料!" plus
// the Latin "No data" variant.
const statNoData = "沒有符合條件的資料"

// Response is the common TWSE JSON envelope. Type alias to model.Response;
// GetStat() method resolves through the alias automatically.
type Response = model.Response