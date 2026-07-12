// bars.go — `Bar` / `BarBatch` plain data structs for the SDK surface.
// Float64 prices (decoded from internal ScaledDecimal by facade converters).
// Originally lived in facade/bars.go; promoted to model/ so any layer
// (cmd, facade, svc, external consumers like stock/data) can depend on
// the type without pulling in svc/norm.

package model

// Bar is one daily OHLCV bar with float64 prices (decoded from ScaledDecimal).
type Bar struct {
	Date         string  `json:"date"` // YYYY-MM-DD (UTC) — derived from Bar.EventTime
	Open         float64 `json:"open"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
	Close        float64 `json:"close"`
	Volume       int64   `json:"volume"`
	Adjusted     bool    `json:"adjusted"`
	CurrencyCode string  `json:"currency_code"`
}

// BarBatch is a series of Bars for one symbol, plus identifying metadata.
type BarBatch struct {
	Symbol string `json:"symbol"`
	MIC    string `json:"mic,omitempty"`
	Bars   []Bar  `json:"bars"`
}
