// client.go — shared client/writer builders used by every subcommand:
// `cliClient` (thin handle the yfin path uses to drive fetch helpers),
// `createClient` (yahoo.Client construction from global config),
// `createBusConfig` (NATS/Kafka bus config with safe fallback),
// `buildTWSEClient` + `twseUserAgent` (TWSE HTTP transport factory and
// the browser UA TWSE rejects the default Go UA for),
// and `writeJSONFile` (the local-export sink used by `pull` and `quote`).
// Capacity: 1 `cliClient` type + 1 `twseUserAgent` const + 6 builder functions.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/svc/twse"
	"github.com/bizshuk/yfin/svc/yahoo"
	"github.com/bizshuk/yfin/utils/bus"
	"github.com/bizshuk/yfin/utils/httpx"
)

// cliClient is a thin handle the yfin path uses to drive fetch helpers. As of
// Step 6 of plans/spicy-singing-swan.md, the yfin CLI no longer constructs a
// facade.Client — facade.Client is the SDK surface and returns plain structs,
// while the yfin bus-publishing code wants the internal *norm.* types for
// emit→proto. cliClient only holds the *yahoo.Client; the fetch.go helpers
// reach through it to call yahoo + norm directly.
type cliClient struct {
	Yahoo *yahoo.Client
}

// createClient creates the yfin CLI's fetch handle. Returns *cliClient — a
// small struct that wraps *yahoo.Client. The facade.Client constructor is no
// longer called from the yfin path (Step 6).
func createClient() (*cliClient, error) {
	// Determine effective config path
	effectivePath := globalConfig.ConfigFile
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
	if globalConfig.QPS > 0 {
		httpConfig.QPS = globalConfig.QPS
	}
	if globalConfig.RetryMax > 0 {
		httpConfig.MaxAttempts = globalConfig.RetryMax
	}
	if globalConfig.Timeout > 0 {
		httpConfig.Timeout = globalConfig.Timeout
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
	return &cliClient{Yahoo: yahoo.NewClient(httpClient, httpxConfig.BaseURL)}, nil
}

// createBusConfig creates bus configuration
func createBusConfig(env, topicPrefix string) *bus.Config {
	// Determine effective config path
	effectivePath := globalConfig.ConfigFile
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

// twseUserAgent is the browser-like UA TWSE rejects the default Go
// `User-Agent` for. It's set on the httpx.Config.UserAgent so every
// fetch from the unified TWSE client carries it without per-method
// header hacks.
const twseUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

// buildTWSEClient returns the process-wide `*twse.Client` used by the
// `yfin twse` subcommand. It wires an `httpx.Client` tuned for TWSE's
// public REST API (low QPS, conservative retry) and registers the
// TWSE-specific request middleware. The client captures
// `twse.BaseURL` at construction time; tests pass `*twse.Client`
// built via `svc/twse.NewClientWithURL` to point at an httptest server.
//
// Wiring pattern for future services (illustrative — actual Yahoo +
// scrape migration is out of Task 4's scope):
//
//	yc := yahoo.NewCrumbManager(httpx.NewClient(...), cookieURL, apiURL)
//	hc.Use(httpx.YahooMiddleware(yc))   // crumb injected per request
//	sc := httpx.NewClient(&httpx.Config{...ScrapeProfile...})
//	sc.Use(httpx.ScrapeMiddleware(ua))  // browser headers
func buildTWSEClient() *twse.Client {
	hc := httpx.NewClient(&httpx.Config{
		// BaseURL is intentionally empty: twse.Client owns the full
		// TWSE host+path prefix and passes pre-built absolute URLs to
		// caller.Get. Setting it here would double-concatenate.
		BaseURL:          "",
		Timeout:          30 * time.Second,
		IdleTimeout:      90 * time.Second,
		MaxConnsPerHost:  10,
		MaxAttempts:      3,
		BackoffBaseMs:    500,
		BackoffJitterMs:  250,
		MaxDelayMs:       8000,
		QPS:              2.0,
		Burst:            4,
		CircuitWindow:    60 * time.Second,
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		UserAgent:        twseUserAgent,
		MaxBodyBytes:     0, // 0 = unlimited; TWSE responses are JSON envelopes, not HTML
	})
	// Pin User-Agent even if Config.UserAgent is overridden by a
	// future config-loader change. TWSEMiddleware with an empty UA
	// would be a no-op; we pass the canonical string explicitly.
	hc.Use(httpx.TWSEMiddleware(twseUserAgent))
	return twse.NewClient(hc)
}

// writeJSONFile writes data to a JSON file
func writeJSONFile(filepath string, data interface{}) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
