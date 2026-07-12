// market_data.go — `MarketData` type alias + `FromMarketData`
// ScaledDecimal → `*float64` converter + `scaledDecimalPtrToFloatPtr` helper.
// Struct lives in model/market_data.go; facade.MarketData is a back-compat alias.
package facade

import (
	"github.com/bizshuk/yfin/model"
)

// MarketData is a plain Go snapshot of a security's current market state, as
// exposed by the SDK facade. Aliased from model.MarketData — new code should
// use model.MarketData directly.
type MarketData = model.MarketData

// FromMarketData converts an internal model.NormalizedMarketData into the
// plain SDK MarketData struct. Each *ScaledDecimal is unwrapped via
// model.FromScaledDecimal to a *float64 (nil stays nil — callers can rely on
// the distinction between "missing" and "zero"). Returns nil for a nil input.
func FromMarketData(m *model.NormalizedMarketData) *MarketData {
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

// scaledDecimalPtrToFloatPtr unwraps a *model.ScaledDecimal to a *float64 via
// model.FromScaledDecimal. nil in -> nil out; nil-valued SD is not reachable
// here (the caller already checked the outer pointer).
func scaledDecimalPtrToFloatPtr(sd *model.ScaledDecimal) *float64 {
	if sd == nil {
		return nil
	}
	v := model.FromScaledDecimal(*sd)
	return &v
}
