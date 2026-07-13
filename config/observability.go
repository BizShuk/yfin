// observability.go — logs (level) + metrics (Prometheus) + tracing
// (OTLP). All sub-types are nested here because each domain has only
// one canonical shape; splitting them out buys no readability and
// forces every reader to chase 5 files. Capacity: 6 structs
// (`ObservabilityConfig`, `LogsConfig`, `MetricsConfig`,
// `PrometheusConfig`, `TracingConfig`, `OTLPConfig`).
package config

// ObservabilityConfig represents observability configuration
type ObservabilityConfig struct {
	Logs    LogsConfig    `yaml:"logs"`
	Metrics MetricsConfig `yaml:"metrics"`
	Tracing TracingConfig `yaml:"tracing"`
}

// LogsConfig represents logging configuration
type LogsConfig struct {
	Level string `yaml:"level"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	OTLP OTLPConfig `yaml:"otlp"`
}

// OTLPConfig represents OTLP configuration
type OTLPConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Endpoint    string  `yaml:"endpoint"`
	SampleRatio float64 `yaml:"sample_ratio"`
}
