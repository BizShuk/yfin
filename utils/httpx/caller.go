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

// Caller is the transport contract for fetching HTTP resources. Relative
// targets are appended to a per-implementation base URL (Config.BaseURL),
// while absolute targets retain their own scheme and host. Query may be nil.
//
// IMPORTANT: Absolute targets are used verbatim even when Config.BaseURL is
// non-empty. This lets callers such as svc/scrape and svc/twse select a host
// per request without forcing a second concatenation. Relative targets still
// require a configured BaseURL.
//
// *Client implements Caller directly; tests can provide a stub.
//
// Implementer scope: as of this writing the only in-repo implementer is
// `*Client` itself. External packages should NOT add new implementations
// without re-checking the absolute-target contract and the
// error-wrapping conventions (status, attempts, duration are surfaced
// through `*Meta`, not the error string) — adding a second
// implementation would silently diverge callers' observability.
type Caller interface {
	Get(ctx context.Context, target string, query url.Values) ([]byte, *Meta, error)
}

// Get implements Caller. It resolves a relative target against Config.BaseURL
// or uses an absolute target verbatim, encodes query, performs a GET, validates
// the response status, decodes the body via readBody (gzip + MaxBodyBytes), and
// returns the body along with a populated `*Meta` describing the fetch.
//
// On non-2xx responses Get still returns the `*Meta` (with Status, Gzip,
// Attempts, Duration set) wrapped inside the error path; the body is
// discarded after a 1 KiB peek used to enrich the error message. On
// transport errors the response body is never read — the connection
// was never fully established, so there is nothing to drain. The base
// URL comes from Config.BaseURL for relative targets; absolute targets do not
// require it.
func (c *Client) Get(ctx context.Context, target string, query url.Values) ([]byte, *Meta, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, nil, fmt.Errorf("httpx: invalid target: %w", err)
	}
	if !u.IsAbs() {
		u, err = url.Parse(c.config.BaseURL + target)
		if err != nil {
			return nil, nil, fmt.Errorf("httpx: invalid target: %w", err)
		}
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
