// fetch.go — TWSE REST helper: `BaseURL`, `FetchJSON[T]` generic decoder, `ErrNoData` sentinel + `StatOK`. Capacity: shared transport for 23 endpoint fetchers.
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

	"github.com/bizshuk/yfin/config"
)

// BaseURL is the TWSE RESTful endpoint root. It is a `var` (not const) so
// tests can override it via httptest. It is the single owner of the TWSE
// host: FetchJSON joins it with each endpoint's path to form the full URL.
var BaseURL = "https://www.twse.com.tw/rwd/zh"

const (
	// StatOK is the value of `stat` field when TWSE returned data successfully.
	StatOK = "OK"

	// userAgent is sent on every request; TWSE rejects the default Go UA.
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
)

// ErrNoData is returned when TWSE responds with "沒有符合條件的資料" (or any
// variant that contains the substring statNoData). It is a sentinel so callers
// can errors.Is it.
var ErrNoData = errors.New("twse: no data for requested date")

// FetchJSON performs a GET on `BaseURL + path` with the supplied query
// params, then decodes the body into T. It automatically:
//   - appends `response=json` to the query string,
//   - returns ErrNoData when the body is empty (TWSE returns 200 + empty
//     body for some no-data cases) or when the envelope's `stat` field
//     contains the "no data" substring.
//
// `path` is the endpoint path (e.g. "/afterTrading/MI_INDEX"), appended
// to BaseURL — this package owns the full URL. The HTTP transport is the
// process-wide shared client pulled from internal/config (no per-call
// injection): a single host-agnostic *http.Client serves every endpoint.
// `query` is optional (nil OK); caller-supplied keys are preserved and
// `response=json` is always added on top.
//
// T must either be (or embed) *Response, or implement GetStat() string.
// Concrete DTOs typically embed `Response` and gain GetStat() via
// promotion; if the embedded name is shadowed, the DTO should provide
// its own GetStat() method.
func FetchJSON[T any](ctx context.Context, path string, query url.Values) (T, error) {
	var zero T
	q := url.Values{}
	for k, vs := range query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("response", "json")

	u, err := url.Parse(BaseURL + path)
	if err != nil {
		return zero, fmt.Errorf("twse: invalid path %q: %w", path, err)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return zero, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := config.HTTPClient().Do(req)
	if err != nil {
		return zero, fmt.Errorf("twse: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return zero, fmt.Errorf("twse: unexpected status %d: %s", resp.StatusCode, string(b))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("twse: read body: %w", err)
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
