// middleware.go — `Meta` request/response tag struct plus `RequestMiddleware` and `ResponseMiddleware` chains registered via `Client.Use` and `Client.UseAfter`. Capacity: 1 struct, 2 func types, 2 register methods, 2 runners.
package httpx

import (
	"fmt"
	"net/http"
	"time"
)

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