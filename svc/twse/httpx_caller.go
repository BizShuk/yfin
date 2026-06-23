// Package twse: httpx_caller.go 為 Caller 介面之 *httpx.Client 配接器。
//
// HttpxCaller 把 production 用 httpx.Client 的呼叫 (含 retry /
// backoff / UA header) 收攏到單一物件,讓 FetchJSON 與各端點 Fetch* 只
// 看見 Caller 介面、與 transport 解耦。
package twse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/AmpyFin/yfinance-go/internal/httpx"
)

// HttpxCaller adapts *httpx.Client to the Caller interface. It owns the
// full HTTP lifecycle for one request: build URL, set method, do, read
// body, surface non-2xx as errors.
type HttpxCaller struct {
	Client *httpx.Client
}

// NewHttpxCaller wraps an existing *httpx.Client in a HttpxCaller.
func NewHttpxCaller(c *httpx.Client) *HttpxCaller { return &HttpxCaller{Client: c} }

// Call implements Caller. It appends `path` to BaseURL, encodes the
// supplied query, performs the GET via h.Client, and returns the body
// bytes. A non-2xx status becomes an error including the first 1 KiB
// of the response body for diagnostics.
func (h *HttpxCaller) Call(ctx context.Context, path string, query url.Values) ([]byte, error) {
	u, err := url.Parse(BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("twse: invalid path: %w", err)
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.Client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("twse: request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("twse: status %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}
