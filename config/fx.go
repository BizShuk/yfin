// fx.go — FX provider settings. The FX provider converts non-USD
// prices to a target currency at emission time. `YahooWebConfig` is the
// embedded sub-config for the Yahoo-Web provider. Capacity: 2 structs
// (`FXConfig`, `YahooWebConfig`).
package config

// FXConfig represents FX configuration
type FXConfig struct {
	Provider   string         `yaml:"provider"`
	Target     string         `yaml:"target"`
	CacheTTLMs int            `yaml:"cache_ttl_ms"`
	RateScale  int            `yaml:"rate_scale"`
	Rounding   string         `yaml:"rounding"`
	YahooWeb   YahooWebConfig `yaml:"yahoo_web"`
}

// YahooWebConfig represents Yahoo Web FX provider configuration
type YahooWebConfig struct {
	QPS               float64 `yaml:"qps"`
	Burst             int     `yaml:"burst"`
	TimeoutMs         int     `yaml:"timeout_ms"`
	BackoffAttempts   int     `yaml:"backoff_attempts"`
	BackoffBaseMs     int     `yaml:"backoff_base_ms"`
	BackoffMaxDelayMs int     `yaml:"backoff_max_delay_ms"`
	CircuitResetMs    int     `yaml:"circuit_reset_ms"`
}
