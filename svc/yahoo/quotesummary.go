// `FetchQuoteSummary` fetches raw quoteSummary JSON for a given module list. Capacity: 2 methods (`FetchQuoteSummary`, internal `doQuoteSummary`); single 401 retry after `CrumbManager.Invalidate`.
package yahoo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/bizshuk/yfin/utils/httpx"
)

// FetchQuoteSummary fetches raw quoteSummary JSON for the given modules,
// attaching the crumb and retrying once on 401.
func (c *Client) FetchQuoteSummary(ctx context.Context, symbol string, modules []string) ([]byte, error) {
	body, err := c.doQuoteSummary(ctx, symbol, modules)
	if err != nil {
		var statusErr *httpx.HTTPError
		if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusUnauthorized || c.crumb == nil {
			return nil, err
		}
		c.crumb.Invalidate()
		return c.doQuoteSummary(ctx, symbol, modules)
	}
	return body, nil
}

func (c *Client) doQuoteSummary(ctx context.Context, symbol string, modules []string) ([]byte, error) {
	u, err := url.Parse(c.baseURL + "/v10/finance/quoteSummary/" + symbol)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("modules", strings.Join(modules, ","))
	if c.crumb != nil {
		crumb, cerr := c.crumb.Crumb(ctx)
		if cerr != nil {
			return nil, cerr
		}
		q.Set("crumb", crumb)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("quoteSummary %s: %w", symbol, err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
