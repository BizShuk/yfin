// twse.go — the facade-owned TWSE client handle. `TwseClient` wraps
// `*svc/twse.Client` so that no caller above facade ever has to name an
// svc/twse type: `cmd/twse` holds a `*facade.TwseClient` and calls
// `(*TwseClient).Dispatch`, keeping the `cmd → facade → svc` edge intact.
//
// Capacity: 1 `TwseClient` struct + 2 constructors (`NewTwseClient`,
// `NewTwseClientWithHTTP`) + 3 accessors/methods (`Dispatch`, `BaseURL`,
// `Caller`) + `TwseIsNoData` error classification. Endpoint registry and
// fetcher dispatch live in `facade/twse_dispatch.go`.
package facade

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/bizshuk/yfin/svc/twse"
	"github.com/bizshuk/yfin/utils/httpx"
)

// twseUserAgent is the browser-like UA TWSE requires; it rejects the
// default Go `User-Agent`.
const twseUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

// TwseClient is the opaque handle for the 23 TWSE endpoints. It hides
// `*svc/twse.Client` so cmd/ never imports svc/twse.
type TwseClient struct {
	inner *twse.Client
}

// NewTwseClient builds the default process-wide TWSE client tuned for
// TWSE's public REST API.
func NewTwseClient() *TwseClient {
	hc := httpx.NewClient(&httpx.Config{
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
		MaxBodyBytes:     0,
	})
	hc.Use(httpx.TWSEMiddleware(twseUserAgent))
	return &TwseClient{inner: twse.NewClient(hc)}
}

// NewTwseClientWithHTTP builds a TWSE client over a caller-supplied
// transport and base URL. Tests point it at an httptest server; callers
// needing custom QPS/retry policy supply their own `httpx.Client`.
func NewTwseClientWithHTTP(caller httpx.Caller, baseURL string) *TwseClient {
	return &TwseClient{inner: twse.NewClientWithURL(caller, baseURL)}
}

// BaseURL reports the TWSE origin this client dispatches against.
func (c *TwseClient) BaseURL() string { return c.inner.BaseURL() }

// Caller reports the underlying HTTP transport.
func (c *TwseClient) Caller() httpx.Caller { return c.inner.Caller() }

// Dispatch validates the endpoint name and its required flags, then runs
// the matching fetcher. It returns the raw envelope — JSON encoding is
// the caller's job.
func (c *TwseClient) Dispatch(ctx context.Context, endpoint, date string, opts url.Values) (any, error) {
	return twseDispatch(ctx, c.inner, endpoint, date, opts)
}

// TwseIsNoData reports whether err is a TWSE no-data error (sentinel
// ErrNoData, or a message containing "no data").
func TwseIsNoData(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, TwseErrNoData) || strings.Contains(err.Error(), "no data")
}
