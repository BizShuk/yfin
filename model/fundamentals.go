// fundamentals.go — `FundamentalsSnapshot` / `FundamentalsLine` plain data
// structs for the SDK surface. Originally lived in facade/fundamentals.go;
// promoted to model/ for cross-layer reuse.

package model

import "time"

// FundamentalsLine is one period-tagged line item (e.g., "revenue",
// "net_income", "eps_basic"). Value is the float64 decoded from the
// internal ScaledDecimal / ampy-proto Decimal; a missing Value is surfaced
// as 0 (float64 cannot distinguish zero from missing). PeriodStart /
// PeriodEnd are UTC.
type FundamentalsLine struct {
	Key          string    `json:"key"`
	Value        float64   `json:"value"`
	CurrencyCode string    `json:"currency_code,omitempty"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
}

// FundamentalsSnapshot is one symbol's fundamentals view, exposed as a plain
// SDK struct. Lines preserves the order it was supplied in (the ampy-proto
// FundamentalsSnapshot.Lines is a []*LineItem, not a map, so ordering is
// inherent).
type FundamentalsSnapshot struct {
	Symbol string             `json:"symbol"`
	MIC    string             `json:"mic,omitempty"`
	Source string             `json:"source,omitempty"`
	AsOf   time.Time          `json:"as_of"`
	Lines  []FundamentalsLine `json:"lines,omitempty"`
}
