// quote.go — `Quote` type alias + `FromQuote` ScaledDecimal → float64
// converter (nil-safe on `RegularMarketPrice`). The struct itself lives in
// model/quote.go; facade.Quote is a back-compat alias.
package facade

import (
	"github.com/bizshuk/yfin/model"
)

// Quote is one real-time-ish quote snapshot. Aliased from model.Quote.
type Quote = model.Quote

// FromQuote converts the internal model quote to a public struct.
// Nil-safe; nil RegularMarketPrice -> 0.
func FromQuote(q *model.NormalizedQuote) *Quote {
	if q == nil {
		return nil
	}
	var price float64
	if q.RegularMarketPrice != nil {
		price = model.FromScaledDecimal(*q.RegularMarketPrice)
	}
	return &Quote{
		Symbol:    q.Security.Symbol,
		Price:     price,
		Currency:  q.CurrencyCode,
		EventTime: q.EventTime,
	}
}
