// quote.go — `Quote` plain data struct for the SDK surface.
// Originally lived in facade/quote.go; promoted to model/ for cross-layer reuse.

package model

import "time"

// Quote is one real-time-ish quote snapshot.
type Quote struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"` // 0 if RegularMarketPrice is nil (market closed)
	Currency  string    `json:"currency"`
	EventTime time.Time `json:"event_time"`
}
