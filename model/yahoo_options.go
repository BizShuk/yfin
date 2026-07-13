// yahoo_options.go — Yahoo option-chain DTOs.
// Originally lived in svc/yahoo/options.go; promoted to model/ so external
// consumers can depend on the shapes without importing the Decode/Fetch
// behavior of svc/yahoo.

package model

// OptionContract is a single option (call or put) inside an OptionExpiry.
type OptionContract struct {
	Strike            RawValue `json:"strike"`
	LastPrice         RawValue `json:"lastPrice"`
	Bid               RawValue `json:"bid"`
	Ask               RawValue `json:"ask"`
	Volume            RawInt   `json:"volume"`
	OpenInterest      RawInt   `json:"openInterest"`
	ImpliedVolatility RawValue `json:"impliedVolatility"`
}

// OptionExpiry is one expiration's call/put chain.
type OptionExpiry struct {
	ExpirationDate int64            `json:"expirationDate"`
	Calls          []OptionContract `json:"calls"`
	Puts           []OptionContract `json:"puts"`
}

// OptionsDTO is the full option-chain result (all available expirations).
type OptionsDTO struct {
	ExpirationDates []int64        `json:"expirationDates"`
	Strikes         []float64      `json:"strikes"`
	Options         []OptionExpiry `json:"options"`
}