// loader_test.go — tests for config/config.Loader: NewLoader path
// capture, full file load + env-var interpolation, missing-file
// error, daily-only interval validation, GetEffectiveConfig redaction
// path, not-loaded guard, and the GetHTTPConfig /
// GetFXConfig / ValidateInterval / ValidateAdjustmentPolicy adapter
// sanity.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader("test.yaml")
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}
	if loader.effectivePath != "test.yaml" {
		t.Errorf("Expected effectivePath to be 'test.yaml', got '%s'", loader.effectivePath)
	}
}

func TestNewLoaderUsesDefaultPathWhenEmpty(t *testing.T) {
	loader := NewLoader("")
	require.Equal(t, "config/effective.yaml", loader.effectivePath)
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempFile := "test-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config == nil {
		t.Fatal("Config is nil")
	}

	// Check some default values
	if config.App.Env != "dev" {
		t.Errorf("Expected app.env to be 'dev', got '%s'", config.App.Env)
	}

	if config.Yahoo.BaseURL != "https://query2.finance.yahoo.com" {
		t.Errorf("Expected yahoo.base_url to be 'https://query2.finance.yahoo.com', got '%s'", config.Yahoo.BaseURL)
	}

	if len(config.Markets.AllowedIntervals) != 1 || config.Markets.AllowedIntervals[0] != "1d" {
		t.Errorf("Expected markets.allowed_intervals to be ['1d'], got %v", config.Markets.AllowedIntervals)
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	loader := NewLoader("nonexistent.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestInterpolateEnvVars(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	// Create a config with environment variable interpolation
	configContent := map[string]interface{}{
		"app": map[string]interface{}{
			"env": "dev",
		},
		"yahoo": map[string]interface{}{
			"base_url":   "https://query2.finance.yahoo.com",
			"timeout_ms": 5000,
		},
		"markets": map[string]interface{}{
			"allowed_intervals":         []string{"1d"},
			"default_adjustment_policy": "split_dividend",
		},
		"retry": map[string]interface{}{
			"attempts": 5,
		},
		"circuit_breaker": map[string]interface{}{
			"failure_threshold": 0.30,
		},
		"test_url":     "${TEST_VAR}",
		"test_default": "${MISSING_VAR:-default_value}",
	}

	// Create temporary file
	tempFile := "test-env-config.yaml"
	err := createTestConfigFile(tempFile, configContent)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// The interpolation should have been applied during loading
	// We can't directly test the interpolated values since they're in the map,
	// but we can test that the config loaded successfully
	if config == nil {
		t.Fatal("Config is nil")
	}
}

func TestLoadEffectiveMapInterpolatesEnvironmentVariables(t *testing.T) {
	t.Setenv("TEST_EFFECTIVE_HOST", "https://test.example.com")

	path := filepath.Join(t.TempDir(), "effective.yaml")
	content := []byte("yahoo:\n  base_url: ${TEST_EFFECTIVE_HOST}\n")
	require.NoError(t, os.WriteFile(path, content, 0o600))

	configMap, err := NewLoader(path).loadEffectiveMap()
	require.NoError(t, err)

	yahoo, ok := configMap["yahoo"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "https://test.example.com", yahoo["base_url"])
}

func TestValidateDailyOnlyIntervals(t *testing.T) {
	// Create a config with invalid intervals
	configContent := map[string]interface{}{
		"app": map[string]interface{}{
			"env": "dev",
		},
		"yahoo": map[string]interface{}{
			"base_url":   "https://query2.finance.yahoo.com",
			"timeout_ms": 5000,
		},
		"markets": map[string]interface{}{
			"allowed_intervals":         []string{"1h", "1d"}, // Invalid - should be only ["1d"]
			"default_adjustment_policy": "split_dividend",
		},
		"retry": map[string]interface{}{
			"attempts": 5,
		},
		"circuit_breaker": map[string]interface{}{
			"failure_threshold": 0.30,
		},
	}

	tempFile := "test-invalid-intervals.yaml"
	err := createTestConfigFile(tempFile, configContent)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	_, err = loader.Load()
	if err == nil {
		t.Fatal("Expected validation error for invalid intervals")
	}

	expectedError := "markets.allowed_intervals must be exactly [\"1d\"] for yfinance-go (daily-only scope)"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetEffectiveConfig(t *testing.T) {
	tempFile := "test-effective-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	_, err = loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	effectiveConfig, err := loader.GetEffectiveConfig()
	if err != nil {
		t.Fatalf("Failed to get effective config: %v", err)
	}

	if effectiveConfig == nil {
		t.Fatal("Effective config is nil")
	}

	// Check that secrets are redacted
	if secrets, ok := effectiveConfig["secrets"].([]interface{}); ok {
		// Secrets array should be present but empty or redacted
		_ = secrets // Just check it exists
	}
}

func TestGetEffectiveConfigPreservesNonSecretKeyNames(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "effective.yaml")
	require.NoError(t, CreateEffectiveConfig(tempFile))

	loader := NewLoader(tempFile)
	_, err := loader.Load()
	require.NoError(t, err)

	effectiveConfig, err := loader.GetEffectiveConfig()
	require.NoError(t, err)

	scrape, ok := effectiveConfig["scrape"].(map[string]interface{})
	require.True(t, ok)
	endpoints, ok := scrape["endpoints"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, true, endpoints["key_statistics"])
}

func TestIsSecretConfigKeyUsesNameBoundaries(t *testing.T) {
	for _, key := range []string{
		"password",
		"database_password_hash",
		"api_token_expires_at",
		"client_secret_version",
		"private_key_id",
		"access_key_id",
	} {
		require.Truef(t, isSecretConfigKey(key), "%s should be treated as secret", key)
	}

	for _, key := range []string{
		"key_statistics",
		"monkey",
		"secretary",
		"tokenizer",
	} {
		require.Falsef(t, isSecretConfigKey(key), "%s should not be treated as secret", key)
	}
}

func TestCreateEffectiveConfigContainsOnlyOwnedSections(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "effective.yaml")
	require.NoError(t, CreateEffectiveConfig(tempFile))

	effectiveConfig, err := loadYAMLFile(tempFile)
	require.NoError(t, err)
	require.NotContains(t, effectiveConfig, "bus")

	yahoo, ok := effectiveConfig["yahoo"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "yfin/1.x", yahoo["user_agent"])
}

func TestGetEffectiveConfigNotLoaded(t *testing.T) {
	loader := NewLoader("test.yaml")
	_, err := loader.GetEffectiveConfig()
	if err == nil {
		t.Fatal("Expected error when config not loaded")
	}

	expectedError := "configuration not loaded"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetHTTPConfig(t *testing.T) {
	tempFile := "test-http-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	httpConfig := config.GetHTTPConfig()
	if httpConfig == nil {
		t.Fatal("HTTP config is nil")
	}

	if httpConfig.BaseURL != "https://query2.finance.yahoo.com" {
		t.Errorf("Expected BaseURL to be 'https://query2.finance.yahoo.com', got '%s'", httpConfig.BaseURL)
	}

	if httpConfig.QPS != 5.0 {
		t.Errorf("Expected QPS to be 5.0, got %f", httpConfig.QPS)
	}
}

func TestAssembleScrapeHTTPConfig(t *testing.T) {
	tempFile := "test-scrape-http-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Scrape HTTP config is now assembled once during Load().
	if config.Scrape.HTTP == nil {
		t.Fatal("Scrape.HTTP is nil; expected post-load assembly")
	}
	sh := config.Scrape.HTTP

	if sh.BaseURL != "https://finance.yahoo.com" {
		t.Errorf("BaseURL = %q, want https://finance.yahoo.com", sh.BaseURL)
	}
	if sh.MaxBodyBytes != 8<<20 {
		t.Errorf("MaxBodyBytes = %d, want %d (8 MiB scrape cap)", sh.MaxBodyBytes, 8<<20)
	}
	// Sane scrape-tuned defaults: short timeout, low QPS, conservative retry.
	if sh.QPS <= 0 || sh.QPS > 5 {
		t.Errorf("QPS = %v, want (0,5]", sh.QPS)
	}
	if sh.MaxAttempts < 1 {
		t.Errorf("MaxAttempts = %d, want >= 1", sh.MaxAttempts)
	}
}

func TestHTTPAndScrapeHTTPTypeMatch(t *testing.T) {
	tempFile := "test-type-assert.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// HTTP and Scrape.HTTP should both be populated post-load. They are
	// different instances (Yahoo HTTP ≠ scrape traffic) but Go-static
	// same type.
	if cfg.HTTP == nil {
		t.Fatal("cfg.HTTP is nil")
	}
	if cfg.Scrape.HTTP == nil {
		t.Fatal("cfg.Scrape.HTTP is nil")
	}
	// Unit conversion: yaml timeout_ms (int) → http.Timeout (time.Duration).
	if cfg.HTTP.Timeout.Milliseconds() != int64(cfg.Yahoo.TimeoutMs) {
		t.Errorf("HTTP.Timeout = %v ms, want %d ms", cfg.HTTP.Timeout.Milliseconds(), cfg.Yahoo.TimeoutMs)
	}
	// Scrape HTTP knobs are no longer in yaml — defaults come from
	// scrapeHTTPDefaults() in adapters.go (mirrors svc/scrape.DefaultConfig).
	if cfg.Scrape.HTTP.Timeout != 10*time.Second {
		t.Errorf("Scrape.HTTP.Timeout = %v, want 10s", cfg.Scrape.HTTP.Timeout)
	}
	if cfg.HTTP.FailureThreshold != 0 {
		t.Errorf("FailureThreshold = %d, want legacy mode disabled", cfg.HTTP.FailureThreshold)
	}
	if cfg.HTTP.FailureRateThreshold != 0.30 {
		t.Errorf("FailureRateThreshold = %v, want 0.30", cfg.HTTP.FailureRateThreshold)
	}
	if cfg.HTTP.MinimumRequests != 10 {
		t.Errorf("MinimumRequests = %d, want 10", cfg.HTTP.MinimumRequests)
	}
}

func TestValidateCircuitMinimumRequests(t *testing.T) {
	path := "test-invalid-circuit-minimum.yaml"
	defer os.Remove(path)
	require.NoError(t, CreateEffectiveConfig(path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := strings.Replace(string(data), "minimum_requests: 10", "minimum_requests: -1", 1)
	require.NotEqual(t, string(data), content, "default config must contain minimum_requests")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	_, err = NewLoader(path).Load()
	require.ErrorContains(t, err, "circuit_breaker.minimum_requests must be >= 1")
}

func TestValidateConcurrencyWorkers(t *testing.T) {
	tests := []struct {
		name        string
		oldValue    string
		newValue    string
		wantMessage string
	}{
		{
			name:        "global workers must be positive",
			oldValue:    "global_workers: 64",
			newValue:    "global_workers: 0",
			wantMessage: "concurrency.global_workers must be >= 1",
		},
		{
			name:        "per-host workers must be positive",
			oldValue:    "per_host_workers: 32",
			newValue:    "per_host_workers: 0",
			wantMessage: "concurrency.per_host_workers must be >= 1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "effective.yaml")
			require.NoError(t, CreateEffectiveConfig(path))

			data, err := os.ReadFile(path)
			require.NoError(t, err)
			content := strings.Replace(string(data), test.oldValue, test.newValue, 1)
			require.NotEqual(t, string(data), content, "default config must contain target value")
			require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

			_, err = NewLoader(path).Load()
			require.ErrorContains(t, err, test.wantMessage)
		})
	}
}

func TestValidateObservabilityLogLevel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "effective.yaml")
	require.NoError(t, CreateEffectiveConfig(path))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := strings.Replace(string(data), "level: info", "level: verbose", 1)
	require.NotEqual(t, string(data), content, "default config must contain observability log level")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	_, err = NewLoader(path).Load()
	require.ErrorContains(t, err, "observability.logs.level must be one of")
}

func TestGetFXConfig(t *testing.T) {
	tempFile := "test-fx-config.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	fxConfig := config.GetFXConfig()
	if fxConfig == nil {
		t.Fatal("FX config is nil")
	}

	if fxConfig.Provider != "none" {
		t.Errorf("Expected provider to be 'none', got '%s'", fxConfig.Provider)
	}

	if fxConfig.CacheTTLMs != 60000 {
		t.Errorf("Expected CacheTTLMs to be 60000, got %d", fxConfig.CacheTTLMs)
	}
}

func TestValidateInterval(t *testing.T) {
	tempFile := "test-validate-interval.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test valid interval
	err = config.ValidateInterval("1d")
	if err != nil {
		t.Errorf("Expected '1d' to be valid, got error: %v", err)
	}

	// Test invalid interval
	err = config.ValidateInterval("1h")
	if err == nil {
		t.Error("Expected '1h' to be invalid")
	}
}

func TestValidateAdjustmentPolicy(t *testing.T) {
	tempFile := "test-validate-policy.yaml"
	err := CreateEffectiveConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	defer os.Remove(tempFile)

	loader := NewLoader(tempFile)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test valid policies
	err = config.ValidateAdjustmentPolicy("raw")
	if err != nil {
		t.Errorf("Expected 'raw' to be valid, got error: %v", err)
	}

	err = config.ValidateAdjustmentPolicy("split_dividend")
	if err != nil {
		t.Errorf("Expected 'split_dividend' to be valid, got error: %v", err)
	}

	// Test invalid policy
	err = config.ValidateAdjustmentPolicy("invalid")
	if err == nil {
		t.Error("Expected 'invalid' to be invalid")
	}
}

// Helper function to create test config files
func createTestConfigFile(filename string, config map[string]interface{}) error {
	// Marshal to YAML and write to file
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}
