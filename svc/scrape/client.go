// client.go — Scraping HTTP client: thin wrapper that enforces a
// `RobotsPolicy` then delegates the GET to an `httpx.Caller`. All
// retry / rate-limit / tracing / metrics / logging is owned by the
// underlying `httpx.Client` (see utils/httpx). Capacity: 1 interface,
// 2 constructors, 1 Fetch method.
package scrape

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/utils/httpx"
)

// Client interface for web scraping operations.
type Client interface {
	Fetch(ctx context.Context, url string) ([]byte, *model.FetchMeta, error)
}

// client implements the Client interface.
type client struct {
	config        *Config
	caller        httpx.Caller
	robotsManager *RobotsManager
}

// NewClientWithCaller wires a scraping client around any `httpx.Caller`
// (typically a `*httpx.Client` configured for `finance.yahoo.com`).
// The caller owns retry / rate-limit / tracing / metrics; this layer
// only enforces `RobotsPolicy` and adapts `*httpx.Meta` to `*model.FetchMeta`.
func NewClientWithCaller(caller httpx.Caller, config *Config) (Client, error) {
	if caller == nil {
		return nil, fmt.Errorf("scrape: caller must not be nil")
	}
	if config == nil {
		config = DefaultConfig()
	}
	return &client{
		config:        config,
		caller:        caller,
		robotsManager: NewRobotsManager(config.RobotsPolicy, time.Duration(config.CacheTTLMs)*time.Millisecond),
	}, nil
}

// NewClient is a deprecated convenience wrapper. It builds a default
// `httpx.Client` from `config.HTTP` and delegates to NewClientWithCaller.
// Pass nil for `pool` to use a fresh client; pass an existing client to
// share a connection pool.
//
// Historical note: pre-consolidation, this wrapper hardcoded 8 scrape-
// side defaults (IdleTimeout, MaxConnsPerHost, BackoffJitterMs,
// CircuitWindow, FailureThreshold, ResetTimeout, MaxBodyBytes, BaseURL)
// inline. Those defaults now live in `DefaultConfig()` so callers see
// the same behavior whether they pass `DefaultConfig()` explicitly or
// rely on the nil-config fallback.
func NewClient(config *Config, pool *httpx.Client) (Client, error) {
	if config == nil {
		config = DefaultConfig()
	}
	var caller httpx.Caller
	if pool != nil {
		caller = pool
	} else {
		caller = httpx.NewClient(config.HTTP)
	}
	return NewClientWithCaller(caller, config)
}

// Fetch applies the configured robots.txt policy, then delegates the complete
// URL to the underlying `httpx.Caller`. The caller is invoked exactly once per
// Fetch — retries and backoff happen inside httpx.
func (c *client) Fetch(ctx context.Context, urlStr string) ([]byte, *model.FetchMeta, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, &ScrapeError{
			Type:    "invalid_url",
			Message: fmt.Sprintf("failed to parse URL: %v", err),
			URL:     urlStr,
		}
	}

	if robotsErr := c.robotsManager.CheckRobots(ctx, u.Host, u.Path); robotsErr != nil {
		return nil, nil, robotsErr
	}

	ctx = yahooWebCircuitContext(ctx)
	body, meta, err := c.caller.Get(ctx, u.String(), u.Query())
	if err != nil {
		return nil, nil, err
	}
	return body, &model.FetchMeta{
		URL:          urlStr,
		Host:         u.Host,
		Status:       meta.Status,
		Bytes:        meta.Bytes,
		Duration:     meta.Duration,
		Attempt:      meta.Attempts,
		RobotsPolicy: c.config.RobotsPolicy,
	}, nil
}
