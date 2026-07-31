// markets.go — exchange metadata: allowed intervals (yfin is
// daily-only), allowed MICs (XNAS/XNYS/...), and the default
// adjustment policy.
package config

// MarketsConfig represents market configuration
type MarketsConfig struct {
	AllowedIntervals        []string `yaml:"allowed_intervals"`
	AllowedMics             []string `yaml:"allowed_mics"`
	DefaultAdjustmentPolicy string   `yaml:"default_adjustment_policy"`
}
