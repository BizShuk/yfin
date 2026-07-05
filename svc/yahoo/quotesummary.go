// Fetches Yahoo quoteSummary modules with crumb auth and 401 retry.

package yahoo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// FetchQuoteSummary fetches raw quoteSummary JSON for the given modules,
// attaching the crumb and retrying once on 401.
func (c *Client) FetchQuoteSummary(ctx context.Context, symbol string, modules []string) ([]byte, error) {
	body, status, err := c.doQuoteSummary(ctx, symbol, modules)
	if err == nil && status == http.StatusUnauthorized && c.crumb != nil {
		c.crumb.Invalidate()
		body, status, err = c.doQuoteSummary(ctx, symbol, modules)
	}
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("quoteSummary %s: unexpected status %d", symbol, status)
	}
	return body, nil
}

func (c *Client) doQuoteSummary(ctx context.Context, symbol string, modules []string) ([]byte, int, error) {
	u, err := url.Parse(c.baseURL + "/v10/finance/quoteSummary/" + symbol)
	if err != nil {
		return nil, 0, err
	}
	q := url.Values{}
	q.Set("modules", strings.Join(modules, ","))
	if c.crumb != nil {
		crumb, cerr := c.crumb.Crumb(ctx)
		if cerr != nil {
			return nil, 0, cerr
		}
		q.Set("crumb", crumb)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}
