// loader.go — `Loader` reads the effective YAML via a local os.ReadFile +
// yaml.Unmarshal reader, interpolates ${VAR} / ${VAR:-default} references,
// maps the result into the `Config` struct tree, and validates business
// invariants (daily-only intervals, QPS / retry bounds, ...).
// `GetEffectiveConfig` re-loads for `--print-effective` output and
// redacts `secrets[].ref` plus any keys matching common secret
// patterns. Capacity: 1 `Loader` + `NewLoader` + `Load` + `GetEffectiveConfig` +
// 5 private helpers (`interpolateEnvVars`, `interpolateString`,
// `mapToConfig`, `validate`, `redactSecrets`, `redactSecretPatterns`).
package types

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadYAMLFile reads the YAML file at path into a generic map, replacing the
// former ampy-config Loader which only provided os.ReadFile+yaml.Unmarshal
// under the hood (no secret injection / multi-file merge / hot-reload).
func loadYAMLFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]interface{}{}
	}
	return out, nil
}

// Loader handles configuration loading from the effective YAML file
type Loader struct {
	effectivePath string
	config        *Config
}

// NewLoader creates a new configuration loader
func NewLoader(effectivePath string) *Loader {
	return &Loader{
		effectivePath: effectivePath,
	}
}

// Load loads and validates configuration from the effective YAML file
func (l *Loader) Load() (*Config, error) {
	// Read the effective YAML into a generic map (env-var interpolation +
	// struct mapping happen locally below).
	configMap, err := loadYAMLFile(l.effectivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load effective config: %w", err)
	}

	// Interpolate environment variables
	l.interpolateEnvVars(configMap)

	// Convert map to our Config struct
	config, err := l.mapToConfig(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}

	// Validate configuration
	if err := l.validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	l.config = config
	return config, nil
}

// interpolateEnvVars interpolates environment variables in the configuration map
func (l *Loader) interpolateEnvVars(configMap map[string]interface{}) {
	for key, value := range configMap {
		if str, ok := value.(string); ok {
			// Handle ${VAR} and ${VAR:-default} syntax
			configMap[key] = l.interpolateString(str)
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively process nested maps
			l.interpolateEnvVars(nestedMap)
		} else if slice, ok := value.([]interface{}); ok {
			// Process slices
			for i, item := range slice {
				if str, ok := item.(string); ok {
					slice[i] = l.interpolateString(str)
				} else if nestedMap, ok := item.(map[string]interface{}); ok {
					l.interpolateEnvVars(nestedMap)
				}
			}
		}
	}
}

// interpolateString interpolates environment variables in a string
func (l *Loader) interpolateString(str string) string {
	// Handle ${VAR} and ${VAR:-default} syntax
	result := str
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}

		end += start
		varExpr := result[start+2 : end]

		var value string
		if strings.Contains(varExpr, ":-") {
			// Handle ${VAR:-default} syntax
			parts := strings.SplitN(varExpr, ":-", 2)
			envVar := parts[0]
			defaultValue := parts[1]
			value = os.Getenv(envVar)
			if value == "" {
				value = defaultValue
			}
		} else {
			// Handle ${VAR} syntax
			value = os.Getenv(varExpr)
		}

		result = result[:start] + value + result[end+1:]
	}

	return result
}

// mapToConfig converts a map to our Config struct
func (l *Loader) mapToConfig(configMap map[string]interface{}) (*Config, error) {
	// Marshal to YAML and unmarshal to struct
	data, err := yaml.Marshal(configMap)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validate validates the configuration
func (l *Loader) validate(config *Config) error {
	// Validate app.env
	if config.App.Env != "dev" && config.App.Env != "staging" && config.App.Env != "prod" {
		// Allow custom env but warn
		// In a real implementation, you might want to log a warning
		_ = config.App.Env // Suppress unused variable warning
	}

	// Validate concurrency constraints
	if config.Concurrency.GlobalWorkers < config.Concurrency.PerHostWorkers {
		return fmt.Errorf("concurrency.global_workers (%d) must be >= per_host_workers (%d)",
			config.Concurrency.GlobalWorkers, config.Concurrency.PerHostWorkers)
	}

	if config.Concurrency.PerHostWorkers < config.Sessions.N {
		return fmt.Errorf("concurrency.per_host_workers (%d) must be >= sessions.n (%d)",
			config.Concurrency.PerHostWorkers, config.Sessions.N)
	}

	// Validate rate limit constraints
	if config.RateLimit.PerSessionQPS*float64(config.Sessions.N) > config.RateLimit.PerHostQPS {
		// This is a warning, not an error
		// In a real implementation, you might want to log a warning
		_ = config.RateLimit.PerSessionQPS // Suppress unused variable warning
	}

	// Validate markets.allowed_intervals (daily-only enforcement)
	if len(config.Markets.AllowedIntervals) != 1 || config.Markets.AllowedIntervals[0] != "1d" {
		return fmt.Errorf("markets.allowed_intervals must be exactly [\"1d\"] for yfinance-go (daily-only scope)")
	}

	// Validate markets.default_adjustment_policy
	if config.Markets.DefaultAdjustmentPolicy != "raw" && config.Markets.DefaultAdjustmentPolicy != "split_dividend" {
		return fmt.Errorf("markets.default_adjustment_policy must be 'raw' or 'split_dividend'")
	}

	// Validate retry.attempts
	if config.Retry.Attempts < 1 {
		return fmt.Errorf("retry.attempts must be >= 1")
	}

	// Validate circuit breaker thresholds
	if config.CircuitBreaker.FailureThreshold <= 0 || config.CircuitBreaker.FailureThreshold > 1 {
		return fmt.Errorf("circuit_breaker.failure_threshold must be between 0 and 1")
	}

	// Validate observability configuration
	if config.Observability.Metrics.Prometheus.Enabled && config.Observability.Metrics.Prometheus.Addr == "" {
		return fmt.Errorf("observability.metrics.prometheus.addr is required when prometheus is enabled")
	}

	if config.Observability.Tracing.OTLP.Enabled && config.Observability.Tracing.OTLP.Endpoint == "" {
		return fmt.Errorf("observability.tracing.otlp.endpoint is required when OTLP tracing is enabled")
	}

	return nil
}

// GetEffectiveConfig returns the effective configuration as a map for printing
func (l *Loader) GetEffectiveConfig() (map[string]interface{}, error) {
	if l.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	// Re-read the raw effective config map for printing
	configMap, err := loadYAMLFile(l.effectivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load effective config: %w", err)
	}

	// Interpolate environment variables
	l.interpolateEnvVars(configMap)

	// Redact secrets
	l.redactSecrets(configMap)

	return configMap, nil
}

// redactSecrets redacts secret values in the configuration map
func (l *Loader) redactSecrets(configMap map[string]interface{}) {
	// Redact secrets section
	if secrets, ok := configMap["secrets"].([]interface{}); ok {
		for i := range secrets {
			if secretMap, ok := secrets[i].(map[string]interface{}); ok {
				if _, ok := secretMap["ref"].(string); ok {
					secretMap["ref"] = "[REDACTED]"
				}
			}
		}
	}

	// Redact known secret patterns
	l.redactSecretPatterns(configMap)
}

// redactSecretPatterns redacts values that match secret patterns
func (l *Loader) redactSecretPatterns(configMap map[string]interface{}) {
	secretPatterns := []string{"password", "token", "api_key", "secret", "key"}

	for key, value := range configMap {
		keyLower := strings.ToLower(key)

		// Skip the secrets array itself - it's handled separately
		if key == "secrets" {
			continue
		}

		// Check if key matches secret patterns
		for _, pattern := range secretPatterns {
			if strings.Contains(keyLower, pattern) {
				configMap[key] = "[REDACTED]"
				continue
			}
		}

		// Recursively process nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			l.redactSecretPatterns(nestedMap)
		}
	}
}
