// types.go — TWSE-internal constants (statNoData) used by FetchJSON's
// no-data detection. The Response envelope + GetStat() accessor live in
// model/twse.go; callers reference model.Response directly.
package twse

// statNoData is TWSE's traditional "no data" message (varies by endpoint, so we
// substring-check). The exact string from TWSE: "很抱歉，沒有符合條件的資料!" plus
// the Latin "No data" variant.
const statNoData = "沒有符合條件的資料"