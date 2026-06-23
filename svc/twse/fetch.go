// Package twse 提供臺灣證券交易所 (Taiwan Stock Exchange, TWSE) 公開
// REST 端點的擷取、列解析與強型別 DTO。本檔定義 FetchJSON 共用函式、
// Caller 傳輸介面與 ErrNoData sentinel error。
//
// 設計重點:
//   - Caller 介面將傳輸層 (HTTP) 與業務層 (DTO 解碼、stat 檢查) 解耦,
//     方便用 httptest.Server 替換為測試 stub。
//   - FetchJSON 統一處理 URL 組裝、response=json 附加、JSON 解碼與
//     「沒有符合條件的資料」stat 字串偵測。
//   - 各端點檔 (mi_index.go、t86.go ...) 實作各自的 FetchXxx 函式,
//     全部以 Caller 為唯一外部相依。
package twse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
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

// Caller is the transport contract for fetching TWSE JSON. Production
// callers use *httpx.Client via HttpxCaller; tests can use a stub. The
// Caller is responsible for everything HTTP — building the full URL,
// setting headers, executing the request, checking status, and reading
// the body. It returns the raw response body (already known to be 2xx).
type Caller interface {
	Call(ctx context.Context, path string, query url.Values) ([]byte, error)
}

// FetchJSON performs a GET on `BaseURL + path` with the supplied query
// params, then decodes the body into T. It automatically:
//   - appends `response=json` to the query string,
//   - returns ErrNoData when the body is empty (TWSE returns 200 + empty
//     body for some no-data cases) or when the envelope's `stat` field
//     contains the "no data" substring.
//
// `path` is the endpoint path (e.g. "/afterTrading/MI_INDEX"), appended
// to BaseURL. `query` is optional (nil OK); caller-supplied keys are
// preserved and `response=json` is always added on top.
//
// T must either be (or embed) *Response, or implement GetStat() string.
// Concrete DTOs typically embed `Response` and gain GetStat() via
// promotion; if the embedded name is shadowed, the DTO should provide
// its own GetStat() method.
func FetchJSON[T any](ctx context.Context, c Caller, path string, query url.Values) (T, error) {
	var zero T
	q := url.Values{}
	for k, vs := range query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("response", "json")

	body, err := c.Call(ctx, path, q)
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
