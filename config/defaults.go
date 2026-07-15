// defaults.go — `CreateEffectiveConfig` writes a self-consistent
// default effective config (sane dev defaults: localhost NATS, dev
// tracing, daily-only markets, 5 QPS per host, ...) so tests and
// first-run users have a working config out of the box. The default
// is intentionally a single literal map (instead of building each
// struct field-by-field) so the YAML output exactly mirrors what the
// loader expects to find. Capacity: 1 public function
// (`CreateEffectiveConfig`).
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// CreateEffectiveConfig creates a default effective config file for testing
func CreateEffectiveConfig(path string) error {
	// Create a default effective config
	defaultConfig := map[string]interface{}{
		"app": map[string]interface{}{
			"env":    "dev",
			"run_id": "",
		},
		"yahoo": map[string]interface{}{
			"base_url":           "https://query2.finance.yahoo.com",
			"timeout_ms":         6000,
			"idle_timeout_ms":    30000,
			"max_conns_per_host": 64,
			"user_agent":         "AmpyFin-yfinance-go/1.x",
		},
		"concurrency": map[string]interface{}{
			"global_workers":   64,
			"per_host_workers": 32,
		},
		"rate_limit": map[string]interface{}{
			"per_host_qps":   5.0,
			"per_host_burst": 5,
		},
		"retry": map[string]interface{}{
			"attempts":     5,
			"base_ms":      250,
			"max_delay_ms": 8000,
		},
		"circuit_breaker": map[string]interface{}{
			"window":            50,
			"failure_threshold": 0.30,
			"reset_timeout_ms":  30000,
		},
		"markets": map[string]interface{}{
			"allowed_intervals":         []string{"1d"},
			"allowed_mics":              []string{"XNAS", "XNYS", "XNMS", "NYQ", "KSC", "XETR", "XTKS"},
			"default_adjustment_policy": "split_dividend",
		},
		"fx": map[string]interface{}{
			"provider":     "none",
			"target":       "",
			"cache_ttl_ms": 60000,
			"rate_scale":   8,
			"rounding":     "half_up",
			"yahoo_web": map[string]interface{}{
				"qps":                  0.5,
				"burst":                1,
				"timeout_ms":           5000,
				"backoff_attempts":     4,
				"backoff_base_ms":      250,
				"backoff_max_delay_ms": 4000,
				"circuit_reset_ms":     30000,
			},
		},
		"bus": map[string]interface{}{
			"enabled":           false,
			"env":               "dev",
			"topic_prefix":      "ampy",
			"max_payload_bytes": 1048576,
			"publisher": map[string]interface{}{
				"backend": "nats",
				"nats": map[string]interface{}{
					"url":           "nats://localhost:4222",
					"subject_style": "topic",
					"ack_wait_ms":   5000,
				},
				"kafka": map[string]interface{}{
					"brokers":     []string{},
					"acks":        "all",
					"compression": "snappy",
				},
			},
			"retry": map[string]interface{}{
				"attempts":     5,
				"base_ms":      250,
				"max_delay_ms": 8000,
			},
			"circuit_breaker": map[string]interface{}{
				"window":            50,
				"failure_threshold": 0.30,
				"reset_timeout_ms":  30000,
				"half_open_probes":  3,
			},
		},
		"scrape": map[string]interface{}{
			"enabled":       true,
			"robots_policy": "enforce",
			"cache_ttl_ms":  60000,
			"endpoints": map[string]interface{}{
				"key_statistics": true,
				"financials":     true,
				"analysis":       true,
				"profile":        true,
				"news":           true,
			},
		},
		"observability": map[string]interface{}{
			"logs": map[string]interface{}{
				"level": "info",
			},
			"metrics": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"enabled": true,
					"addr":    ":9090",
				},
			},
			"tracing": map[string]interface{}{
				"otlp": map[string]interface{}{
					"enabled":      true,
					"endpoint":     "http://localhost:4317",
					"sample_ratio": 0.05,
				},
			},
		},
		"secrets": []interface{}{},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0600)
}
