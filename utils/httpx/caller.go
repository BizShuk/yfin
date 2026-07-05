// Caller is the transport contract; Client.Call performs a GET and returns the body bytes.

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
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("httpx: status %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}
