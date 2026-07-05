# TWSE 端點 svc/ Package 實作計畫

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `README.tsme.md` 21 個 TWSE RESTful 端點完整實作為 Go `svc/twse` 套件,提供 downloader + parser + JSON DTO + 統一 registry;並加 `yfin twse` CLI 子指令可直接呼叫任一端點。

**Architecture:**
- 採用 **JSON 為主** 解析(TWSE 官方 `response=json` 回傳結構化 `tables[]` 與 `stat` 欄位,易驗證「無資料」、易清洗)。
- **One package + one file per endpoint**:`svc/twse/<endpoint>.go`,21 個端點對應 21 個檔案 + 1 個 `registry.go`。
- 通用 helper:`svc/twse/fetch.go` 提供 `FetchJSON[T any](ctx, client, path, query) (T, error)`,所有端點共用同一個 GET + JSON decode + stat 檢查邏輯。
- 共用 HTTP client:複用現有 `internal/httpx.Client`(已含 cookie jar、backoff、rate limit)。
- CLI:`cmd/yfin/twse.go` 加 21 個 cobra subcommand + 1 個 dispatcher `--endpoint <name>` 簡化用法。

**Tech Stack:** Go 1.23+, encoding/json, net/http, 既有 internal/httpx, testify, cobra (CLI)。

**Refresh tiers / rate limiting:** 沿用 README.tsme.md §「RESTful 端點呼叫通則與參數設計」:`TWSE_REQUEST_INTERVAL` 0.5-1.0 秒隨機延遲(實作在底層 limiter);每請求 `Timeout=30s`;HTTP 5xx 指數退避重試 3 次。

---

## 21 個端點對應檔案

### 盤後交易資訊 (8 個)
- `mi_index.go` — `MI_INDEX` 每日收盤行情
- `stock_day.go` — `STOCK_DAY` 個股日成交
- `bwibbu_d.go` — `BWIBBU_d` 本益比/殖利率/PB
- `mi_index_plus.go` — `MI_INDEX_PLUS` 盤後定價交易
- `mi_index_odd.go` — `MI_INDEX_ODD` 零股交易
- `mi_5mins.go` — `MI_5MINS` 每 5 秒委託成交
- `twtb4u.go` — `TWTB4U` 當日沖銷
- `mi_margn.go` — `MI_MARGN` 融資融券餘額

### 三大法人 (6 個)
- `t86.go` — `T86` 三大法人買賣超日報
- `mi_qfiis.go` — `MI_QFIIS` 外資陸資持股
- `bfi82u.go` — `BFI82U` 三大法人買賣金額
- `twt38u.go` — `TWT38U` 外資陸資買賣超彙總
- `twt43u.go` — `TWT43U` 投信買賣超彙總
- `twt44u.go` — `TWT44U` 自營商買賣超彙總

### 鉅額交易 (4 個)
- `bfiauu.go` — `BFIAUU` 鉅額交易(逐筆)
- `bfiauu_stock.go` — `BFIAUU_STOCK` 單一證券鉅額
- `bfimuu.go` — `BFIMUU` 鉅額交易月報
- `bfiauu_year.go` — `BFIAUU_YEAR` 鉅額交易年報

### 統計報表 (5 個)
- `fmtqik.go` — `FMTQIK` 臺股指數及交易量
- `stock_day_avg.go` — `STOCK_DAY_AVG` 個股月均價
- `fmsrfk.go` — `FMSRFK` 個股月成交
- `bfiamu.go` — `BFIAMU` 每日各類指數成交量
- `mi_week.go` — `MI_WEEK` 股票市值週報

> 註:`BFIAUU_STOCK` 與 `BFIAUU` 端點相同(只是帶 `stockNo` 參數),所以 `bfiauu.go` 內的 `Fetch` 接受可選 `StockNo` 參數;獨立 `bfiauu_stock.go` 提供薄封裝。

---

## 共用元件檔案

- `svc/twse/fetch.go` — `FetchJSON[T]`,`Endpoint` 介面,`Response[T]` 通用 envelope
- `svc/twse/registry.go` — `Registry` 把 21 個端點映射到 name → fetcher
- `svc/twse/types.go` — 共用型別:`Response[T]`,`Table`,`Field`,`StatError`
- `svc/twse/fetch_test.go` — `FetchJSON` 與 error 處理測試
- `svc/twse/registry_test.go` — `Registry` 21 端點全覆蓋測試

---

## Task 1: 共用型別 + FetchJSON helper

**Files:**
- Create: `svc/twse/fetch.go`
- Create: `svc/twse/types.go`
- Test: `svc/twse/fetch_test.go`

### Step 1: 寫失敗測試

```go
// svc/twse/fetch_test.go
package twse

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bizshuk/yfinance-go/internal/httpx"
	"github.com/stretchr/testify/require"
)

func TestFetchJSON_Decodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stat":"OK","title":"MI_INDEX","fields":["a","b"],"data":[["1","2"],["3","4"]]}`))
	}))
	defer srv.Close()

	got, err := FetchJSON[TestResponse](context.Background(), httpx.NewClient(httpx.DefaultConfig()), srv.URL, nil)
	require.NoError(t, err)
	require.Equal(t, "OK", got.Stat)
	require.Equal(t, "MI_INDEX", got.Title)
	require.Len(t, got.Data, 2)
}

func TestFetchJSON_NoDataReturnsErrNoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stat":"沒有符合條件的資料","fields":[],"data":[]}`))
	}))
	defer srv.Close()

	_, err := FetchJSON[TestResponse](context.Background(), httpx.NewClient(httpx.DefaultConfig()), srv.URL, nil)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoData))
}

func TestFetchJSON_StatAtTopLevel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stat":"OK","data":[]}`))
	}))
	defer srv.Close()

	got, err := FetchJSON[TestResponse](context.Background(), httpx.NewClient(httpx.DefaultConfig()), srv.URL, nil)
	require.NoError(t, err)
	require.Equal(t, "OK", got.Stat)
}

// TestResponse is a sample struct matching the TWSE JSON envelope.
type TestResponse struct {
	Stat   string     `json:"stat"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}
```

### Step 2: 執行,確認 fail
`go test ./svc/twse/ -v` → FAIL (`FetchJSON` undefined)。

### Step 3: 實作

```go
// svc/twse/types.go
package twse

// StatNoData is TWSE's traditional "no data" message (varies by endpoint, so we
// substring-check). The exact string from TWSE: "很抱歉，沒有符合條件的資料!" plus
// the Latin "No data" variant.
const statNoData = "沒有符合條件的資料"

// Response is the common TWSE JSON envelope; concrete endpoints embed this
// and add their own extra fields (e.g. "date", "stockNo").
type Response struct {
	Stat    string     `json:"stat"`
	Title   string     `json:"title,omitempty"`
	Fields  []string   `json:"fields"`
	Data    [][]string `json:"data"`
	Notes   []string   `json:"notes,omitempty"`
	Total   int        `json:"total,omitempty"`
	// Catch-all for endpoint-specific fields (date, stockNo, etc.)
	Extra map[string]any `json:"-"`
}
```

```go
// svc/twse/fetch.go
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

	"github.com/bizshuk/yfinance-go/internal/httpx"
)

const (
	BaseURL = "https://www.twse.com.tw/rwd/zh"
	StatOK  = "OK"
)

var ErrNoData = errors.New("twse: no data for requested date")

// FetchJSON performs a GET on `path?query` and decodes the body into T. It
// automatically checks the envelope's `stat` field; if it indicates "no data"
// it returns ErrNoData.
//
// `path` is the endpoint path (e.g. "/afterTrading/MI_INDEX"), joined with BaseURL.
// `query` is optional (nil OK); `response=json` is added automatically.
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

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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
	if err := checkStat(out); err != nil {
		return zero, err
	}
	return out, nil
}

// checkStat inspects the response's `stat` field via reflection-free type
// assertion to a struct with a `Stat string` field. If the embedded struct
// doesn't expose Stat, this is a no-op.
func checkStat(v any) error {
	type statGetter interface{ GetStat() string }
	if g, ok := v.(statGetter); ok {
		stat := g.GetStat()
		if stat != "" && stat != StatOK && strings.Contains(stat, statNoData) {
			return ErrNoData
		}
		if stat == "沒有符合條件的資料" {
			return ErrNoData
		}
	}
	return nil
}

// DefaultTimeout is the per-request timeout suggested by TWSE engineering notes.
const DefaultTimeout = 30 * time.Second
```

> 註:`checkStat` 需要每個回應 DTO 暴露 `GetStat() string`。簡化做法:每個 DTO 都用 `Response` 內嵌(`type MIIndexResponse struct { Response; Date string; StockNo string; ... }`),然後 `GetStat()` 由內嵌自動提供。**先 commit 這個版本(TDD 通過),再以同樣模式實作每個端點的 DTO。**

### Step 4: 執行確認通過 + Commit
`go test ./svc/twse/ -v` → PASS
```
git add svc/twse/fetch.go svc/twse/types.go svc/twse/fetch_test.go
git commit -m "feat(svc/twse): add FetchJSON helper + Response envelope"
```

---

## Task 2: 統一 registry(21 端點 dispatch)

**Files:**
- Create: `svc/twse/registry.go`
- Test: `svc/twse/registry_test.go`

### Step 1: 寫失敗測試

```go
// svc/twse/registry_test.go
package twse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistry_CoversAllEndpoints(t *testing.T) {
	want := []string{
		// afterTrading
		"MI_INDEX", "STOCK_DAY", "BWIBBU_d", "MI_INDEX_PLUS", "MI_INDEX_ODD",
		"MI_5MINS", "TWTB4U", "MI_MARGN",
		// fund
		"T86", "MI_QFIIS", "BFI82U", "TWT38U", "TWT43U", "TWT44U",
		// block
		"BFIAUU", "BFIAUU_STOCK", "BFIMUU", "BFIAUU_YEAR",
		// statistics
		"FMTQIK", "STOCK_DAY_AVG", "FMSRFK", "BFIAMU", "MI_WEEK",
	}
	for _, name := range want {
		_, ok := Registry[name]
		require.Truef(t, ok, "endpoint %q missing from registry", name)
	}
	require.Len(t, Registry, len(want))
}
```

### Step 2: 執行確認 fail
`go test ./svc/twse/ -run TestRegistry -v` → FAIL (`Registry` undefined)。

### Step 3: 實作(佔位實作,後續 Tasks 補上對應的 Endpoint 結構)

```go
// svc/twse/registry.go
package twse

import (
	"context"
	"net/url"
	"time"

	"github.com/bizshuk/yfinance-go/internal/httpx"
)

// Fetcher is the per-endpoint contract: build a query, call FetchJSON, return
// the typed DTO. Each endpoint file provides its own concrete Fetcher.
type Fetcher func(ctx context.Context, c *httpx.Client, baseDate string, opts url.Values) (any, error)

// Endpoint describes a TWSE endpoint for CLI help/dispatch.
type Endpoint struct {
	Name        string
	Board       string // "afterTrading" | "fund" | "block" | "statistics" | "marginTrading"
	Path        string
	Description string
	NeedsStock  bool
	NeedsMonth  bool
	Fetch       Fetcher
}

// Registry maps endpoint code (e.g. "MI_INDEX") to its descriptor. 21 entries.
var Registry = map[string]Endpoint{
	"MI_INDEX":       {Name: "MI_INDEX", Board: "afterTrading", Path: "/afterTrading/MI_INDEX", Description: "每日收盤行情"},
	"STOCK_DAY":      {Name: "STOCK_DAY", Board: "afterTrading", Path: "/afterTrading/STOCK_DAY", Description: "個股日成交資訊", NeedsStock: true},
	"BWIBBU_d":       {Name: "BWIBBU_d", Board: "afterTrading", Path: "/afterTrading/BWIBBU_d", Description: "個股日本益比、殖利率及股價淨值比"},
	"MI_INDEX_PLUS":  {Name: "MI_INDEX_PLUS", Board: "afterTrading", Path: "/afterTrading/MI_INDEX_PLUS", Description: "盤後定價交易"},
	"MI_INDEX_ODD":   {Name: "MI_INDEX_ODD", Board: "afterTrading", Path: "/afterTrading/MI_INDEX_ODD", Description: "零股交易行情單"},
	"MI_5MINS":       {Name: "MI_5MINS", Board: "afterTrading", Path: "/afterTrading/MI_5MINS", Description: "每5秒委託成交統計"},
	"TWTB4U":         {Name: "TWTB4U", Board: "afterTrading", Path: "/afterTrading/TWTB4U", Description: "當日沖銷交易標的及統計"},
	"MI_MARGN":       {Name: "MI_MARGN", Board: "marginTrading", Path: "/marginTrading/MI_MARGN", Description: "融資融券餘額"},
	"T86":            {Name: "T86", Board: "fund", Path: "/fund/T86", Description: "三大法人買賣超日報"},
	"MI_QFIIS":       {Name: "MI_QFIIS", Board: "fund", Path: "/fund/MI_QFIIS", Description: "外資及陸資投資持股統計"},
	"BFI82U":         {Name: "BFI82U", Board: "fund", Path: "/fund/BFI82U", Description: "三大法人買賣金額統計表"},
	"TWT38U":         {Name: "TWT38U", Board: "fund", Path: "/fund/TWT38U", Description: "外資及陸資買賣超彙總表"},
	"TWT43U":         {Name: "TWT43U", Board: "fund", Path: "/fund/TWT43U", Description: "投信買賣超彙總表"},
	"TWT44U":         {Name: "TWT44U", Board: "fund", Path: "/fund/TWT44U", Description: "自營商買賣超彙總表"},
	"BFIAUU":         {Name: "BFIAUU", Board: "block", Path: "/block/BFIAUU", Description: "鉅額交易日成交資訊"},
	"BFIAUU_STOCK":   {Name: "BFIAUU_STOCK", Board: "block", Path: "/block/BFIAUU", Description: "單一證券日成交資訊", NeedsStock: true},
	"BFIMUU":         {Name: "BFIMUU", Board: "block", Path: "/block/BFIMUU", Description: "鉅額交易月成交資訊", NeedsMonth: true},
	"BFIAUU_YEAR":    {Name: "BFIAUU_YEAR", Board: "block", Path: "/block/BFIAUU_YEAR", Description: "鉅額交易年成交資訊"},
	"FMTQIK":         {Name: "FMTQIK", Board: "statistics", Path: "/exchangeReport/FMTQIK", Description: "臺股指數及交易量表", NeedsMonth: true},
	"STOCK_DAY_AVG":  {Name: "STOCK_DAY_AVG", Board: "statistics", Path: "/exchangeReport/STOCK_DAY_AVG", Description: "個股月均價", NeedsStock: true, NeedsMonth: true},
	"FMSRFK":         {Name: "FMSRFK", Board: "statistics", Path: "/exchangeReport/FMSRFK", Description: "個股月成交資訊", NeedsStock: true},
	"BFIAMU":         {Name: "BFIAMU", Board: "statistics", Path: "/afterTrading/BFIAMU", Description: "每日各類指數成交量值"},
	"MI_WEEK":        {Name: "MI_WEEK", Board: "statistics", Path: "/statistics/MI_WEEK", Description: "股票市值週報"},
}

// DefaultClient returns an httpx client configured for TWSE (timeout 30s, modest QPS).
func DefaultClient() *httpx.Client {
	cfg := httpx.DefaultConfig()
	cfg.Timeout = 30 * time.Second
	return httpx.NewClient(cfg)
}
```

### Step 4: 執行確認通過 + Commit
`go test ./svc/twse/ -run TestRegistry -v` → PASS
```
git add svc/twse/registry.go svc/twse/registry_test.go
git commit -m "feat(svc/twse): endpoint registry (21 endpoints across 4 boards)"
```

---

## Task 3..23: 21 個端點各一個 Task(可並行)

> 為節省篇幅,每個端點用同一個樣板。Task 3 是樣板(TWSE 最常用的 MI_INDEX),Task 4-23 用 placeholder 結構簡述,實作時 copy-paste 即可。

### 樣板(以 MI_INDEX 為例)

**Files:**
- Create: `svc/twse/mi_index.go`
- Test: `svc/twse/mi_index_test.go`

#### Step 1: 寫失敗測試

```go
// svc/twse/mi_index_test.go
package twse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMIIndex_DecodesAndParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/rwd/zh/afterTrading/MI_INDEX", r.URL.Path)
		require.Equal(t, "json", r.URL.Query().Get("response"))
		require.Equal(t, "20221230", r.URL.Query().Get("date"))
		_, _ = w.Write([]byte(`{
		  "stat":"OK","title":"每日收盤行情",
		  "fields":["指數","收盤指數","漲跌點數","漲跌百分比"],
		  "data":[["發行量加權股價指數","14104.20","-50.10","-0.35%"]],
		  "date":"20221230"
		}`))
	}))
	defer srv.Close()
	// Redirect the package BaseURL via test override.
	oldBase := BaseURL
	BaseURL = srv.URL + "/rwd/zh"
	defer func() { BaseURL = oldBase }()

	resp, err := FetchMIIndex(context.Background(), DefaultClient(), "20221230", nil)
	require.NoError(t, err)
	require.Equal(t, "OK", resp.GetStat())
	require.Len(t, resp.Data, 1)
	row, err := ParseMIIndexRow(resp.Data[0])
	require.NoError(t, err)
	require.Equal(t, "發行量加權股價指數", row.IndexName)
}

func TestMIIndex_NoData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stat":"很抱歉，沒有符合條件的資料!","fields":[],"data":[]}`))
	}))
	defer srv.Close()
	oldBase := BaseURL
	BaseURL = srv.URL + "/rwd/zh"
	defer func() { BaseURL = oldBase }()

	_, err := FetchMIIndex(context.Background(), DefaultClient(), "20221230", nil)
	require.ErrorIs(t, err, ErrNoData)
}
```

#### Step 2: 實作

```go
// svc/twse/mi_index.go
package twse

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/bizshuk/yfinance-go/internal/httpx"
)

// MIIndexResponse is the JSON envelope for /afterTrading/MI_INDEX.
type MIIndexResponse struct {
	Response
	Date string `json:"date"`
}

func (r *MIIndexResponse) GetStat() string { return r.Response.Stat }

// MIIndexRow is one parsed row from the data table.
type MIIndexRow struct {
	IndexName    string
	Close        float64
	Change       float64
	ChangePct    float64
}

// FetchMIIndex retrieves the daily closing market data.
func FetchMIIndex(ctx context.Context, c *httpx.Client, date string, opts url.Values) (any, error) {
	if date == "" {
		return nil, fmt.Errorf("twse/MI_INDEX: date is required")
	}
	q := url.Values{}
	q.Set("date", date)
	q.Set("type", "ALL")
	for k, vs := range opts {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return FetchJSON[MIIndexResponse](ctx, c, "/afterTrading/MI_INDEX", q)
}

// ParseMIIndexRow converts one raw `data` row into a typed struct. Tolerant
// to missing/empty cells.
func ParseMIIndexRow(row []string) (MIIndexRow, error) {
	if len(row) < 4 {
		return MIIndexRow{}, fmt.Errorf("MI_INDEX: row too short: %d cols", len(row))
	}
	return MIIndexRow{
		IndexName: row[0],
		Close:     parseFloat(row[1]),
		Change:    parseFloat(row[2]),
		ChangePct: parsePercent(row[3]),
	}, nil
}

func parseFloat(s string) float64 {
	s = stripNumberFmt(s)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parsePercent(s string) float64 {
	s = stripNumberFmt(s)
	s = trimSuffix(s, "%")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func stripNumberFmt(s string) string {
	s = trimAll(s, ",\"")
	s = removeComma(s)
	return s
}

func trimAll(s, cutset string) string {
	for len(s) > 0 && contains(cutset, s[:1]) {
		s = s[1:]
	}
	for len(s) > 0 && contains(cutset, s[len(s)-1:]) {
		s = s[:len(s)-1]
	}
	return s
}

func trimSuffix(s, suf string) string {
	if len(s) >= len(suf) && s[len(s)-len(suf):] == suf {
		return s[:len(s)-len(suf)]
	}
	return s
}

func removeComma(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r != ',' {
			out = append(out, r)
		}
	}
	return string(out)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

> 註:`parseFloat` / `parsePercent` / `stripNumberFmt` 等是共用 helper,放進 `svc/twse/parse.go`(在 Task 3 內 commit)。其他 20 個端點的 parser 函式會共用。

#### Step 3: 在 registry 內補上對應 Fetcher

回頭編輯 `svc/twse/registry.go`:
```go
"MI_INDEX": {Name: "MI_INDEX", Board: "afterTrading", Path: "/afterTrading/MI_INDEX", Description: "每日收盤行情",
    Fetch: func(ctx context.Context, c *httpx.Client, date string, opts url.Values) (any, error) {
        return FetchMIIndex(ctx, c, date, opts)
    }},
```

#### Step 4: 執行確認通過 + Commit
`go test ./svc/twse/ -run TestMIIndex -v` → PASS
```
git add svc/twse/mi_index.go svc/twse/mi_index_test.go svc/twse/parse.go svc/twse/registry.go
git commit -m "feat(svc/twse): MI_INDEX endpoint (daily closing market data)"
```

---

### 其餘 20 個端點(Tasks 4-23,逐一實作)

每個 Task 結構完全相同(測試 → 實作 → registry 補 fetch → 測試通過 → commit)。差異在:

| Task | 端點 | 檔名 | 特殊參數 | 特殊解析 |
| --- | --- | --- | --- | --- |
| 4 | STOCK_DAY | `stock_day.go` | `stockNo` | 個股 OHLC + 成交 |
| 5 | BWIBBU_d | `bwibbu_d.go` | `selectType=ALL` | PER / 殖利率 / PB 三欄 |
| 6 | MI_INDEX_PLUS | `mi_index_plus.go` | — | 盤後定價 |
| 7 | MI_INDEX_ODD | `mi_index_odd.go` | — | 零股 |
| 8 | MI_5MINS | `mi_5mins.go` | — | 5 秒委買委賣 |
| 9 | TWTB4U | `twtb4u.go` | — | 當沖標的 + 統計 |
| 10 | MI_MARGN | `mi_margn.go` | `selectType=ALL` | 融資/融券/借券 |
| 11 | T86 | `t86.go` | `selectType=ALL` | 5 個子項買賣超(外資/投信/自營商) |
| 12 | MI_QFIIS | `mi_qfiis.go` | `selectType=ALL` | 外資持股 |
| 13 | BFI82U | `bfi82u.go` | `type=day` | 三大法人金額 |
| 14 | TWT38U | `twt38u.go` | — | 外資彙總 |
| 15 | TWT43U | `twt43u.go` | — | 投信彙總 |
| 16 | TWT44U | `twt44u.go` | — | 自營商彙總 |
| 17 | BFIAUU | `bfiauu.go` | `stockNo?` | 鉅額逐筆 |
| 18 | BFIAUU_STOCK | `bfiauu_stock.go` | `stockNo=*` | 薄封裝(委派 `FetchBFIAUU`) |
| 19 | BFIMUU | `bfimuu.go` | `date=YYYYMM01` | 月報 |
| 20 | BFIAUU_YEAR | `bfiauu_year.go` | `date=YYYY0101` | 年報 |
| 21 | FMTQIK | `fmtqik.go` | `date=YYYYMMDD` | 月指數 + 成交量 |
| 22 | STOCK_DAY_AVG | `stock_day_avg.go` | `stockNo+date` | 月均價 |
| 23 | FMSRFK | `fmsrfk.go` | `stockNo+date` | 月成交(年維度) |
| 24 | BFIAMU | `bfiamu.go` | `date=YYYYMMDD` | 各類指數量值 |
| 25 | MI_WEEK | `mi_week.go` | `date=YYYYMMDD` | 市值週報 |

> 註:原文有 21 個端點表格(8+6+4+5=23 行,但 `BFIAUU_STOCK` 跟 `BFIAUU` 同端點、`MI_MARGN` 屬 `marginTrading` 板塊獨立計入),此處展開為 **25 個 Task**。每 Task 獨立 commit + 獨立測試。

---

## Task 26: CLI 子指令 `yfin twse`

**Files:**
- Create: `cmd/yfin/twse.go`
- Test: `cmd/yfin/twse_test.go`

### Step 1: 寫失敗測試

```go
// cmd/yfin/twse_test.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bizshuk/yfinance-go/svc/twse"
	"github.com/stretchr/testify/require"
)

func TestRunTwseEndpoint_HitsRegistryAndWritesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stat":"OK","fields":["a"],"data":[["x"]],"date":"20231015"}`))
	}))
	defer srv.Close()
	old := twse.BaseURL
	twse.BaseURL = srv.URL + "/rwd/zh"
	defer func() { twse.BaseURL = old }()

	var out bytes.Buffer
	err := runTwseEndpoint(context.Background(), twse.DefaultClient(),
		"MI_INDEX", "20231015", nil, &out)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	require.Equal(t, "OK", got["stat"])
	require.Equal(t, "20231015", got["date"])
}

func TestRunTwseEndpoint_UnknownName(t *testing.T) {
	var out bytes.Buffer
	err := runTwseEndpoint(context.Background(), twse.DefaultClient(),
		"NONEXISTENT", "20231015", nil, &out)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "unknown endpoint"))
}
```

### Step 2: 執行確認 fail
`go test ./cmd/yfin/ -run TestRunTwseEndpoint -v` → FAIL (`runTwseEndpoint` undefined)。

### Step 3: 實作

```go
// cmd/yfin/twse.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/bizshuk/yfinance-go/internal/httpx"
	"github.com/bizshuk/yfinance-go/svc/twse"
	"github.com/spf13/cobra"
)

var (
	twseEndpoint string
	twseDate     string
	twseStockNo  string
)

var twseCmd = &cobra.Command{
	Use:   "twse",
	Short: "Query Taiwan Stock Exchange (TWSE) public REST endpoints",
	RunE:  runTwse,
}

func init() {
	twseCmd.Flags().StringVar(&twseEndpoint, "endpoint", "",
		"Endpoint code (e.g. MI_INDEX, T86, BFIAUU, FMTQIK). Run 'yfin twse --help' for list.")
	twseCmd.Flags().StringVar(&twseDate, "date", "", "Query date (YYYYMMDD)")
	twseCmd.Flags().StringVar(&twseStockNo, "stock", "", "Stock number (for STOCK_DAY, BFIAUU_STOCK, STOCK_DAY_AVG, FMSRFK)")
}

func runTwse(cmd *cobra.Command, args []string) error {
	if twseEndpoint == "" {
		return fmt.Errorf("must specify --endpoint; available: %v", registryKeys())
	}
	opts := url.Values{}
	if twseStockNo != "" {
		opts.Set("stockNo", twseStockNo)
	}
	return runTwseEndpoint(cmd.Context(), twse.DefaultClient(), twseEndpoint, twseDate, opts, cmd.OutOrStdout())
}

// runTwseEndpoint is the unit-testable core: build query, fetch via registry,
// write JSON to w. It is intentionally a free function so tests can call it
// without a Cobra command.
func runTwseEndpoint(ctx context.Context, c *httpx.Client, name, date string, opts url.Values, w io.Writer) error {
	ep, ok := twse.Registry[name]
	if !ok {
		return fmt.Errorf("unknown endpoint %q; available: %v", name, registryKeys())
	}
	if ep.Fetch == nil {
		return fmt.Errorf("endpoint %q has no fetcher wired", name)
	}
	if date == "" {
		return fmt.Errorf("--date is required (YYYYMMDD)")
	}
	if ep.NeedsStock && opts.Get("stockNo") == "" {
		return fmt.Errorf("endpoint %q requires --stock", name)
	}
	result, err := ep.Fetch(ctx, c, date, opts)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", name, err)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func registryKeys() []string {
	out := make([]string, 0, len(twse.Registry))
	for k := range twse.Registry {
		out = append(out, k)
	}
	return out
}
```

`runTwseEndpoint` 同時在 `cmd/yfin/main.go` 註冊:`rootCmd.AddCommand(twseCmd)`(一行)。

### Step 4: 執行確認通過 + Commit
`go test ./cmd/yfin/ -run TestRunTwseEndpoint -v` → PASS
```
git add cmd/yfin/twse.go cmd/yfin/twse_test.go cmd/yfin/main.go
git commit -m "feat(cli): yfin twse subcommand (query any of 21 endpoints)"
```

---

## Task 27: 整合測試 + 文件

**Files:**
- Create: `tests/twse_integration_test.go`(可選,僅在有網路時跑)
- Modify: `README.md` — 新增 `## 🏛️ TWSE Public Data (svc/twse)` 段

### Step 1: 整合測試骨架(不連真實 API,只驗證 registry 與 fetch 流程)

```go
// tests/twse_integration_test.go
//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/bizshuk/yfinance-go/svc/twse"
	"github.com/stretchr/testify/require"
)

func TestIntegration_All21Endpoints_HaveFetchers(t *testing.T) {
	for name, ep := range twse.Registry {
		require.NotNilf(t, ep.Fetch, "endpoint %s missing fetcher", name)
	}
}
```

### Step 2: 文件
在 `README.md` 新增:

```
## 🏛️ TWSE Public Data (svc/twse)

`svc/twse` 套件完整封裝 21 個 TWSE RESTful 端點(盤後交易/三大法人/鉅額交易/統計報表),全部走 `response=json` 並自動驗證「無資料」。

```bash
# 直接呼叫任一端點
yfin twse --endpoint MI_INDEX --date 20221230
yfin twse --endpoint T86 --date 20231015
yfin twse --endpoint BFIAUU --date 20240110
yfin twse --endpoint FMTQIK --date 20210326

# 需要 stock 參數的端點
yfin twse --endpoint STOCK_DAY --date 20240901 --stock 2330
```

對應 Python:未實作。Go 是 TWSE 開放資料的第一個結構化 binding。
```

### Step 3: 全套件測試

`go test ./...` → 全綠

### Step 4: Commit
```
git add tests/twse_integration_test.go README.md
git commit -m "test+docs: TWSE integration skeleton + README"
```

---

## Self-Review Checklist

- Spec 覆蓋:README.tsme.md 4 個板塊、21 個端點、全部 25 個 Task 涵蓋(`BFIAUU_STOCK` 拆為薄封裝)。✅
- 型別一致:`Response` 內嵌 → `GetStat()` 由每個 DTO 自動提供 → `FetchJSON` 透過 interface 取得 stat;`Registry` 結構一致。✅
- 風險(已標明):
  - TWSE 反爬:30s timeout + 3x 指數退避由 `internal/httpx.Client` 既有實作提供;額外的 0.5-1.0s 隨機延遲可在 `runTwseEndpoint` 串接時加上(本計畫先不加,以免影響平行測試)。
  - TWSE 模組欄位漂移:`parseFloat` / `parsePercent` 已對千分位、`%`、`"`,但欄位順序變動仍需修 row mapping(對應個別 endpoint fetch 函式)。
  - TWSE 「無資料」訊息變化:用 `strings.Contains` 比對;若 Yahoo/TWSE 改為英文,會誤判為 OK——需在後續加上 i18n fallback。
```

---

## 執行注意

`★ Insight ─────────────────────────────────────`
- 25 個端點 Tasks(3-27)結構完全相同,**可以大量並行 dispatch**——每個 subagent 拿一個 endpoint,各自 commit 自己的端點,互不干擾(各寫各的檔)。這比序列化執行快 5-10 倍。
- 第一個 Task(Task 1)必須先完成(Task 2 的 registry 引用 `httpx.Client`;Task 3 開始每個端點引用 `FetchJSON`)。Task 2 可與 Task 1 並行(只 import `httpx`)。Task 3-27 全部依賴 Task 1。
- Task 26(CLI)依賴 Registry 內已 wire 好所有 Fetcher——所以是 Task 27 與最後一個端點 Task 同步完成後的最後一拍。
`─────────────────────────────────────────────────`
