package norm

import "time"

type NormalizedHolder struct {
	Organization string         `json:"organization"`
	PercentHeld  *ScaledDecimal `json:"percent_held,omitempty"`
	Position     *int64         `json:"position,omitempty"`
	Value        *int64         `json:"value,omitempty"`
}

type NormalizedHolders struct {
	Security                Security           `json:"security"`
	InsidersPercentHeld     *ScaledDecimal     `json:"insiders_percent_held,omitempty"`
	InstitutionsPercentHeld *ScaledDecimal     `json:"institutions_percent_held,omitempty"`
	Institutional           []NormalizedHolder `json:"institutional"`
	MutualFund              []NormalizedHolder `json:"mutual_fund"`
	AsOf                    time.Time          `json:"as_of"`
	Meta                    Meta               `json:"meta"`
}
