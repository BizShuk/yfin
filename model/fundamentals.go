// fundamentals.go — `FundamentalsSnapshot` / `FundamentalsLine` plain data
// structs for the SDK surface. Originally lived in facade/fundamentals.go;
// promoted to model/ for cross-layer reuse.

package model

import "time"

// FundamentalsLine is one period-tagged line item (e.g., "revenue",
// "net_income", "eps_basic"). Value is the float64-decoded normalized value;
// a missing Value is surfaced as 0 because float64 cannot distinguish it from
// a reported zero. An absent provider currency remains empty. PeriodStart and
// PeriodEnd are UTC.
type FundamentalsLine struct {
	Key          string    `json:"key"`
	Value        float64   `json:"value"`
	CurrencyCode string    `json:"currency_code,omitempty"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
}

// FundamentalsSnapshot is one symbol's fundamentals view, exposed as a plain
// SDK struct. Lines preserves provider/normalizer order.
type FundamentalsSnapshot struct {
	Symbol string             `json:"symbol"`
	MIC    string             `json:"mic,omitempty"`
	Source string             `json:"source,omitempty"`
	AsOf   time.Time          `json:"as_of"`
	Lines  []FundamentalsLine `json:"lines,omitempty"`
}
