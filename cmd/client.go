// client.go — shared client builders used by every sub-package:
// `cliClient` (thin handle the yfin path uses to drive fetch helpers),
// `CreateClient` (yahoo.Client construction from global config), and
// `CreateBusConfig` (NATS/Kafka bus config with safe fallback).
// TWSE-specific builders (`buildTWSEClient` + `twseUserAgent`) live in
// `cmd/twse/client.go`; the local JSON sink (`writeJSONFile`) used by
// `pull` + `quote` lives in `cmd/market/client_json.go`.
// Capacity: 1 `cliClient` type + 2 exported builder functions.
package cmd

import (
	"fmt"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/svc/yahoo"
	"github.com/bizshuk/yfin/utils/bus"
	"github.com/bizshuk/yfin/utils/httpx"
)

// CliClient is a thin handle the yfin path uses to drive fetch helpers. As of
// Step 6 of plans/spicy-singing-swan.md, the yfin CLI no longer constructs a
// facade.Client — facade.Client is the SDK surface and returns plain structs,
// while the yfin bus-publishing code wants the internal *norm.* types for
// emit→proto. CliClient only holds the *yahoo.Client; the fetch.go helpers
// reach through it to call yahoo + norm directly.
type CliClient struct {
	Yahoo *yahoo.Client
}

// CreateClient creates the yfin CLI's fetch handle. Returns *CliClient — a
// small struct that wraps *yahoo.Client. The facade.Client constructor is no
// longer called from the yfin path (Step 6).
func CreateClient() (*CliClient, error) {
	// Determine effective config path
	effectivePath := Global.ConfigFile
	if effectivePath == "" {
		// Default to a standard effective config path
		effectivePath = "config/effective.yaml"
	}

	// Load configuration using ampy-config
	loader := config.NewLoader(effectivePath)
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Convert to HTTP config
	httpConfig := cfg.GetHTTPConfig()

	// Apply global flags if set (CLI flags override config)
	if Global.QPS > 0 {
		httpConfig.QPS = Global.QPS
	}
	if Global.RetryMax > 0 {
		httpConfig.MaxAttempts = Global.RetryMax
	}
	if Global.Timeout > 0 {
		httpConfig.Timeout = Global.Timeout
	}

	// Create httpx config from our config
	httpxConfig := &httpx.Config{
		BaseURL:          httpConfig.BaseURL,
		Timeout:          httpConfig.Timeout,
		IdleTimeout:      httpConfig.IdleTimeout,
		MaxConnsPerHost:  httpConfig.MaxConnsPerHost,
		UserAgent:        httpConfig.UserAgent,
		MaxAttempts:      httpConfig.MaxAttempts,
		BackoffBaseMs:    httpConfig.BackoffBaseMs,
		BackoffJitterMs:  httpConfig.BackoffJitterMs,
		MaxDelayMs:       httpConfig.MaxDelayMs,
		QPS:              httpConfig.QPS,
		Burst:            httpConfig.Burst,
		CircuitWindow:    httpConfig.CircuitWindow,
		FailureThreshold: int(httpConfig.FailureThreshold * 100), // Convert to percentage
		ResetTimeout:     httpConfig.ResetTimeout,
	}

	// Build the underlying *yahoo.Client directly. The CLI does not need a
	// facade.Client — its process* helpers route through fetchDailyBarsNorm /
	// fetchQuoteNorm / fetchFundamentalsNorm in cmd/fetch.go.
	httpClient := httpx.NewClient(httpxConfig)
	return &CliClient{Yahoo: yahoo.NewClient(httpClient, httpxConfig.BaseURL)}, nil
}

// CreateBusConfig creates bus configuration
func CreateBusConfig(env, topicPrefix string) *bus.Config {
	// Determine effective config path
	effectivePath := Global.ConfigFile
	if effectivePath == "" {
		// Default to a standard effective config path
		effectivePath = "config/effective.yaml"
	}

	// Load configuration using ampy-config
	loader := config.NewLoader(effectivePath)
	cfg, err := loader.Load()
	if err != nil {
		// Fallback to default config if loading fails
		return &bus.Config{
			Enabled:         true,
			Env:             env,
			TopicPrefix:     topicPrefix,
			MaxPayloadBytes: 1024 * 1024, // 1 MiB
			Publisher: bus.PublisherConfig{
				Backend: "nats",
				NATS: bus.NATSConfig{
					URL:          "nats://localhost:4222",
					SubjectStyle: "topic",
					AckWaitMs:    5000,
				},
			},
			Retry: bus.RetryConfig{
				Attempts:   5,
				BaseMs:     250,
				MaxDelayMs: 8000,
			},
			CircuitBreaker: bus.CircuitBreakerConfig{
				Window:           50,
				FailureThreshold: 0.30,
				ResetTimeoutMs:   30000,
				HalfOpenProbes:   3,
			},
		}
	}

	// Get bus config from loaded configuration
	busConfig := cfg.GetBusConfig()

	// Override with CLI parameters
	busConfig.Enabled = true
	busConfig.Env = env
	busConfig.TopicPrefix = topicPrefix

	// Convert to bus.Config
	return &bus.Config{
		Enabled:         busConfig.Enabled,
		Env:             busConfig.Env,
		TopicPrefix:     busConfig.TopicPrefix,
		MaxPayloadBytes: busConfig.MaxPayloadBytes,
		Publisher: bus.PublisherConfig{
			Backend: busConfig.Publisher.Backend,
			NATS: bus.NATSConfig{
				URL:          busConfig.Publisher.NATS.URL,
				SubjectStyle: busConfig.Publisher.NATS.SubjectStyle,
				AckWaitMs:    busConfig.Publisher.NATS.AckWaitMs,
			},
			Kafka: bus.KafkaConfig{
				Brokers:     busConfig.Publisher.Kafka.Brokers,
				Acks:        busConfig.Publisher.Kafka.Acks,
				Compression: busConfig.Publisher.Kafka.Compression,
			},
		},
		Retry: bus.RetryConfig{
			Attempts:   busConfig.Retry.Attempts,
			BaseMs:     busConfig.Retry.BaseMs,
			MaxDelayMs: busConfig.Retry.MaxDelayMs,
		},
		CircuitBreaker: bus.CircuitBreakerConfig{
			Window:           busConfig.CircuitBreaker.Window,
			FailureThreshold: busConfig.CircuitBreaker.FailureThreshold,
			ResetTimeoutMs:   busConfig.CircuitBreaker.ResetTimeoutMs,
			HalfOpenProbes:   busConfig.CircuitBreaker.HalfOpenProbes,
		},
	}
}
