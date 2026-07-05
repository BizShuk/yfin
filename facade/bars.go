// Package facade re-exports the yfinance-go normalized bar/quote/company-info
// types as plain Go structs so external consumers (e.g. data/stock) can avoid
// reflection against the internal/norm package. The facade lives inside the
// yfinance-go module so it can legally import the internal types; consumers
// stay clean of `internal/`.
package facade

import (
	"time"

	"github.com/bizshuk/yfinance-go/svc/norm"
)

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

// FromBarBatch converts the internal norm batch into a clean public struct.
// Returns nil if b is nil so callers can treat nil-checks uniformly.
func FromBarBatch(b *norm.NormalizedBarBatch) *BarBatch {
	if b == nil {
		return nil
	}
	out := &BarBatch{
		Symbol: b.Security.Symbol,
		MIC:    b.Security.MIC,
		Bars:   make([]Bar, 0, len(b.Bars)),
	}
	for _, nb := range b.Bars {
		out.Bars = append(out.Bars, Bar{
			Date:         nb.EventTime.UTC().Format("2006-01-02"),
			Open:         norm.FromScaledDecimal(nb.Open),
			High:         norm.FromScaledDecimal(nb.High),
			Low:          norm.FromScaledDecimal(nb.Low),
			Close:        norm.FromScaledDecimal(nb.Close),
			Volume:       nb.Volume,
			Adjusted:     nb.Adjusted,
			CurrencyCode: nb.CurrencyCode,
		})
	}
	return out
}

// helper to expose time formatting in tests if needed elsewhere.
func formatUTC(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

var _ = formatUTC // reserved for future use by other facade files