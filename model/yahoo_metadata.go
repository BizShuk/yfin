// yahoo_metadata.go — Yahoo chart-metadata DTO (subset of ChartMeta used by
// the historical 1-day chart shortcut that powers Python yfinance's
// get_history_metadata). Originally lived in svc/yahoo/metadata.go; promoted
// to model/ so external consumers can depend on the shape without importing
// the Extract/Fetch behavior of svc/yahoo.

package model

// ChartMetadata is the slimmed chart-meta view returned by the 1-day chart call.
type ChartMetadata struct {
	Symbol             string  `json:"symbol"`
	Currency           string  `json:"currency"`
	ExchangeName       string  `json:"exchangeName"`
	InstrumentType     string  `json:"instrumentType"`
	Timezone           string  `json:"timezone"`
	GmtOffset          int     `json:"gmtoffset"`
	FirstTradeDate     int64   `json:"firstTradeDate"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}