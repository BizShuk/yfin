// fx_interface.go — `FXConverter` contract for currency translation (`ConvertValue(ctx, value, fromCurrency, toCurrency, at)` returns `ScaledDecimal` + `FXMeta`) and the accompanying `FXMeta` provenance struct (provider, base, symbols, as-of, rate scale, cache hit, attempts, backoff profile, stale flag).

package norm

import (
	"context"
	"time"
)

// FXConverter interface for currency conversion
type FXConverter interface {
	ConvertValue(ctx context.Context, value ScaledDecimal, fromCurrency, toCurrency string, at time.Time) (ScaledDecimal, *FXMeta, error)
}

// FXMeta contains metadata about FX conversion
type FXMeta struct {
	Provider       string    `json:"provider"`        // "none" | "yahoo-web"
	Base           string    `json:"base"`            // e.g., "EUR"
	Symbols        []string  `json:"symbols"`         // e.g., ["USD"]
	AsOf           time.Time `json:"as_of"`           // timestamp of FX rates
	RateScale      int       `json:"rate_scale"`      // scale of the rate decimals (e.g., 8)
	CacheHit       bool      `json:"cache_hit"`       // whether this was a cache hit
	Attempts       int       `json:"attempts"`        // number of attempts made
	BackoffProfile string    `json:"backoff_profile"` // backoff profile used
	Stale          bool      `json:"stale"`           // whether rates are stale
}
