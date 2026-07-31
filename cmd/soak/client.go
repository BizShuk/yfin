// client.go — builds the soak facade client from one loaded config snapshot.
package main

import (
	"fmt"
	"net/http"

	"github.com/bizshuk/yfin/config"
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/utils/httpx"
)

func newSoakClient(loaded *config.Config, global GlobalConfig, soak *SoakConfig) (*facade.Client, error) {
	if loaded == nil {
		return nil, fmt.Errorf("loaded config is required")
	}
	if soak == nil {
		return nil, fmt.Errorf("soak config is required")
	}

	httpConfig, err := soakHTTPConfig(loaded.HTTP, global, soak.QPS, soak.FailureRate)
	if err != nil {
		return nil, err
	}
	scrapeHTTP, err := cloneHTTPConfig(loaded.Scrape.HTTP)
	if err != nil {
		return nil, fmt.Errorf("scrape HTTP config: %w", err)
	}
	scrapeHTTP.QPS = soak.QPS
	if global.RetryMax > 0 {
		scrapeHTTP.MaxAttempts = global.RetryMax
	}
	if global.Timeout > 0 {
		scrapeHTTP.Timeout = global.Timeout
	}

	client := facade.NewClientWithConfig(httpConfig)
	scrapeConfig := loaded.Scrape
	scrapeConfig.HTTP = scrapeHTTP
	if err := client.ConfigureScrape(&scrapeConfig); err != nil {
		return nil, fmt.Errorf("configure scrape client: %w", err)
	}
	return client, nil
}

func soakHTTPConfig(
	loaded *httpx.Config,
	global GlobalConfig,
	qps float64,
	failureRate float64,
) (*httpx.Config, error) {
	config, err := cloneHTTPConfig(loaded)
	if err != nil {
		return nil, err
	}
	if qps <= 0 {
		return nil, fmt.Errorf("soak QPS must be greater than zero")
	}
	if failureRate < 0 || failureRate > 1 {
		return nil, fmt.Errorf("failure rate must be between 0 and 1")
	}
	if global.RetryMax < 0 {
		return nil, fmt.Errorf("retry max must not be negative")
	}
	if global.Timeout < 0 {
		return nil, fmt.Errorf("timeout must not be negative")
	}

	config.QPS = qps
	config.Burst = max(1, int(qps)+1)
	if global.RetryMax > 0 {
		config.MaxAttempts = global.RetryMax
	}
	if global.Timeout > 0 {
		config.Timeout = global.Timeout
	}
	if failureRate > 0 {
		previous := config.WrapTransport
		config.WrapTransport = func(next http.RoundTripper) http.RoundTripper {
			if previous != nil {
				next = previous(next)
			}
			return newFailureTransport(next, failureRate)
		}
	}
	return config, nil
}

func cloneHTTPConfig(loaded *httpx.Config) (*httpx.Config, error) {
	if loaded == nil {
		return nil, fmt.Errorf("loaded HTTP config is required")
	}
	clone := *loaded
	return &clone, nil
}
