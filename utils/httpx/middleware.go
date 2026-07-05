// middleware.go — `Meta` request/response tag struct plus `RequestMiddleware` and `ResponseMiddleware` chains registered via `Client.Use` and `Client.UseAfter`. Capacity: 1 struct, 2 func types, 2 register methods, 2 runners, 3 service-specific middleware factories.
package httpx

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// CrumbProvider is the neutral interface YahooMiddleware needs. The real
// implementation lives in svc/yahoo (which depends on httpx, not the
// other way around) — passing a CrumbProvider lets utils/httpx stay a
// leaf package while still wiring Yahoo's auth flow.
//
// *yahoo.CrumbManager satisfies this interface; tests may stub it.
type CrumbProvider interface {
	Crumb(ctx context.Context) (string, error)
}

// TWSEMiddleware returns a request middleware that pins User-Agent to
// the supplied string. Pass an empty string to disable the override
// (httpx.Client already sets UA from Config.UserAgent). TWSE rejects
// requests that don't carry a browser-like UA, so production wiring
// supplies the chrome-on-macOS string.
func TWSEMiddleware(userAgent string) RequestMiddleware {
	return func(req *http.Request, meta *Meta) error {
		if userAgent == "" {
			return nil
		}
		req.Header.Set("User-Agent", userAgent)
		return nil
	}
}

// YahooMiddleware returns a request middleware that injects Yahoo's
// auth crumb into the outgoing request's URL query (`crumb=<value>`).
// It calls CrumbProvider.Crumb on every request; CrumbManager caches the
// value internally so the upstream `/v1/test/getcrumb` round-trip is a
// once-per-process cost.
func YahooMiddleware(cm CrumbProvider) RequestMiddleware {
	return func(req *http.Request, meta *Meta) error {
		if cm == nil {
			return nil
		}
		crumb, err := cm.Crumb(req.Context())
		if err != nil {
			return fmt.Errorf("yahoo crumb: %w", err)
		}
		q := req.URL.Query()
		q.Set("crumb", crumb)
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// ScrapeMiddleware returns a request middleware that overrides
// User-Agent with a browser-like string and sets `Accept` /
// `Accept-Language` headers. Some hosts (notably finance.yahoo.com)
// reject requests that don't carry a full set of browser headers.
func ScrapeMiddleware(userAgent string) RequestMiddleware {
	return func(req *http.Request, meta *Meta) error {
		if userAgent != "" {
			req.Header.Set("User-Agent", userAgent)
		}
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		return nil
	}
}

// Meta is the per-request side-channel that middleware reads and writes.
// It is constructed at the top of `Client.Do` (Host + Endpoint populated
// eagerly) and threaded through every registered middleware. Status /
// Bytes / Duration / Attempts / Gzip are filled in by the response path
// of later tasks and may stay zero-valued for now.
type Meta struct {
	Status   int
	Bytes    int
	Duration time.Duration
	Attempts int
	Host     string
	Endpoint string // e.g. "twse:STOCK_DAY", "yahoo:bars_1d"
	Gzip     bool
}

// RequestMiddleware mutates or short-circuits an outgoing request. A
// non-nil return aborts the chain; `Client.Do` wraps the error with the
// failing middleware index and propagates it.
type RequestMiddleware func(req *http.Request, meta *Meta) error

// ResponseMiddleware observes a successful response. Errors are ignored;
// response middleware is for telemetry / header decoration only.
type ResponseMiddleware func(resp *http.Response, meta *Meta)

// Use appends one or more request middleware to the chain. Middleware run
// in the order they were registered. Typically called once during
// `NewClient`-equivalent setup, before any concurrent request.
func (c *Client) Use(mw ...RequestMiddleware) {
	c.reqMW = append(c.reqMW, mw...)
}

// UseAfter appends one or more response middleware. They fire on the
// success path of `Client.Do`, after retry / circuit / rate-limit have
// produced a 2xx response.
func (c *Client) UseAfter(mw ...ResponseMiddleware) {
	c.respMW = append(c.respMW, mw...)
}

// runRequestMW executes every registered request middleware in order,
// wrapping the first error with the failing index for diagnosability.
// Panics are not recovered — keep them simple.
func (c *Client) runRequestMW(req *http.Request, meta *Meta) error {
	for i, mw := range c.reqMW {
		if err := mw(req, meta); err != nil {
			return fmt.Errorf("middleware[%d]: %w", i, err)
		}
	}
	return nil
}

// runResponseMW executes every registered response middleware in order.
// Response middleware cannot fail the request; their return values are
// ignored. Panics are not recovered — keep them simple.
func (c *Client) runResponseMW(resp *http.Response, meta *Meta) {
	for _, mw := range c.respMW {
		mw(resp, meta)
	}
}
