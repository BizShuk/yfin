// quote.go — `Quote` plain SDK struct + `FromQuote` ScaledDecimal → float64
// converter (nil-safe on `RegularMarketPrice`). Capacity: 1 struct + 1 converter.
package facade

import (
	"time"

	"github.com/bizshuk/yfin/svc/norm"
)

// Quote is one real-time-ish quote snapshot.
type Quote struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"` // 0 if RegularMarketPrice is nil (market closed)
	Currency  string    `json:"currency"`
	EventTime time.Time `json:"event_time"`
}

// FromQuote converts the internal norm quote to a public struct.
// Nil-safe; nil RegularMarketPrice -> 0.
func FromQuote(q *norm.NormalizedQuote) *Quote {
	if q == nil {
		return nil
	}
	var price float64
	if q.RegularMarketPrice != nil {
		price = norm.FromScaledDecimal(*q.RegularMarketPrice)
	}
	return &Quote{
		Symbol:    q.Security.Symbol,
		Price:     price,
		Currency:  q.CurrencyCode,
		EventTime: q.EventTime,
	}
}
