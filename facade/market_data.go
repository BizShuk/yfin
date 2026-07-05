// market_data.go — `MarketData` plain SDK struct (nullable `*float64` price
// fields + `*int64` volume; nil = missing, not zero) + `FromMarketData`
// ScaledDecimal → `*float64` converter + `scaledDecimalPtrToFloatPtr` helper.
// Capacity: 1 struct + 1 converter + 1 helper.
package facade

import (
	"time"

	"github.com/bizshuk/yfin/svc/norm"
)

// MarketData is a plain Go snapshot of a security's current market state, as
// exposed by the SDK facade. Prices are returned as float64 (decoded from the
// internal ScaledDecimal via norm.FromScaledDecimal); nil *float64/*int64 fields
// indicate the value was not reported by the source (e.g., after-hours).
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

// FromMarketData converts an internal norm.NormalizedMarketData into the
// plain SDK MarketData struct. Each *ScaledDecimal is unwrapped via
// norm.FromScaledDecimal to a *float64 (nil stays nil — callers can rely on
// the distinction between "missing" and "zero"). Returns nil for a nil input.
func FromMarketData(m *norm.NormalizedMarketData) *MarketData {
	if m == nil {
		return nil
	}
	return &MarketData{
		Symbol:              m.Security.Symbol,
		MIC:                 m.Security.MIC,
		RegularMarketPrice:  scaledDecimalPtrToFloatPtr(m.RegularMarketPrice),
		RegularMarketHigh:   scaledDecimalPtrToFloatPtr(m.RegularMarketHigh),
		RegularMarketLow:    scaledDecimalPtrToFloatPtr(m.RegularMarketLow),
		RegularMarketVolume: m.RegularMarketVolume,
		FiftyTwoWeekHigh:    scaledDecimalPtrToFloatPtr(m.FiftyTwoWeekHigh),
		FiftyTwoWeekLow:     scaledDecimalPtrToFloatPtr(m.FiftyTwoWeekLow),
		PreviousClose:       scaledDecimalPtrToFloatPtr(m.PreviousClose),
		CurrencyCode:        m.CurrencyCode,
		EventTime:           m.EventTime,
	}
}

// scaledDecimalPtrToFloatPtr unwraps a *norm.ScaledDecimal to a *float64 via
// norm.FromScaledDecimal. nil in -> nil out; nil-valued SD is not reachable
// here (the caller already checked the outer pointer).
func scaledDecimalPtrToFloatPtr(sd *norm.ScaledDecimal) *float64 {
	if sd == nil {
		return nil
	}
	v := norm.FromScaledDecimal(*sd)
	return &v
}
