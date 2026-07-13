// yahoo_actions.go — Yahoo dividends + splits DTOs.
// Originally lived in svc/yahoo/actions.go; promoted to model/ so external
// consumers can depend on the shapes without importing the Extract/Fetch
// behavior of svc/yahoo.

package model

// Dividend represents a single dividend payment extracted from Yahoo chart events.
type Dividend struct {
	Date   int64   `json:"date"`
	Amount float64 `json:"amount"`
}

// Split represents a single stock split extracted from Yahoo chart events.
type Split struct {
	Date        int64  `json:"date"`
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	SplitRatio  string `json:"splitRatio"`
}

// ActionsDTO bundles the dividends + splits extracted for a single symbol.
type ActionsDTO struct {
	Dividends []Dividend
	Splits    []Split
}