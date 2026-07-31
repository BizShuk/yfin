// fx.go — compatibility schema for the retired FX conversion surface.
// Runtime code does not construct an FX rate provider or conversion cache;
// pull --fx-target only validates that source and target currencies match.
package config

// FXConfig preserves previously accepted YAML fields for compatibility.
type FXConfig struct {
	Provider   string         `yaml:"provider"`
	Target     string         `yaml:"target"`
	CacheTTLMs int            `yaml:"cache_ttl_ms"`
	RateScale  int            `yaml:"rate_scale"`
	Rounding   string         `yaml:"rounding"`
	YahooWeb   YahooWebConfig `yaml:"yahoo_web"`
}

// YahooWebConfig preserves the inactive Yahoo Web policy fields.
type YahooWebConfig struct {
	QPS               float64 `yaml:"qps"`
	Burst             int     `yaml:"burst"`
	TimeoutMs         int     `yaml:"timeout_ms"`
	BackoffAttempts   int     `yaml:"backoff_attempts"`
	BackoffBaseMs     int     `yaml:"backoff_base_ms"`
	BackoffMaxDelayMs int     `yaml:"backoff_max_delay_ms"`
	CircuitResetMs    int     `yaml:"circuit_reset_ms"`
}
