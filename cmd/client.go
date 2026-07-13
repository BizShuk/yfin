// client.go — shared builders used by every sub-package:
// `CreateClient` (facade.NewClientWithConfig with cmd.Global flag overrides)
// + `CreateBusConfig` (NATS/Kafka bus config with safe fallback).
//
// facade is the single handler bridging cmd/ → svc/; CreateClient returns
// *facade.Client. bus publishing stays in cmd (transport, not data
// fetching), so CreateBusConfig remains a cmd-local helper.
//
// Capacity: 2 builder functions.
package cmd

import (
	"fmt"

	"github.com/bizshuk/yfin/config/types"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/bus"
	"github.com/bizshuk/yfin/utils/httpx"
)

// CreateClient builds the yfin CLI's fetch handle. Returns *facade.Client —
// the same SDK surface external consumers use. CLI flag overrides
// (`--qps`, `--retry-max`, `--timeout`) are applied on top of the
// ampy-config HTTP settings so `yfin --qps=10` wins over `cfg.qps=2`.
func CreateClient() (*facade.Client, error) {
	// Determine effective config path
	effectivePath := Global.ConfigFile
	if effectivePath == "" {
		effectivePath = "config/effective.yaml"
	}

	// Load configuration using ampy-config
	loader := types.NewLoader(effectivePath)
	cfg, err := loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Convert types.HTTPConfig → httpx.Config and apply CLI flag overrides.
	httpxConfig := httpConfigToHttpx(cfg.GetHTTPConfig())

	if Global.QPS > 0 {
		httpxConfig.QPS = Global.QPS
	}
	if Global.RetryMax > 0 {
		httpxConfig.MaxAttempts = Global.RetryMax
	}
	if Global.Timeout > 0 {
		httpxConfig.Timeout = Global.Timeout
	}

	return facade.NewClientWithConfig(httpxConfig), nil
}

// httpConfigToHttpx converts the flat types.HTTPConfig (loaded from YAML)
// into the *httpx.Config facade.NewClientWithConfig expects. Field-by-field
// mapping is mechanical; FailureThreshold gets converted from a 0–1 fraction
// to a 0–100 percentage (httpx's expected unit).
func httpConfigToHttpx(cfg *types.HTTPConfig) *httpx.Config {
	return &httpx.Config{
		BaseURL:          cfg.BaseURL,
		Timeout:          cfg.Timeout,
		IdleTimeout:      cfg.IdleTimeout,
		MaxConnsPerHost:  cfg.MaxConnsPerHost,
		UserAgent:        cfg.UserAgent,
		MaxAttempts:      cfg.MaxAttempts,
		BackoffBaseMs:    cfg.BackoffBaseMs,
		BackoffJitterMs:  cfg.BackoffJitterMs,
		MaxDelayMs:       cfg.MaxDelayMs,
		QPS:              cfg.QPS,
		Burst:            cfg.Burst,
		CircuitWindow:    cfg.CircuitWindow,
		FailureThreshold: int(cfg.FailureThreshold * 100), // 0–1 → 0–100
		ResetTimeout:     cfg.ResetTimeout,
		// MaxBodyBytes defaulting to 0 (unlimited) matches the previous
		// cmd/client.go behaviour before the facade indirection.
		MaxBodyBytes: 0,
	}
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
	loader := types.NewLoader(effectivePath)
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

// _ keeps httpx package referenced; helps future fields find the import.
