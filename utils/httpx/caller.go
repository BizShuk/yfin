// caller.go — `Caller` transport contract plus `Client.Call` GET helper (path + query -> body bytes). Capacity: 1 interface, 1 method.
package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Caller is the transport contract for fetching HTTP resources. The path
// is appended to a per-implementation base URL; query may be nil.
//
// *Client implements Caller directly; tests can provide a stub.
type Caller interface {
	Call(ctx context.Context, path string, query url.Values) ([]byte, error)
}

// Call implements Caller. It builds the URL as Config.BaseURL + path,
// encodes query, performs a GET, validates the response status, and
// returns the body bytes. The base URL comes from Config.BaseURL and
// must be set on the Client (use NewClient).
//
// On 2xx responses the body is decoded via readBody, which honours
// gzip Content-Encoding and Config.MaxBodyBytes.
func (c *Client) Call(ctx context.Context, path string, query url.Values) ([]byte, error) {
	u, err := url.Parse(c.config.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("httpx: invalid path: %w", err)
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("httpx: request failed: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("httpx: status %d: %s", resp.StatusCode, string(body))
	}
	return readBody(resp, c.config.MaxBodyBytes)
}
