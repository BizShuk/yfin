// caller.go — `Caller` transport contract plus `Client.Get` GET helper (path + query -> body bytes + Meta). Capacity: 1 interface, 1 method.
package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ctxMetaKey is the unexported context key used by `Get` to thread the
// per-request `*Meta` through `Client.Do` and back out to the caller.
// Callers that invoke `Do` directly without pre-populating meta get a
// fresh `Meta` built from `req.URL`.
type ctxMetaKey struct{}

// Caller is the transport contract for fetching HTTP resources. The
// path argument is normally appended to a per-implementation base URL
// (Config.BaseURL); query may be nil.
//
// IMPORTANT: When Config.BaseURL is empty (""), the path argument is
// used verbatim as an absolute URL — this lets callers like svc/twse
// pre-compose their own host+path prefix without forcing a second
// concatenation. Implementations that take a BaseURL must therefore
// special-case "" to mean "use the caller's string as-is".
//
// *Client implements Caller directly; tests can provide a stub.
type Caller interface {
	Get(ctx context.Context, path string, query url.Values) ([]byte, *Meta, error)
}

// Get implements Caller. It builds the URL as Config.BaseURL + path,
// encodes query, performs a GET, validates the response status, decodes
// the body via readBody (gzip + MaxBodyBytes), and returns the body
// along with a populated `*Meta` describing the fetch.
//
// On non-2xx responses Get still returns the `*Meta` (with Status, Gzip,
// Attempts, Duration set) wrapped inside the error path; the body is
// discarded. The base URL comes from Config.BaseURL and must be set on
// the Client (use NewClient).
func (c *Client) Get(ctx context.Context, path string, query url.Values) ([]byte, *Meta, error) {
	u, err := url.Parse(c.config.BaseURL + path)
	if err != nil {
		return nil, nil, fmt.Errorf("httpx: invalid path: %w", err)
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	// Pre-populate meta so middleware / Do can write into it; Get reads
	// the same pointer back after Do returns.
	meta := &Meta{
		Host:     req.URL.Host,
		Endpoint: extractEndpoint(req.URL.Path),
	}
	ctx = context.WithValue(ctx, ctxMetaKey{}, meta)
	req = req.WithContext(ctx)

	start := time.Now()
	resp, err := c.Do(ctx, req)
	meta.Duration = time.Since(start)

	if err != nil {
		return nil, meta, fmt.Errorf("httpx: request failed: %w", err)
	}

	meta.Status = resp.StatusCode
	meta.Gzip = resp.Header.Get("Content-Encoding") == "gzip"

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
		return nil, meta, fmt.Errorf("httpx: status %d: %s", resp.StatusCode, string(body))
	}

	body, err := readBody(resp, c.config.MaxBodyBytes)
	meta.Bytes = len(body)
	return body, meta, err
}