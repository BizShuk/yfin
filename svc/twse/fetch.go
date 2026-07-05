// fetch.go — TWSE REST helper: `BaseURL`, `Client.FetchJSON[T]` generic decoder,
// `ErrNoData` sentinel + `StatOK`. Capacity: 1 `Client` per TWSE host +
// `statGetter` interface used by every endpoint fetcher.
package twse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/bizshuk/yfin/utils/httpx"
)

// BaseURL is the TWSE RESTful endpoint root. It is a `var` (not const) so
// tests and alternate deployments can override it via `NewClientWithURL`.
// `NewClient` reads it at construction time; tests that need a per-suite
// override should construct their Client with `NewClientWithURL`.
var BaseURL = "https://www.twse.com.tw/rwd/zh"

const (
	// StatOK is the value of `stat` field when TWSE returned data successfully.
	StatOK = "OK"
)

var (
	// ErrNoData is returned when TWSE responds with "沒有符合條件的資料" (or any
	// variant that contains the substring statNoData). It is a sentinel so callers
	// can errors.Is it.
	ErrNoData = errors.New("twse: no data for requested date")
)

// Client is a TWSE-scoped HTTP client. It owns the base URL and the
// underlying `httpx.Caller` that performs the actual transport — retry,
// backoff, rate-limit, circuit-breaker, gzip, body-cap and observability
// are all delegated to the caller.
type Client struct {
	caller  httpx.Caller
	baseURL string
}

// NewClient returns a Client whose base URL is captured from the
// package-level `BaseURL` at the time of call. Pass a caller wired with
// `httpx.NewClient(...)` that has `TWSEMiddleware` registered on it.
func NewClient(caller httpx.Caller) *Client {
	return &Client{caller: caller, baseURL: BaseURL}
}

// NewClientWithURL returns a Client that targets an explicit base URL.
// Use this for tests (httptest server), mirror deployments, or any
// non-default TWSE host.
func NewClientWithURL(caller httpx.Caller, baseURL string) *Client {
	if baseURL == "" {
		baseURL = BaseURL
	}
	return &Client{caller: caller, baseURL: baseURL}
}

// Caller exposes the underlying `httpx.Caller` (rarely needed; mostly for
// advanced callers that want to issue non-FetchJSON requests).
func (c *Client) Caller() httpx.Caller { return c.caller }

// BaseURL returns the base URL this client resolves paths against.
func (c *Client) BaseURL() string { return c.baseURL }

// FetchJSON performs a GET on `client.BaseURL() + path` with the supplied
// query params, then decodes the body into T. It automatically:
//   - appends `response=json` to the query string,
//   - returns ErrNoData when the body is empty (TWSE returns 200 + empty
//     body for some no-data cases) or when the envelope's `stat` field
//     contains the "no data" substring.
//
// `path` is the endpoint path (e.g. "/afterTrading/MI_INDEX"), appended
// to the client's base URL. All transport-level concerns (retry, body cap,
// status validation, gzip, observability) are owned by the injected
// caller.
//
// Non-2xx responses are surfaced by `Caller.Get` as a wrapped error,
// which FetchJSON re-wraps as `"twse: request failed: ..."`; there is
// no separate status check in this function.
//
// T must either be (or embed) *Response, or implement GetStat() string.
//
// This is a free generic function rather than a method on `*Client` only
// because Go 1.18+ allows generic methods on generic receivers, not on
// non-generic ones — the endpoint call sites use
// `FetchJSON[STOCK_DAYResponse](ctx, client, ...)` which is one keyword
// away from the method form and remains fully type-safe.
func FetchJSON[T any](ctx context.Context, client *Client, path string, query url.Values) (T, error) {
	var zero T
	q := url.Values{}
	for k, vs := range query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("response", "json")

	// Compose the absolute URL ourselves and pass it to caller.Get with an
	// empty Config.BaseURL on the injected caller (see buildTWSEClient) so
	// caller.Get's `path` argument is used verbatim. The local name
	// `absURL` documents the boundary contract clearly.
	absURL := client.baseURL + path
	body, _, err := client.caller.Get(ctx, absURL, q)
	if err != nil {
		return zero, fmt.Errorf("twse: request failed: %w", err)
	}
	if len(body) == 0 {
		return zero, ErrNoData
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
