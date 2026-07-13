// markets.go — exchange metadata: allowed intervals (yfinance-go is
// daily-only), allowed MICs (XNAS/XNYS/...), and the default
// adjustment policy. Capacity: 1 struct (`MarketsConfig`).
package types

// MarketsConfig represents market configuration
type MarketsConfig struct {
	AllowedIntervals        []string `yaml:"allowed_intervals"`
	AllowedMics             []string `yaml:"allowed_mics"`
	DefaultAdjustmentPolicy string   `yaml:"default_adjustment_policy"`
}
