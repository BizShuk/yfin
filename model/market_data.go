// market_data.go — `MarketData` plain data struct for the SDK surface.
// Nullable `*float64` price fields + `*int64` volume; nil = missing, not zero.
// Originally lived in facade/market_data.go; promoted to model/ for cross-layer reuse.

package model

import "time"

// MarketData is a plain Go snapshot of a security's current market state, as
// exposed by the SDK facade. Prices are returned as float64 (decoded from the
// internal ScaledDecimal); nil *float64/*int64 fields indicate the value was
// not reported by the source (e.g., after-hours).
//
// Times are UTC. Volume is share count, not currency-denominated.
type MarketData struct {
	Symbol              string    `json:"symbol"`
	MIC                 string    `json:"mic,omitempty"`
	RegularMarketPrice  *float64  `json:"regular_market_price,omitempty"`
	RegularMarketHigh   *float64  `json:"regular_market_high,omitempty"`
	RegularMarketLow    *float64  `json:"regular_market_low,omitempty"`
	RegularMarketVolume *int64    `json:"regular_market_volume,omitempty"`
	FiftyTwoWeekHigh    *float64  `json:"fifty_two_week_high,omitempty"`
	FiftyTwoWeekLow     *float64  `json:"fifty_two_week_low,omitempty"`
	PreviousClose       *float64  `json:"previous_close,omitempty"`
	CurrencyCode        string    `json:"currency_code,omitempty"`
	EventTime           time.Time `json:"event_time"`
}
