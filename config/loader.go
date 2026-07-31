// loader.go — `Loader` reads the effective YAML via a local os.ReadFile +
// yaml.Unmarshal reader, interpolates ${VAR} / ${VAR:-default} references,
// maps the result into the `Config` struct tree, and validates business
// invariants (daily-only intervals, QPS / retry bounds, ...).
// `GetEffectiveConfig` re-loads for `--print-effective` output and
// redacts `secrets[].ref` plus any keys matching common secret
// patterns.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultEffectivePath is the repository-relative configuration used when a
// caller does not provide --config.
const DefaultEffectivePath = "config/effective.yaml"

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
	if effectivePath == "" {
		effectivePath = DefaultEffectivePath
	}

	return &Loader{
		effectivePath: effectivePath,
	}
}

// loadEffectiveMap reads the effective YAML file and interpolates environment
// variables before handing the raw map to a caller-specific processing path.
func (l *Loader) loadEffectiveMap() (map[string]interface{}, error) {
	configMap, err := loadYAMLFile(l.effectivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load effective config: %w", err)
	}

	l.interpolateEnvVars(configMap)
	return configMap, nil
}

// Load loads and validates configuration from the effective YAML file
func (l *Loader) Load() (*Config, error) {
	// Read and interpolate the effective YAML into a generic map; struct
	// mapping happens locally below.
	configMap, err := l.loadEffectiveMap()
	if err != nil {
		return nil, err
	}

	// Convert map to our Config struct
	config, err := l.mapToConfig(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}

	// Validate configuration
	if err := l.validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Post-load: assemble the canonical HTTP-layer configs in one place.
	// All three mappers (cmd/client.go:httpConfigToHttpx,
	// facade/scrape.go:NewScrapeClientFromConfig,
	// svc/scrape/client.go:NewClient) collapse to direct field reads.
	config.HTTP = config.assembleHTTPConfig()
	config.Scrape.HTTP = config.assembleScrapeHTTPConfig()

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
	applyConcurrencyDefaults(configMap, &config.Concurrency)
	if config.CircuitBreaker.MinimumRequests == 0 {
		config.CircuitBreaker.MinimumRequests = defaultCircuitMinimumRequests
	}

	return &config, nil
}

// validate validates the configuration
func (l *Loader) validate(config *Config) error {
	// Validate concurrency constraints
	if config.Concurrency.GlobalWorkers < 1 {
		return fmt.Errorf("concurrency.global_workers must be >= 1")
	}
	if config.Concurrency.PerHostWorkers < 1 {
		return fmt.Errorf("concurrency.per_host_workers must be >= 1")
	}
	if config.Concurrency.GlobalWorkers < config.Concurrency.PerHostWorkers {
		return fmt.Errorf("concurrency.global_workers (%d) must be >= per_host_workers (%d)",
			config.Concurrency.GlobalWorkers, config.Concurrency.PerHostWorkers)
	}

	// Sessions were removed (CLAUDE.md: session rotation dropped); the
	// per_session rate-limit knobs that depended on Sessions.N are also
	// gone, so no constraint check is needed here.

	// Validate markets.allowed_intervals (daily-only enforcement)
	if len(config.Markets.AllowedIntervals) != 1 || config.Markets.AllowedIntervals[0] != "1d" {
		return fmt.Errorf("markets.allowed_intervals must be exactly [\"1d\"] for yfin (daily-only scope)")
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
	if config.CircuitBreaker.MinimumRequests < 1 {
		return fmt.Errorf("circuit_breaker.minimum_requests must be >= 1")
	}

	// Validate observability configuration
	if err := config.Observability.Logs.validate(); err != nil {
		return err
	}
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
	configMap, err := l.loadEffectiveMap()
	if err != nil {
		return nil, err
	}

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
	for key, value := range configMap {
		// Skip the secrets array itself - it's handled separately
		if key == "secrets" {
			continue
		}

		if isSecretConfigKey(key) {
			configMap[key] = "[REDACTED]"
			continue
		}

		// Recursively process nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			l.redactSecretPatterns(nestedMap)
		}
	}
}

func isSecretConfigKey(key string) bool {
	key = strings.ToLower(key)

	switch key {
	case "key", "password", "passwd", "token", "api_key", "apikey",
		"secret", "private_key", "client_secret", "access_key", "secret_key":
		return true
	}

	parts := strings.FieldsFunc(key, func(r rune) bool {
		return r == '_' || r == '-' || r == '.'
	})
	for index, part := range parts {
		switch part {
		case "password", "passwd", "token", "secret", "apikey":
			return true
		case "key":
			if index > 0 {
				switch parts[index-1] {
				case "api", "private", "access":
					return true
				}
			}
		}
	}

	for _, suffix := range []string{
		"_password",
		"_passwd",
		"_token",
		"_api_key",
		"_apikey",
		"_secret",
		"_private_key",
		"_client_secret",
		"_access_key",
		"_secret_key",
	} {
		if strings.HasSuffix(key, suffix) {
			return true
		}
	}

	return false
}
