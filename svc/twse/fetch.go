package twse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AmpyFin/yfinance-go/internal/httpx"
)

// BaseURL is the TWSE RESTful endpoint root. It is a `var` (not const) so
// tests can override it via httptest.
var BaseURL = "https://www.twse.com.tw/rwd/zh"

const (
	// StatOK is the value of `stat` field when TWSE returned data successfully.
	StatOK = "OK"
)

// ErrNoData is returned when TWSE responds with "沒有符合條件的資料" (or any
// variant that contains the substring statNoData). It is a sentinel so callers
// can errors.Is it.
var ErrNoData = errors.New("twse: no data for requested date")

// DefaultTimeout is the per-request timeout suggested by TWSE engineering notes.
const DefaultTimeout = 30 * time.Second

// FetchJSON performs a GET on `BaseURL + path` with optional query params and
// decodes the body into T. It automatically checks the envelope's `stat`
// field; if it indicates "no data" it returns ErrNoData.
//
// `path` is the endpoint path (e.g. "/afterTrading/MI_INDEX"), appended to
// BaseURL. `query` is optional (nil OK); `response=json` is added automatically.
//
// T must either be (or embed) *Response, or implement GetStat() string.
// Concrete DTOs typically embed `Response` and gain GetStat() via promotion;
// if the embedded name is shadowed, the DTO should provide its own
// GetStat() method.
func FetchJSON[T any](ctx context.Context, c *httpx.Client, path string, query url.Values) (T, error) {
	var zero T
	u, err := url.Parse(BaseURL + path)
	if err != nil {
		return zero, fmt.Errorf("twse: invalid path %q: %w", path, err)
	}
	q := url.Values{}
	for k, vs := range query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("response", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return zero, err
	}
	resp, err := c.Do(ctx, req)
	if err != nil {
		return zero, fmt.Errorf("twse: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return zero, fmt.Errorf("twse: unexpected status %d: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("twse: read body: %w", err)
	}

	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return zero, fmt.Errorf("twse: decode json: %w", err)
	}
	if err := checkStat(&out); err != nil {
		return zero, err
	}
	return out, nil
}

// statGetter is the optional contract that FetchJSON uses to read a
// response's `stat` field without reflection.
type statGetter interface {
	GetStat() string
}

// checkStat inspects the response via the statGetter interface. If the value
// doesn't expose GetStat() (e.g. a flat struct used in tests), this is a
// no-op. The actual stat is matched against the no-data substring (TWSE
// sometimes prefixes with "很抱歉，").
func checkStat(v any) error {
	g, ok := v.(statGetter)
	if !ok {
		return nil
	}
	stat := g.GetStat()
	if stat == "" || stat == StatOK {
		return nil
	}
	if strings.Contains(stat, statNoData) {
		return ErrNoData
	}
	return nil
}