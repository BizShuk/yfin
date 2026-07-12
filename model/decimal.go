// decimal.go — `ScaledDecimal` precision type + stateless conversion helpers.
// Originally lived in svc/norm/decimal.go (helpers) and svc/norm/types.go
// (struct). Promoted to model/ so any layer can depend on the precision type
// without pulling in the rest of the normalize pipeline.

package model

import (
	"fmt"
	"math"
	"math/big"
)

// ScaledDecimal represents a decimal value with explicit scale
type ScaledDecimal struct {
	Scaled int64 `json:"scaled"`
	Scale  int   `json:"scale"`
}

// Float64 returns the float64 value of ScaledDecimal.
func (s ScaledDecimal) Float64() float64 {
	if s.Scale == 0 {
		return float64(s.Scaled)
	}
	multiplier := math.Pow(10, float64(s.Scale))
	return float64(s.Scaled) / multiplier
}

// String returns a human-readable representation of ScaledDecimal.
func (s ScaledDecimal) String() string {
	if s.Scale == 0 {
		return fmt.Sprintf("%d", s.Scaled)
	}
	return fmt.Sprintf("%.6f", s.Float64())
}

// IsZero reports whether the ScaledDecimal represents zero.
func (s ScaledDecimal) IsZero() bool {
	return s.Scaled == 0
}

// GetScaleForCurrency returns the appropriate scale for a given currency
// USD/EUR/GBP use scale 2 (cents), JPY uses scale 2
func GetScaleForCurrency(currency string) int {
	switch currency {
	case "JPY":
		return 2
	case "USD", "EUR", "GBP", "CAD", "AUD", "CHF", "NZD":
		return 2 // Use scale 2 for cents
	default:
		// Default to scale 2 for most currencies
		return 2
	}
}

// ToScaledDecimal converts a float64 price to a scaled decimal
func ToScaledDecimal(price float64, scale int) (ScaledDecimal, error) {
	if math.IsNaN(price) {
		return ScaledDecimal{}, fmt.Errorf("NaN price")
	}
	if math.IsInf(price, 0) {
		return ScaledDecimal{}, fmt.Errorf("infinite price")
	}
	multiplier := math.Pow(10, float64(scale))
	scaled := int64(math.Round(price * multiplier))
	return ScaledDecimal{
		Scaled: scaled,
		Scale:  scale,
	}, nil
}

// ToScaledDecimalWithCurrency converts a float64 price to a scaled decimal using currency-appropriate scale
func ToScaledDecimalWithCurrency(price float64, currency string) (ScaledDecimal, error) {
	scale := GetScaleForCurrency(currency)
	return ToScaledDecimal(price, scale)
}

// FromScaledDecimal converts a scaled decimal back to float64. Alias of
// ScaledDecimal.Float64(); kept as a free function for callers that have a
// value (not pointer) and prefer function-call syntax.
func FromScaledDecimal(sd ScaledDecimal) float64 {
	return sd.Float64()
}

// ToScaledDecimalPtr converts a float64 pointer to a ScaledDecimal pointer.
// nil in -> nil out. Originally in svc/norm/market_data.go.
func ToScaledDecimalPtr(value *float64, currency string) *ScaledDecimal {
	if value == nil {
		return nil
	}
	scaled, _ := ToScaledDecimalWithCurrency(*value, currency)
	return &scaled
}

// ValidateScaledDecimal validates a scaled decimal
func ValidateScaledDecimal(sd ScaledDecimal) error {
	if sd.Scale < 0 {
		return fmt.Errorf("negative scale: %d", sd.Scale)
	}
	if sd.Scale > 8 {
		return fmt.Errorf("scale too large: %d", sd.Scale)
	}
	return nil
}

// GetPriceScaleForCurrency returns the appropriate scale for price data in a given currency.
// This is an alias for GetScaleForCurrency for compatibility.
func GetPriceScaleForCurrency(currency string) int {
	return GetScaleForCurrency(currency)
}

// MultiplyAndRound multiplies two scaled decimals and rounds to the target scale
func MultiplyAndRound(a ScaledDecimal, b ScaledDecimal, targetScale int) (ScaledDecimal, error) {
	if err := ValidateScaledDecimal(a); err != nil {
		return ScaledDecimal{}, fmt.Errorf("invalid first operand: %w", err)
	}
	if err := ValidateScaledDecimal(b); err != nil {
		return ScaledDecimal{}, fmt.Errorf("invalid second operand: %w", err)
	}
	if targetScale < 0 || targetScale > 8 {
		return ScaledDecimal{}, fmt.Errorf("invalid target scale: %d", targetScale)
	}

	totalScale := a.Scale + b.Scale
	scaleDiff := totalScale - targetScale

	if scaleDiff < 0 {
		multiplier := int64(math.Pow(10, float64(-scaleDiff)))
		result := a.Scaled * b.Scaled * multiplier
		return ScaledDecimal{
			Scaled: result,
			Scale:  targetScale,
		}, nil
	} else if scaleDiff > 0 {
		divisor := int64(math.Pow(10, float64(scaleDiff)))
		result := a.Scaled * b.Scaled / divisor
		remainder := (a.Scaled * b.Scaled) % divisor
		if remainder >= divisor/2 {
			result++
		}
		return ScaledDecimal{
			Scaled: result,
			Scale:  targetScale,
		}, nil
	} else {
		return ScaledDecimal{
			Scaled: a.Scaled * b.Scaled,
			Scale:  targetScale,
		}, nil
	}
}

// RoundHalfUp rounds a big.Int value from one scale to another using half-up rounding
func RoundHalfUp(value *big.Int, fromScale, toScale int) *big.Int {
	if fromScale == toScale {
		return new(big.Int).Set(value)
	}

	if fromScale < toScale {
		multiplier := big.NewInt(1)
		for i := 0; i < toScale-fromScale; i++ {
			multiplier.Mul(multiplier, big.NewInt(10))
		}
		return new(big.Int).Mul(value, multiplier)
	}

	divisor := big.NewInt(1)
	for i := 0; i < fromScale-toScale; i++ {
		divisor.Mul(divisor, big.NewInt(10))
	}

	quotient := new(big.Int).Div(value, divisor)
	remainder := new(big.Int).Mod(value, divisor)

	half := new(big.Int).Div(divisor, big.NewInt(2))
	if remainder.Cmp(half) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}

	return quotient
}
