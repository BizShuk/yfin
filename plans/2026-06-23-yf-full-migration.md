# Yahoo Finance 全資料維度遷移 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 讓 Go client (`yfinance-go`) 完整取代 Python `yf/scripts/*` 管線——補齊 13 個缺失指令、實作 crumb 認證、並提供批次抓取 + 分級快取編排。

**Architecture:** 以 Yahoo 免費的 `cookie + crumb` 認證解鎖 `v10/quoteSummary` 端點(yfinance 同款,非付費)。在 `internal/yahoo` 新增通用 quoteSummary fetcher,各資料維度解析為 `internal/norm` 的 `NormalizedXxx` 型別(Go 原生 normalized 契約)。最後在 `cmd/yfin` 新增 `batch` 子指令複刻 `all_ticker_yf.py` 的 ticker_list + 並行 + retry + 分級快取。

**Tech Stack:** Go 1.23+, cobra (CLI), net/http + net/http/cookiejar, 既有 `internal/httpx` / `internal/norm` / `internal/emit` 架構, testify。

**Refresh tiers (沿用 Python `config.py` REFRESH_MAP):** daily / monthly / quarterly / annually。

---

## 缺口對照(計畫覆蓋範圍)

| Python 指令 | 取得來源 | Go 現況 | 計畫 Task |
| --- | --- | --- | --- |
| major-holders / institutional-holders / mutualfund-holders | quoteSummary | ❌ | Task 5 |
| insider-transactions / insider-purchases / insider-roster | quoteSummary | ❌ | Task 6 |
| upgrades | quoteSummary `upgradeDowngradeHistory` | ❌ | Task 7 |
| calendar / earnings-dates | quoteSummary `calendarEvents` | ❌ | Task 8 |
| sec-filings | quoteSummary `secFilings` | ❌ | Task 9 |
| sustainability | quoteSummary `esgScores` | ❌ | Task 10 |
| recommendations / recommendations-summary | quoteSummary `recommendationTrend` | ⚠️ 部分 | Task 11 |
| info | quoteSummary 多模組合併 | ⚠️ 部分 | Task 12 |
| options | `/v7/finance/options/` | ❌ | Task 13 |
| isin | business-insider 查詢 | ❌ | Task 14 |
| actions | chart `events=div,split` | ⚠️ 部分 | Task 15 |
| metadata | chart meta | ⚠️ 部分 | Task 16 |
| (批次/快取/ticker_list) | — | ❌ | Task 17-20 |

> 既有已覆蓋(本計畫不重做):history、income/balance/cashflow、analysis 六項、price-targets、news、quote。

---

## File Structure

新增/修改檔案及其職責:

- Create `internal/yahoo/auth.go` — cookie jar + crumb 取得/快取/輪替(Yahoo 認證地基)
- Modify `internal/httpx/client.go` — 為底層 `http.Client` 掛上 `cookiejar.Jar`
- Create `internal/yahoo/quotesummary.go` — 通用 quoteSummary fetcher(`modules` 參數 + crumb 注入 + 401 重試)
- Create `internal/yahoo/rawvalue.go` — Yahoo `{raw, fmt, longFmt}` 數值解析輔助
- Create `internal/yahoo/holders.go` / `insider.go` / `upgrades.go` / `calendar.go` / `secfilings.go` / `esg.go` / `recommendations.go` / `info.go` — 各模組 DTO + decode
- Create `internal/yahoo/options.go` — `/v7/finance/options/` 端點
- Create `internal/yahoo/isin.go` — business-insider ISIN 查詢
- Modify `internal/yahoo/bars.go` — 從 chart events 抽出 actions(div/split)與 metadata
- Create `internal/norm/holders.go` 等 — 對應 `NormalizedXxx` 型別 + normalizer
- Create `internal/yahoo/normalize_*.go` — 各 DTO → Normalized 轉換
- Create `cmd/yfin/batch.go` — `batch` 子指令(ticker_list + 並行 + retry + 分級快取 + 輸出)
- Create `internal/cache/refresh.go` — `ShouldSkip` 分級快取(複刻 `should_skip`)
- Modify `cmd/yfin/main.go` — 註冊新子指令、`client.go` 暴露新 Fetch 方法
- Modify `client.go` — 新增 `FetchHolders` / `FetchInsider` / ... 對外方法

---

## Phase 0 — 認證地基 (crumb / cookie)

### Task 1: httpx.Client 掛載 cookie jar

**Files:**
- Modify: `internal/httpx/client.go:81-130`(`Client` struct 與建構處)
- Test: `internal/httpx/cookiejar_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/httpx/cookiejar_test.go
package httpx

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_PersistsCookiesAcrossRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/set" {
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "token123", Path: "/"})
			return
		}
		// /echo: reflect whether the cookie came back
		if c, err := r.Cookie("A1"); err == nil {
			_, _ = w.Write([]byte(c.Value))
		}
	}))
	defer srv.Close()

	c := NewClient(DefaultConfig())
	u, _ := url.Parse(srv.URL)

	req1, _ := http.NewRequest("GET", srv.URL+"/set", nil)
	resp1, err := c.Do(req1.Context(), req1)
	require.NoError(t, err)
	resp1.Body.Close()

	// jar must now hold the cookie for this host
	require.NotEmpty(t, c.Jar().Cookies(u))
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/httpx/ -run TestClient_PersistsCookiesAcrossRequests -v`
Expected: FAIL — `c.Jar undefined` 編譯錯誤。

- [ ] **Step 3: 最小實作**

在 `internal/httpx/client.go` 的 `Client` struct 新增欄位並於建構時初始化 jar:

```go
import "net/http/cookiejar"

// (Client struct 內新增)
type Client struct {
	httpClient *http.Client
	jar        *cookiejar.Jar
	// ... 既有欄位保留
}

// 在建立 *http.Client 之前/之後:
jar, _ := cookiejar.New(nil) // err 僅在 nil options 下不可能發生
httpClient := &http.Client{
	Jar: jar,
	// ... 既有設定保留
}
// 將 jar 存入 Client
c.jar = jar

// 新增存取器
func (c *Client) Jar() *cookiejar.Jar { return c.jar }
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/httpx/ -run TestClient_PersistsCookiesAcrossRequests -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/httpx/client.go internal/httpx/cookiejar_test.go
git commit -m "feat(httpx): add cookie jar to client for Yahoo crumb auth"
```

---

### Task 2: Crumb manager(取得 cookie + crumb,快取與 401 輪替)

Yahoo 流程:① `GET https://fc.yahoo.com/`(取得 A1/A3 cookie,403 也無妨)② `GET https://query2.finance.yahoo.com/v1/test/getcrumb`(帶 cookie → 回傳純文字 crumb)③ 後續請求附 `&crumb=<crumb>`。EU 需先過 consent。

**Files:**
- Create: `internal/yahoo/auth.go`
- Test: `internal/yahoo/auth_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/auth_test.go
package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bizshuk/yfinance-go/internal/httpx"
	"github.com/stretchr/testify/require"
)

func TestCrumbManager_FetchesAndCachesCrumb(t *testing.T) {
	var crumbCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/": // cookie endpoint
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "tok", Path: "/"})
		case "/v1/test/getcrumb":
			crumbCalls++
			_, _ = w.Write([]byte("abc123CRUMB"))
		}
	}))
	defer srv.Close()

	cm := NewCrumbManager(httpx.NewClient(httpx.DefaultConfig()), srv.URL, srv.URL)
	got, err := cm.Crumb(context.Background())
	require.NoError(t, err)
	require.Equal(t, "abc123CRUMB", got)

	// 第二次應走快取,不再打 getcrumb
	_, _ = cm.Crumb(context.Background())
	require.Equal(t, 1, crumbCalls)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestCrumbManager -v`
Expected: FAIL — `NewCrumbManager undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/auth.go
package yahoo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/bizshuk/yfinance-go/internal/httpx"
)

// CrumbManager handles Yahoo's cookie + crumb authentication.
type CrumbManager struct {
	httpClient *httpx.Client
	cookieURL  string // e.g. https://fc.yahoo.com
	apiBaseURL string // e.g. https://query2.finance.yahoo.com

	mu    sync.Mutex
	crumb string
}

func NewCrumbManager(httpClient *httpx.Client, cookieURL, apiBaseURL string) *CrumbManager {
	if cookieURL == "" {
		cookieURL = "https://fc.yahoo.com"
	}
	if apiBaseURL == "" {
		apiBaseURL = "https://query2.finance.yahoo.com"
	}
	return &CrumbManager{httpClient: httpClient, cookieURL: cookieURL, apiBaseURL: apiBaseURL}
}

// Crumb returns a cached crumb, fetching cookie+crumb on first use.
func (m *CrumbManager) Crumb(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.crumb != "" {
		return m.crumb, nil
	}
	if err := m.bootstrapCookie(ctx); err != nil {
		return "", err
	}
	crumb, err := m.fetchCrumb(ctx)
	if err != nil {
		return "", err
	}
	m.crumb = crumb
	return crumb, nil
}

// Invalidate clears the cached crumb (call on 401 to force re-fetch).
func (m *CrumbManager) Invalidate() {
	m.mu.Lock()
	m.crumb = ""
	m.mu.Unlock()
}

func (m *CrumbManager) bootstrapCookie(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", m.cookieURL+"/", nil)
	if err != nil {
		return err
	}
	resp, err := m.httpClient.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("cookie bootstrap failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // 403 亦可接受,只為取 Set-Cookie
	return nil
}

func (m *CrumbManager) fetchCrumb(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", m.apiBaseURL+"/v1/test/getcrumb", nil)
	if err != nil {
		return "", err
	}
	resp, err := m.httpClient.Do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("getcrumb failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	crumb := strings.TrimSpace(string(body))
	if crumb == "" || strings.Contains(crumb, "<html") {
		return "", fmt.Errorf("empty or invalid crumb (consent flow may be required)")
	}
	return crumb, nil
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestCrumbManager -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/auth.go internal/yahoo/auth_test.go
git commit -m "feat(yahoo): add crumb/cookie auth manager"
```

---

## Phase 1 — quoteSummary 基礎設施

### Task 3: 通用 quoteSummary fetcher

**Files:**
- Create: `internal/yahoo/quotesummary.go`
- Test: `internal/yahoo/quotesummary_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/quotesummary_test.go
package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bizshuk/yfinance-go/internal/httpx"
	"github.com/stretchr/testify/require"
)

func TestFetchQuoteSummary_InjectsCrumbAndModules(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			http.SetCookie(w, &http.Cookie{Name: "A1", Value: "tok", Path: "/"})
		case "/v1/test/getcrumb":
			_, _ = w.Write([]byte("CR"))
		default:
			gotURL = r.URL.String()
			_, _ = w.Write([]byte(`{"quoteSummary":{"result":[{}],"error":null}}`))
		}
	}))
	defer srv.Close()

	hc := httpx.NewClient(httpx.DefaultConfig())
	cm := NewCrumbManager(hc, srv.URL, srv.URL)
	c := NewClientWithAuth(hc, srv.URL, cm)

	raw, err := c.FetchQuoteSummary(context.Background(), "AAPL", []string{"esgScores", "secFilings"})
	require.NoError(t, err)
	require.Contains(t, string(raw), "quoteSummary")
	require.Contains(t, gotURL, "crumb=CR")
	require.True(t, strings.Contains(gotURL, "modules=esgScores%2CsecFilings") ||
		strings.Contains(gotURL, "modules=esgScores,secFilings"))
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestFetchQuoteSummary -v`
Expected: FAIL — `NewClientWithAuth` / `FetchQuoteSummary` undefined。

- [ ] **Step 3: 最小實作**

於 `internal/yahoo/client.go` 的 `Client` struct 增 `crumb *CrumbManager` 欄位,並新增建構子;`quotesummary.go` 新增 fetcher。

```go
// internal/yahoo/client.go 內新增建構子
func NewClientWithAuth(httpClient *httpx.Client, baseURL string, cm *CrumbManager) *Client {
	c := NewClient(httpClient, baseURL)
	c.crumb = cm
	return c
}
```

```go
// internal/yahoo/quotesummary.go
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
// transparently attaching the crumb and retrying once on 401.
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
```

注意:`Client` struct 需在 `client.go` 加上 `crumb *CrumbManager` 欄位。

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestFetchQuoteSummary -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/quotesummary.go internal/yahoo/quotesummary_test.go internal/yahoo/client.go
git commit -m "feat(yahoo): generic quoteSummary fetcher with crumb injection"
```

---

### Task 4: Yahoo `{raw, fmt}` 數值解析輔助

quoteSummary 數值多為 `{"raw": 123.4, "fmt": "123.4", "longFmt": "123"}` 形態,需共用解析。

**Files:**
- Create: `internal/yahoo/rawvalue.go`
- Test: `internal/yahoo/rawvalue_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/rawvalue_test.go
package yahoo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawValue_DecodesRawAndFmt(t *testing.T) {
	var v RawValue
	require.NoError(t, json.Unmarshal([]byte(`{"raw":123.45,"fmt":"123.45","longFmt":"123"}`), &v))
	require.NotNil(t, v.Raw)
	require.Equal(t, 123.45, *v.Raw)
	require.Equal(t, "123.45", v.Fmt)
}

func TestRawValue_HandlesEmptyObject(t *testing.T) {
	var v RawValue
	require.NoError(t, json.Unmarshal([]byte(`{}`), &v))
	require.Nil(t, v.Raw)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestRawValue -v`
Expected: FAIL — `RawValue undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/rawvalue.go
package yahoo

// RawValue models Yahoo's {raw, fmt, longFmt} value object.
type RawValue struct {
	Raw     *float64 `json:"raw,omitempty"`
	Fmt     string   `json:"fmt,omitempty"`
	LongFmt string   `json:"longFmt,omitempty"`
}

// RawInt is the integer-flavoured variant (timestamps, share counts).
type RawInt struct {
	Raw     *int64 `json:"raw,omitempty"`
	Fmt     string `json:"fmt,omitempty"`
	LongFmt string `json:"longFmt,omitempty"`
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestRawValue -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/rawvalue.go internal/yahoo/rawvalue_test.go
git commit -m "feat(yahoo): add RawValue/RawInt decoders for quoteSummary"
```

---

## Phase 2 — quoteSummary 資料維度

> 以下每個 Task 共用 Task 3 的 `FetchQuoteSummary` 與 Task 4 的 `RawValue`。每個維度新增:DTO(對應模組 JSON)→ `internal/norm` 的 `NormalizedXxx` 型別 → normalizer。CLI 串接統一在 Phase 5。

### Task 5: Holders(major / institutional / mutualfund)

**Modules:** `majorHoldersBreakdown`, `institutionOwnership`, `fundOwnership`

**Files:**
- Create: `internal/yahoo/holders.go`
- Create: `internal/norm/holders.go`
- Test: `internal/yahoo/holders_test.go`

- [ ] **Step 1: 寫失敗測試(用 testdata fixture)**

```go
// internal/yahoo/holders_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeHolders_ParsesBreakdownAndInstitutions(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "majorHoldersBreakdown":{"insidersPercentHeld":{"raw":0.012,"fmt":"1.20%"},
	    "institutionsPercentHeld":{"raw":0.74,"fmt":"74.00%"}},
	  "institutionOwnership":{"ownershipList":[
	    {"organization":"Vanguard","pctHeld":{"raw":0.08,"fmt":"8.00%"},
	     "position":{"raw":1000000},"value":{"raw":250000000}}]}
	}],"error":null}}`)

	d, err := DecodeHolders(raw)
	require.NoError(t, err)
	require.NotNil(t, d.MajorBreakdown.InstitutionsPercentHeld.Raw)
	require.Equal(t, 0.74, *d.MajorBreakdown.InstitutionsPercentHeld.Raw)
	require.Len(t, d.InstitutionOwnership, 1)
	require.Equal(t, "Vanguard", d.InstitutionOwnership[0].Organization)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeHolders -v`
Expected: FAIL — `DecodeHolders undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/holders.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type HoldersDTO struct {
	MajorBreakdown       MajorHoldersBreakdown `json:"majorHoldersBreakdown"`
	InstitutionOwnership []HolderRow           `json:"-"`
	FundOwnership        []HolderRow           `json:"-"`
}

type MajorHoldersBreakdown struct {
	InsidersPercentHeld     RawValue `json:"insidersPercentHeld"`
	InstitutionsPercentHeld RawValue `json:"institutionsPercentHeld"`
	InstitutionsFloatPctHeld RawValue `json:"institutionsFloatPercentHeld"`
	InstitutionsCount       RawInt   `json:"institutionsCount"`
}

type HolderRow struct {
	Organization string   `json:"organization"`
	PctHeld      RawValue `json:"pctHeld"`
	Position     RawInt   `json:"position"`
	Value        RawInt   `json:"value"`
}

// 中介結構(對映 quoteSummary 巢狀清單)
type holdersResult struct {
	QuoteSummary struct {
		Result []struct {
			MajorHoldersBreakdown MajorHoldersBreakdown `json:"majorHoldersBreakdown"`
			InstitutionOwnership  struct {
				OwnershipList []HolderRow `json:"ownershipList"`
			} `json:"institutionOwnership"`
			FundOwnership struct {
				OwnershipList []HolderRow `json:"ownershipList"`
			} `json:"fundOwnership"`
		} `json:"result"`
		Error *struct{ Description string } `json:"error"`
	} `json:"quoteSummary"`
}

func DecodeHolders(data []byte) (*HoldersDTO, error) {
	var r holdersResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("holders: empty result")
	}
	res := r.QuoteSummary.Result[0]
	return &HoldersDTO{
		MajorBreakdown:       res.MajorHoldersBreakdown,
		InstitutionOwnership: res.InstitutionOwnership.OwnershipList,
		FundOwnership:        res.FundOwnership.OwnershipList,
	}, nil
}

func (c *Client) FetchHolders(ctx context.Context, symbol string) (*HoldersDTO, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol,
		[]string{"majorHoldersBreakdown", "institutionOwnership", "fundOwnership"})
	if err != nil {
		return nil, err
	}
	return DecodeHolders(raw)
}
```

```go
// internal/norm/holders.go
package norm

import "time"

type NormalizedHolder struct {
	Organization string         `json:"organization"`
	PercentHeld  *ScaledDecimal `json:"percent_held,omitempty"`
	Position     *int64         `json:"position,omitempty"`
	Value        *int64         `json:"value,omitempty"`
}

type NormalizedHolders struct {
	Security                Security           `json:"security"`
	InsidersPercentHeld     *ScaledDecimal     `json:"insiders_percent_held,omitempty"`
	InstitutionsPercentHeld *ScaledDecimal     `json:"institutions_percent_held,omitempty"`
	Institutional           []NormalizedHolder `json:"institutional"`
	MutualFund              []NormalizedHolder `json:"mutual_fund"`
	AsOf                    time.Time          `json:"as_of"`
	Meta                    Meta               `json:"meta"`
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeHolders -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/holders.go internal/norm/holders.go internal/yahoo/holders_test.go
git commit -m "feat(yahoo): holders (major/institutional/mutualfund) via quoteSummary"
```

---

### Task 6: Insider(transactions / purchases / roster)

**Modules:** `insiderTransactions`, `netSharePurchaseActivity`, `insiderHolders`

**Files:**
- Create: `internal/yahoo/insider.go`
- Create: `internal/norm/insider.go`
- Test: `internal/yahoo/insider_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/insider_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeInsider_ParsesTransactions(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "insiderTransactions":{"transactions":[
	    {"filerName":"DOE JOHN","transactionText":"Sale at price 150.00",
	     "shares":{"raw":1000},"value":{"raw":150000},
	     "startDate":{"raw":1700000000}}]},
	  "netSharePurchaseActivity":{"period":"6m","buyInfoShares":{"raw":5000},
	     "sellInfoShares":{"raw":2000},"netInfoShares":{"raw":3000}},
	  "insiderHolders":{"holders":[
	    {"name":"DOE JANE","relation":"Director","positionDirect":{"raw":20000}}]}
	}],"error":null}}`)

	d, err := DecodeInsider(raw)
	require.NoError(t, err)
	require.Len(t, d.Transactions, 1)
	require.Equal(t, "DOE JOHN", d.Transactions[0].FilerName)
	require.NotNil(t, d.PurchaseActivity.NetInfoShares.Raw)
	require.Len(t, d.Roster, 1)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeInsider -v`
Expected: FAIL — `DecodeInsider undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/insider.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type InsiderDTO struct {
	Transactions     []InsiderTransaction
	PurchaseActivity NetSharePurchaseActivity
	Roster           []InsiderHolder
}

type InsiderTransaction struct {
	FilerName       string `json:"filerName"`
	TransactionText string `json:"transactionText"`
	Shares          RawInt `json:"shares"`
	Value           RawInt `json:"value"`
	StartDate       RawInt `json:"startDate"`
	OwnershipType   string `json:"ownership"`
}

type NetSharePurchaseActivity struct {
	Period         string `json:"period"`
	BuyInfoShares  RawInt `json:"buyInfoShares"`
	SellInfoShares RawInt `json:"sellInfoShares"`
	NetInfoShares  RawInt `json:"netInfoShares"`
}

type InsiderHolder struct {
	Name           string `json:"name"`
	Relation       string `json:"relation"`
	PositionDirect RawInt `json:"positionDirect"`
	LatestTransDate RawInt `json:"latestTransDate"`
}

type insiderResult struct {
	QuoteSummary struct {
		Result []struct {
			InsiderTransactions struct {
				Transactions []InsiderTransaction `json:"transactions"`
			} `json:"insiderTransactions"`
			NetSharePurchaseActivity NetSharePurchaseActivity `json:"netSharePurchaseActivity"`
			InsiderHolders           struct {
				Holders []InsiderHolder `json:"holders"`
			} `json:"insiderHolders"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeInsider(data []byte) (*InsiderDTO, error) {
	var r insiderResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("insider: empty result")
	}
	res := r.QuoteSummary.Result[0]
	return &InsiderDTO{
		Transactions:     res.InsiderTransactions.Transactions,
		PurchaseActivity: res.NetSharePurchaseActivity,
		Roster:           res.InsiderHolders.Holders,
	}, nil
}

func (c *Client) FetchInsider(ctx context.Context, symbol string) (*InsiderDTO, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol,
		[]string{"insiderTransactions", "netSharePurchaseActivity", "insiderHolders"})
	if err != nil {
		return nil, err
	}
	return DecodeInsider(raw)
}
```

```go
// internal/norm/insider.go
package norm

import "time"

type NormalizedInsiderTxn struct {
	FilerName string         `json:"filer_name"`
	Text      string         `json:"text"`
	Shares    *int64         `json:"shares,omitempty"`
	Value     *int64         `json:"value,omitempty"`
	Date      *time.Time     `json:"date,omitempty"`
}

type NormalizedInsider struct {
	Security     Security               `json:"security"`
	Transactions []NormalizedInsiderTxn `json:"transactions"`
	NetBuyShares *int64                 `json:"net_buy_shares,omitempty"`
	AsOf         time.Time              `json:"as_of"`
	Meta         Meta                   `json:"meta"`
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeInsider -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/insider.go internal/norm/insider.go internal/yahoo/insider_test.go
git commit -m "feat(yahoo): insider transactions/purchases/roster via quoteSummary"
```

---

### Task 7: Upgrades / Downgrades

**Module:** `upgradeDowngradeHistory`

**Files:**
- Create: `internal/yahoo/upgrades.go`
- Test: `internal/yahoo/upgrades_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/upgrades_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeUpgrades(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "upgradeDowngradeHistory":{"history":[
	    {"epochGradeDate":1700000000,"firm":"Morgan Stanley",
	     "toGrade":"Overweight","fromGrade":"Equal-Weight","action":"up"}]}
	}],"error":null}}`)

	rows, err := DecodeUpgrades(raw)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "Morgan Stanley", rows[0].Firm)
	require.Equal(t, "up", rows[0].Action)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeUpgrades -v`
Expected: FAIL — `DecodeUpgrades undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/upgrades.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type UpgradeRow struct {
	EpochGradeDate int64  `json:"epochGradeDate"`
	Firm           string `json:"firm"`
	ToGrade        string `json:"toGrade"`
	FromGrade      string `json:"fromGrade"`
	Action         string `json:"action"`
}

type upgradesResult struct {
	QuoteSummary struct {
		Result []struct {
			UpgradeDowngradeHistory struct {
				History []UpgradeRow `json:"history"`
			} `json:"upgradeDowngradeHistory"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeUpgrades(data []byte) ([]UpgradeRow, error) {
	var r upgradesResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("upgrades: empty result")
	}
	return r.QuoteSummary.Result[0].UpgradeDowngradeHistory.History, nil
}

func (c *Client) FetchUpgrades(ctx context.Context, symbol string) ([]UpgradeRow, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"upgradeDowngradeHistory"})
	if err != nil {
		return nil, err
	}
	return DecodeUpgrades(raw)
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeUpgrades -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/upgrades.go internal/yahoo/upgrades_test.go
git commit -m "feat(yahoo): upgrades/downgrades history via quoteSummary"
```

---

### Task 8: Calendar + Earnings dates

**Module:** `calendarEvents`(含 `earnings.earningsDate`、`dividendDate`、`exDividendDate`)

**Files:**
- Create: `internal/yahoo/calendar.go`
- Test: `internal/yahoo/calendar_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/calendar_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeCalendar(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "calendarEvents":{
	    "earnings":{"earningsDate":[{"raw":1701000000}],
	      "earningsAverage":{"raw":2.1},"revenueAverage":{"raw":100000000}},
	    "exDividendDate":{"raw":1699000000},
	    "dividendDate":{"raw":1699500000}}
	}],"error":null}}`)

	d, err := DecodeCalendar(raw)
	require.NoError(t, err)
	require.Len(t, d.EarningsDates, 1)
	require.Equal(t, int64(1701000000), d.EarningsDates[0])
	require.NotNil(t, d.ExDividendDate.Raw)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeCalendar -v`
Expected: FAIL — `DecodeCalendar undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/calendar.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type CalendarDTO struct {
	EarningsDates   []int64
	EarningsAverage RawValue
	RevenueAverage  RawValue
	ExDividendDate  RawInt
	DividendDate    RawInt
}

type calendarResult struct {
	QuoteSummary struct {
		Result []struct {
			CalendarEvents struct {
				Earnings struct {
					EarningsDate    []RawInt `json:"earningsDate"`
					EarningsAverage RawValue `json:"earningsAverage"`
					RevenueAverage  RawValue `json:"revenueAverage"`
				} `json:"earnings"`
				ExDividendDate RawInt `json:"exDividendDate"`
				DividendDate   RawInt `json:"dividendDate"`
			} `json:"calendarEvents"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeCalendar(data []byte) (*CalendarDTO, error) {
	var r calendarResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("calendar: empty result")
	}
	res := r.QuoteSummary.Result[0].CalendarEvents
	dto := &CalendarDTO{
		EarningsAverage: res.Earnings.EarningsAverage,
		RevenueAverage:  res.Earnings.RevenueAverage,
		ExDividendDate:  res.ExDividendDate,
		DividendDate:    res.DividendDate,
	}
	for _, e := range res.Earnings.EarningsDate {
		if e.Raw != nil {
			dto.EarningsDates = append(dto.EarningsDates, *e.Raw)
		}
	}
	return dto, nil
}

func (c *Client) FetchCalendar(ctx context.Context, symbol string) (*CalendarDTO, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"calendarEvents"})
	if err != nil {
		return nil, err
	}
	return DecodeCalendar(raw)
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeCalendar -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/calendar.go internal/yahoo/calendar_test.go
git commit -m "feat(yahoo): calendar + earnings dates via quoteSummary"
```

---

### Task 9: SEC filings

**Module:** `secFilings`

**Files:**
- Create: `internal/yahoo/secfilings.go`
- Test: `internal/yahoo/secfilings_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/secfilings_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeSecFilings(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "secFilings":{"filings":[
	    {"date":"2024-01-15","type":"10-K","title":"Annual Report",
	     "edgarUrl":"https://www.sec.gov/...","epochDate":1705276800}]}
	}],"error":null}}`)

	rows, err := DecodeSecFilings(raw)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "10-K", rows[0].Type)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeSecFilings -v`
Expected: FAIL — `DecodeSecFilings undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/secfilings.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type SecFiling struct {
	Date     string `json:"date"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	EdgarURL string `json:"edgarUrl"`
	EpochDate int64 `json:"epochDate"`
}

type secFilingsResult struct {
	QuoteSummary struct {
		Result []struct {
			SecFilings struct {
				Filings []SecFiling `json:"filings"`
			} `json:"secFilings"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeSecFilings(data []byte) ([]SecFiling, error) {
	var r secFilingsResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("secFilings: empty result")
	}
	return r.QuoteSummary.Result[0].SecFilings.Filings, nil
}

func (c *Client) FetchSecFilings(ctx context.Context, symbol string) ([]SecFiling, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"secFilings"})
	if err != nil {
		return nil, err
	}
	return DecodeSecFilings(raw)
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeSecFilings -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/secfilings.go internal/yahoo/secfilings_test.go
git commit -m "feat(yahoo): SEC filings via quoteSummary"
```

---

### Task 10: Sustainability (ESG)

**Module:** `esgScores`

**Files:**
- Create: `internal/yahoo/esg.go`
- Test: `internal/yahoo/esg_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/esg_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeESG(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "esgScores":{"totalEsg":{"raw":21.5},"environmentScore":{"raw":5.1},
	    "socialScore":{"raw":8.2},"governanceScore":{"raw":8.2},
	    "ratingYear":2024,"highestControversy":{"raw":3}}
	}],"error":null}}`)

	d, err := DecodeESG(raw)
	require.NoError(t, err)
	require.NotNil(t, d.TotalEsg.Raw)
	require.Equal(t, 21.5, *d.TotalEsg.Raw)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeESG -v`
Expected: FAIL — `DecodeESG undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/esg.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type ESGDTO struct {
	TotalEsg         RawValue `json:"totalEsg"`
	EnvironmentScore RawValue `json:"environmentScore"`
	SocialScore      RawValue `json:"socialScore"`
	GovernanceScore  RawValue `json:"governanceScore"`
	RatingYear       int      `json:"ratingYear"`
	HighestControversy RawValue `json:"highestControversy"`
}

type esgResult struct {
	QuoteSummary struct {
		Result []struct {
			ESGScores ESGDTO `json:"esgScores"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeESG(data []byte) (*ESGDTO, error) {
	var r esgResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("esg: empty result")
	}
	d := r.QuoteSummary.Result[0].ESGScores
	return &d, nil
}

func (c *Client) FetchESG(ctx context.Context, symbol string) (*ESGDTO, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"esgScores"})
	if err != nil {
		return nil, err
	}
	return DecodeESG(raw)
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeESG -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/esg.go internal/yahoo/esg_test.go
git commit -m "feat(yahoo): sustainability (ESG) via quoteSummary"
```

---

### Task 11: Recommendations 完整化

**Module:** `recommendationTrend`(補齊 Python `recommendations` / `recommendations-summary`)

**Files:**
- Create: `internal/yahoo/recommendations.go`
- Test: `internal/yahoo/recommendations_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/recommendations_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeRecommendationTrend(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "recommendationTrend":{"trend":[
	    {"period":"0m","strongBuy":5,"buy":10,"hold":3,"sell":1,"strongSell":0}]}
	}],"error":null}}`)

	rows, err := DecodeRecommendationTrend(raw)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, 5, rows[0].StrongBuy)
	require.Equal(t, "0m", rows[0].Period)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeRecommendationTrend -v`
Expected: FAIL — `DecodeRecommendationTrend undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/recommendations.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

type RecommendationTrendRow struct {
	Period     string `json:"period"`
	StrongBuy  int    `json:"strongBuy"`
	Buy        int    `json:"buy"`
	Hold       int    `json:"hold"`
	Sell       int    `json:"sell"`
	StrongSell int    `json:"strongSell"`
}

type recTrendResult struct {
	QuoteSummary struct {
		Result []struct {
			RecommendationTrend struct {
				Trend []RecommendationTrendRow `json:"trend"`
			} `json:"recommendationTrend"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func DecodeRecommendationTrend(data []byte) ([]RecommendationTrendRow, error) {
	var r recTrendResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("recommendationTrend: empty result")
	}
	return r.QuoteSummary.Result[0].RecommendationTrend.Trend, nil
}

func (c *Client) FetchRecommendationTrend(ctx context.Context, symbol string) ([]RecommendationTrendRow, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, []string{"recommendationTrend"})
	if err != nil {
		return nil, err
	}
	return DecodeRecommendationTrend(raw)
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeRecommendationTrend -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/recommendations.go internal/yahoo/recommendations_test.go
git commit -m "feat(yahoo): recommendation trend via quoteSummary"
```

---

### Task 12: Info 合併(多模組)

**Modules:** `assetProfile`, `summaryProfile`, `summaryDetail`, `defaultKeyStatistics`, `financialData`, `quoteType`(對齊 yfinance `.info` 的合併行為,輸出單一 map)

**Files:**
- Create: `internal/yahoo/info.go`
- Test: `internal/yahoo/info_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/info_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeInfo_MergesModules(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "assetProfile":{"sector":"Technology","industry":"Semiconductors",
	    "fullTimeEmployees":50000,"longBusinessSummary":"..."},
	  "summaryDetail":{"marketCap":{"raw":600000000000},
	    "trailingPE":{"raw":18.5},"dividendYield":{"raw":0.018}},
	  "quoteType":{"longName":"TSMC","symbol":"2330.TW","quoteType":"EQUITY"}
	}],"error":null}}`)

	info, err := DecodeInfo(raw)
	require.NoError(t, err)
	require.Equal(t, "Technology", info["sector"])
	require.Equal(t, "TSMC", info["longName"])
	require.InDelta(t, 6.0e11, info["marketCap"], 1)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeInfo -v`
Expected: FAIL — `DecodeInfo undefined`。

- [ ] **Step 3: 最小實作**

策略:對選定模組各自解為 `map[string]json.RawMessage`,標量直接展平,`{raw,...}` 物件取 `raw`,合併進單一 `map[string]any`。

```go
// internal/yahoo/info.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

var InfoModules = []string{
	"assetProfile", "summaryProfile", "summaryDetail",
	"defaultKeyStatistics", "financialData", "quoteType",
}

func DecodeInfo(data []byte) (map[string]any, error) {
	var r struct {
		QuoteSummary struct {
			Result []map[string]json.RawMessage `json:"result"`
		} `json:"quoteSummary"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("info: empty result")
	}
	out := map[string]any{}
	for _, modRaw := range r.QuoteSummary.Result[0] {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(modRaw, &fields); err != nil {
			continue
		}
		for k, v := range fields {
			out[k] = flattenValue(v)
		}
	}
	return out, nil
}

// flattenValue 將 {raw,...} 物件取 raw,其餘標量原樣回傳。
func flattenValue(v json.RawMessage) any {
	var obj struct {
		Raw json.RawMessage `json:"raw"`
	}
	if err := json.Unmarshal(v, &obj); err == nil && obj.Raw != nil {
		var raw any
		_ = json.Unmarshal(obj.Raw, &raw)
		return raw
	}
	var scalar any
	_ = json.Unmarshal(v, &scalar)
	return scalar
}

func (c *Client) FetchInfo(ctx context.Context, symbol string) (map[string]any, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, InfoModules)
	if err != nil {
		return nil, err
	}
	return DecodeInfo(raw)
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeInfo -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/info.go internal/yahoo/info_test.go
git commit -m "feat(yahoo): merged .info equivalent via quoteSummary modules"
```

---

## Phase 3 — 非 quoteSummary 維度

### Task 13: Options(`/v7/finance/options/`)

**Files:**
- Create: `internal/yahoo/options.go`
- Test: `internal/yahoo/options_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/options_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeOptions(t *testing.T) {
	raw := []byte(`{"optionChain":{"result":[{
	  "expirationDates":[1701000000,1701600000],
	  "strikes":[100.0,110.0],
	  "options":[{"expirationDate":1701000000,
	    "calls":[{"strike":{"raw":100},"lastPrice":{"raw":5.2},"volume":{"raw":120}}],
	    "puts":[{"strike":{"raw":100},"lastPrice":{"raw":2.1}}]}]
	}],"error":null}}`)

	d, err := DecodeOptions(raw)
	require.NoError(t, err)
	require.Len(t, d.ExpirationDates, 2)
	require.Len(t, d.Options[0].Calls, 1)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestDecodeOptions -v`
Expected: FAIL — `DecodeOptions undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/options.go
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OptionContract struct {
	Strike     RawValue `json:"strike"`
	LastPrice  RawValue `json:"lastPrice"`
	Bid        RawValue `json:"bid"`
	Ask        RawValue `json:"ask"`
	Volume     RawInt   `json:"volume"`
	OpenInterest RawInt `json:"openInterest"`
	ImpliedVolatility RawValue `json:"impliedVolatility"`
}

type OptionExpiry struct {
	ExpirationDate int64            `json:"expirationDate"`
	Calls          []OptionContract `json:"calls"`
	Puts           []OptionContract `json:"puts"`
}

type OptionsDTO struct {
	ExpirationDates []int64        `json:"expirationDates"`
	Strikes         []float64      `json:"strikes"`
	Options         []OptionExpiry `json:"options"`
}

type optionsResult struct {
	OptionChain struct {
		Result []OptionsDTO `json:"result"`
	} `json:"optionChain"`
}

func DecodeOptions(data []byte) (*OptionsDTO, error) {
	var r optionsResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.OptionChain.Result) == 0 {
		return nil, fmt.Errorf("options: empty result")
	}
	d := r.OptionChain.Result[0]
	return &d, nil
}

func (c *Client) FetchOptions(ctx context.Context, symbol string) (*OptionsDTO, error) {
	url := c.baseURL + "/v7/finance/options/" + symbol
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return DecodeOptions(body)
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestDecodeOptions -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/options.go internal/yahoo/options_test.go
git commit -m "feat(yahoo): options chain via v7 endpoint"
```

---

### Task 14: ISIN(business-insider 查詢)

> 注意:ISIN 在 yfinance 不走 Yahoo,而是查 `markets.businessinsider.com/ajax/SearchController_Suggest`。僅對含 `-` 的非美股有效。屬低頻 annually,容許失敗。

**Files:**
- Create: `internal/yahoo/isin.go`
- Test: `internal/yahoo/isin_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/isin_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseISINResponse(t *testing.T) {
	// business-insider 回傳格式: "0\tAAPL|US0378331005|Apple Inc"
	body := "0\tAAPL|US0378331005|Apple Inc.\n1\t..."
	isin, err := parseISIN(body, "AAPL")
	require.NoError(t, err)
	require.Equal(t, "US0378331005", isin)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestParseISINResponse -v`
Expected: FAIL — `parseISIN undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/isin.go
package yahoo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const isinSearchURL = "https://markets.businessinsider.com/ajax/SearchController_Suggest"

func parseISIN(body, ticker string) (string, error) {
	for _, line := range strings.Split(body, "\n") {
		// 欄位以 | 分隔:SYMBOL|ISIN|NAME
		parts := strings.Split(line, "|")
		if len(parts) >= 2 && strings.Contains(strings.ToUpper(parts[0]), strings.ToUpper(ticker)) {
			return strings.TrimSpace(parts[1]), nil
		}
	}
	return "", fmt.Errorf("isin not found for %s", ticker)
}

func (c *Client) FetchISIN(ctx context.Context, symbol string) (string, error) {
	// 去掉交易所後綴 (2330.TW -> 2330)
	q := symbol
	if i := strings.Index(symbol, "."); i > 0 {
		q = symbol[:i]
	}
	u := isinSearchURL + "?max_results=25&query=" + url.QueryEscape(q)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return parseISIN(string(body), q)
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestParseISINResponse -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/isin.go internal/yahoo/isin_test.go
git commit -m "feat(yahoo): ISIN lookup via business-insider"
```

---

### Task 15: Actions(dividends + splits)從 chart events 抽出

`buildBarsURL` 已帶 `events=div,split`。chart 回應內 `result[0].events.{dividends,splits}` 需解析為獨立輸出。

**Files:**
- Modify: `internal/yahoo/bars.go`(`BarsResponse` 加 `Events`)
- Create: `internal/yahoo/actions.go`
- Test: `internal/yahoo/actions_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/actions_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractActions(t *testing.T) {
	raw := []byte(`{"chart":{"result":[{
	  "meta":{"symbol":"AAPL"},
	  "events":{
	    "dividends":{"1700000000":{"amount":0.24,"date":1700000000}},
	    "splits":{"1600000000":{"numerator":4,"denominator":1,"splitRatio":"4:1","date":1600000000}}}
	}],"error":null}}`)

	acts, err := ExtractActions(raw)
	require.NoError(t, err)
	require.Len(t, acts.Dividends, 1)
	require.Equal(t, 0.24, acts.Dividends[0].Amount)
	require.Len(t, acts.Splits, 1)
	require.Equal(t, "4:1", acts.Splits[0].SplitRatio)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestExtractActions -v`
Expected: FAIL — `ExtractActions undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/actions.go
package yahoo

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Dividend struct {
	Date   int64   `json:"date"`
	Amount float64 `json:"amount"`
}

type Split struct {
	Date        int64  `json:"date"`
	Numerator   int    `json:"numerator"`
	Denominator int    `json:"denominator"`
	SplitRatio  string `json:"splitRatio"`
}

type ActionsDTO struct {
	Dividends []Dividend
	Splits    []Split
}

type actionsResult struct {
	Chart struct {
		Result []struct {
			Events struct {
				Dividends map[string]Dividend `json:"dividends"`
				Splits    map[string]Split    `json:"splits"`
			} `json:"events"`
		} `json:"result"`
	} `json:"chart"`
}

func ExtractActions(data []byte) (*ActionsDTO, error) {
	var r actionsResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.Chart.Result) == 0 {
		return nil, fmt.Errorf("actions: empty result")
	}
	ev := r.Chart.Result[0].Events
	out := &ActionsDTO{}
	for _, d := range ev.Dividends {
		out.Dividends = append(out.Dividends, d)
	}
	for _, s := range ev.Splits {
		out.Splits = append(out.Splits, s)
	}
	sort.Slice(out.Dividends, func(i, j int) bool { return out.Dividends[i].Date < out.Dividends[j].Date })
	sort.Slice(out.Splits, func(i, j int) bool { return out.Splits[i].Date < out.Splits[j].Date })
	return out, nil
}
```

`FetchActions(ctx, symbol)`:呼叫既有 `buildBarsURL`(預設 1y range)取得 raw chart bytes 後 `ExtractActions`。在 `bars.go` 已有 fetch 邏輯,新增一個回傳原始 bytes 的私有 helper 供此處重用。

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestExtractActions -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/actions.go internal/yahoo/bars.go internal/yahoo/actions_test.go
git commit -m "feat(yahoo): extract dividends/splits actions from chart events"
```

---

### Task 16: Metadata(chart meta 獨立輸出)

`BarsResponse.GetMetadata()` 已存在,僅需暴露為對外 `FetchMetadata`。

**Files:**
- Create: `internal/yahoo/metadata.go`
- Test: `internal/yahoo/metadata_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/yahoo/metadata_test.go
package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractMetadata(t *testing.T) {
	raw := []byte(`{"chart":{"result":[{
	  "meta":{"symbol":"AAPL","currency":"USD","exchangeName":"NMS",
	    "instrumentType":"EQUITY","timezone":"EST","gmtoffset":-18000}
	}],"error":null}}`)

	m, err := ExtractMetadata(raw)
	require.NoError(t, err)
	require.Equal(t, "AAPL", m.Symbol)
	require.Equal(t, "USD", m.Currency)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/yahoo/ -run TestExtractMetadata -v`
Expected: FAIL — `ExtractMetadata undefined`。

- [ ] **Step 3: 最小實作**

```go
// internal/yahoo/metadata.go
package yahoo

import (
	"encoding/json"
	"fmt"
)

type ChartMetadata struct {
	Symbol         string  `json:"symbol"`
	Currency       string  `json:"currency"`
	ExchangeName   string  `json:"exchangeName"`
	InstrumentType string  `json:"instrumentType"`
	Timezone       string  `json:"timezone"`
	GmtOffset      int     `json:"gmtoffset"`
	FirstTradeDate int64   `json:"firstTradeDate"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}

type metaResult struct {
	Chart struct {
		Result []struct {
			Meta ChartMetadata `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}

func ExtractMetadata(data []byte) (*ChartMetadata, error) {
	var r metaResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.Chart.Result) == 0 {
		return nil, fmt.Errorf("metadata: empty result")
	}
	m := r.Chart.Result[0].Meta
	return &m, nil
}
```

`FetchMetadata` 重用 Task 15 的私有 raw-chart helper。

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/yahoo/ -run TestExtractMetadata -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/yahoo/metadata.go internal/yahoo/metadata_test.go
git commit -m "feat(yahoo): expose chart metadata extraction"
```

---

## Phase 4 — 批次編排 + 分級快取

### Task 17: ticker_list CSV 讀取

複刻 `get_ticker_list`:跳過首行,取每行最後一個逗號欄位。

**Files:**
- Create: `internal/cache/tickerlist.go`
- Test: `internal/cache/tickerlist_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/cache/tickerlist_test.go
package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadTickerList(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ticker_list.csv")
	require.NoError(t, os.WriteFile(p, []byte("market, ticker\nTPEx, 3081.TWO\nTWSE, 2330.TW\n"), 0o644))

	got, err := ReadTickerList(p)
	require.NoError(t, err)
	require.Equal(t, []string{"3081.TWO", "2330.TW"}, got)
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/cache/ -run TestReadTickerList -v`
Expected: FAIL — package/func 不存在。

- [ ] **Step 3: 最小實作**

```go
// internal/cache/tickerlist.go
package cache

import (
	"bufio"
	"os"
	"strings"
)

func ReadTickerList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tickers []string
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first { // skip header
			first = false
			continue
		}
		parts := strings.Split(strings.TrimSpace(sc.Text()), ",")
		if len(parts) == 0 {
			continue
		}
		t := strings.TrimSpace(parts[len(parts)-1])
		if t != "" {
			tickers = append(tickers, t)
		}
	}
	return tickers, sc.Err()
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/cache/ -run TestReadTickerList -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cache/tickerlist.go internal/cache/tickerlist_test.go
git commit -m "feat(cache): ticker_list CSV reader"
```

---

### Task 18: 分級快取 ShouldSkip(複刻 `should_skip`)

**Files:**
- Create: `internal/cache/refresh.go`
- Test: `internal/cache/refresh_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// internal/cache/refresh_test.go
package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShouldSkip_DailyTierSkipsSameDay(t *testing.T) {
	root := t.TempDir()
	today := time.Now()
	cmdDir := filepath.Join(root, "history")
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	fn := filepath.Join(cmdDir, "AAPL."+today.Format("2006-01-02")+".json")
	require.NoError(t, os.WriteFile(fn, []byte("{}"), 0o644))

	// daily tier: 同日已抓 → skip
	require.True(t, ShouldSkip("history", "AAPL", false, root, today))
	// force → 不 skip
	require.False(t, ShouldSkip("history", "AAPL", true, root, today))
}

func TestShouldSkip_QuarterlyTier(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 6, 23, 0, 0, 0, 0, time.UTC)
	cmdDir := filepath.Join(root, "income")
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	// 同一季 (Q2: Apr-Jun) 已抓 → skip
	fn := filepath.Join(cmdDir, "AAPL.2026-04-10.json")
	require.NoError(t, os.WriteFile(fn, []byte("{}"), 0o644))
	require.True(t, ShouldSkip("income", "AAPL", false, root, now))
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./internal/cache/ -run TestShouldSkip -v`
Expected: FAIL — `ShouldSkip` / `RefreshMap` 不存在。

- [ ] **Step 3: 最小實作**

```go
// internal/cache/refresh.go
package cache

import (
	"path/filepath"
	"time"
)

// RefreshMap 對齊 Python config.py 的 REFRESH_MAP。
var RefreshMap = map[string]string{
	"history": "daily", "recommendations": "daily", "recommendations-summary": "daily",
	"upgrades": "daily", "news": "daily", "metadata": "daily",
	"info": "monthly", "insider-transactions": "monthly", "insider-purchases": "monthly",
	"insider-roster": "monthly", "calendar": "monthly",
	"actions": "quarterly", "income": "quarterly", "balance": "quarterly", "cashflow": "quarterly",
	"major-holders": "quarterly", "institutional-holders": "quarterly", "mutualfund-holders": "quarterly",
	"earnings-dates": "quarterly", "earnings-history": "quarterly", "eps-trend": "quarterly",
	"eps-revisions": "quarterly", "earnings-estimates": "quarterly", "revenue-estimates": "quarterly",
	"growth-estimates": "quarterly", "price-targets": "quarterly", "sec-filings": "quarterly",
	"sustainability": "quarterly",
	"isin": "annually", "options": "annually",
}

func quarter(m time.Month) int { return (int(m) - 1) / 3 + 1 }

// ShouldSkip 依分級快取判斷是否跳過抓取。rawDir 為 raw 根目錄。
func ShouldSkip(command, ticker string, force bool, rawDir string, now time.Time) bool {
	if force {
		return false
	}
	tier := RefreshMap[command]
	if tier == "" {
		tier = "daily"
	}
	matches, _ := filepath.Glob(filepath.Join(rawDir, command, ticker+".*.json"))
	for _, f := range matches {
		base := filepath.Base(f)
		// {ticker}.{YYYY-MM-DD}.json
		stem := base[len(ticker)+1 : len(base)-len(".json")]
		fd, err := time.Parse("2006-01-02", stem)
		if err != nil {
			continue
		}
		switch tier {
		case "daily":
			if fd.YearDay() == now.YearDay() && fd.Year() == now.Year() {
				return true
			}
		case "weekly":
			if now.Sub(fd).Hours() < 7*24 {
				return true
			}
		case "monthly":
			if fd.Year() == now.Year() && fd.Month() == now.Month() {
				return true
			}
		case "quarterly":
			if fd.Year() == now.Year() && quarter(fd.Month()) == quarter(now.Month()) {
				return true
			}
		case "annually":
			if fd.Year() == now.Year() {
				return true
			}
		}
	}
	return false
}
```

- [ ] **Step 4: 執行確認通過**

Run: `go test ./internal/cache/ -run TestShouldSkip -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cache/refresh.go internal/cache/refresh_test.go
git commit -m "feat(cache): tiered refresh ShouldSkip (daily/monthly/quarterly/annually)"
```

---

### Task 19: 指令分派表(command → Fetch 方法)

把 30 指令對映到對應的 client Fetch 函式,回傳可序列化結果。

**Files:**
- Create: `cmd/yfin/dispatch.go`
- Test: `cmd/yfin/dispatch_test.go`

- [ ] **Step 1: 寫失敗測試**

```go
// cmd/yfin/dispatch_test.go
package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandRegistry_CoversAllCommands(t *testing.T) {
	// 對齊 Python config.py 的 30 指令
	want := []string{
		"info", "history", "actions", "income", "balance", "cashflow",
		"major-holders", "institutional-holders", "mutualfund-holders",
		"insider-transactions", "insider-purchases", "insider-roster",
		"recommendations", "recommendations-summary", "upgrades",
		"earnings-dates", "earnings-history", "eps-trend", "eps-revisions",
		"earnings-estimates", "revenue-estimates", "growth-estimates",
		"price-targets", "news", "calendar", "sec-filings", "sustainability",
		"isin", "options", "metadata",
	}
	for _, cmd := range want {
		_, ok := commandRegistry[cmd]
		require.Truef(t, ok, "command %q missing from registry", cmd)
	}
	require.Len(t, commandRegistry, len(want))
}
```

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./cmd/yfin/ -run TestCommandRegistry -v`
Expected: FAIL — `commandRegistry undefined`。

- [ ] **Step 3: 最小實作**

```go
// cmd/yfin/dispatch.go
package main

import (
	"context"

	"github.com/bizshuk/yfinance-go/svc/yahoo"
)

// fetchFunc 取得單一指令的資料並回傳可 JSON 序列化的值。
type fetchFunc func(ctx context.Context, c *yahoo.Client, symbol string) (any, error)

var commandRegistry = map[string]fetchFunc{
	"info":    func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchInfo(ctx, s) },
	"actions": func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchActions(ctx, s) },
	"metadata": func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchMetadata(ctx, s) },
	"major-holders":         holdersFetch, // 三者共用 FetchHolders,輸出時取對應切片
	"institutional-holders": holdersFetch,
	"mutualfund-holders":    holdersFetch,
	"insider-transactions": func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchInsider(ctx, s) },
	"insider-purchases":    func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchInsider(ctx, s) },
	"insider-roster":       func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchInsider(ctx, s) },
	"upgrades":       func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchUpgrades(ctx, s) },
	"calendar":       func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchCalendar(ctx, s) },
	"earnings-dates": func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchCalendar(ctx, s) },
	"sec-filings":    func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchSecFilings(ctx, s) },
	"sustainability": func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchESG(ctx, s) },
	"recommendations":         func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchRecommendationTrend(ctx, s) },
	"recommendations-summary": func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchRecommendationTrend(ctx, s) },
	"options": func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchOptions(ctx, s) },
	"isin":    func(ctx context.Context, c *yahoo.Client, s string) (any, error) { return c.FetchISIN(ctx, s) },
	// 既有能力(走現有 scrape/yahoo 管線,以 adapter 包裝)
	"history":            historyFetch,
	"income":             financialsFetch("income"),
	"balance":            financialsFetch("balance"),
	"cashflow":           financialsFetch("cashflow"),
	"earnings-history":   analysisFetch,
	"eps-trend":          analysisFetch,
	"eps-revisions":      analysisFetch,
	"earnings-estimates": analysisFetch,
	"revenue-estimates":  analysisFetch,
	"growth-estimates":   analysisFetch,
	"price-targets":      analystInsightsFetch,
	"news":               newsFetch,
}

func holdersFetch(ctx context.Context, c *yahoo.Client, s string) (any, error) {
	return c.FetchHolders(ctx, s)
}
```

> `historyFetch` / `financialsFetch` / `analysisFetch` / `analystInsightsFetch` / `newsFetch` 為包裝既有 `client.go` 方法的 adapter(在本 Task 一併實作,各自呼叫對應既有 `Scrape*` / `Fetch*` 並回傳結果物件)。

- [ ] **Step 4: 執行確認通過**

Run: `go test ./cmd/yfin/ -run TestCommandRegistry -v`
Expected: PASS(30 指令全覆蓋)

- [ ] **Step 5: Commit**

```bash
git add cmd/yfin/dispatch.go cmd/yfin/dispatch_test.go
git commit -m "feat(cli): command registry mapping 30 commands to fetchers"
```

---

### Task 20: `batch` 子指令(並行 + retry + 快取 + 輸出)

複刻 `all_ticker_yf.py`:`--ticker` / `--max-workers`(預設 10)/ `--force`,輸出 `<rawDir>/<command>/<ticker>.<YYYY-MM-DD>.json`,失敗寫 `<rawDir>/_failed/<ticker>.<command>.err`。

**Files:**
- Create: `cmd/yfin/batch.go`
- Modify: `cmd/yfin/main.go:373-381`(`rootCmd.AddCommand(batchCmd)`)
- Test: `cmd/yfin/batch_test.go`

- [ ] **Step 1: 寫失敗測試(快取跳過 + 輸出路徑)**

```go
// cmd/yfin/batch_test.go
package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunBatchForTicker_WritesOutputAndRespectsCache(t *testing.T) {
	root := t.TempDir()
	now := time.Now()

	// stub fetcher:回傳固定 JSON
	reg := map[string]fetchFunc{
		"info": func(ctx context.Context, c *clientStub, s string) (any, error) {
			return map[string]string{"symbol": s}, nil
		},
	}
	_ = reg // 介面對齊見實作

	res := runBatchForTicker(context.Background(), nil, "AAPL", []string{"info"}, false, root, now)
	require.Equal(t, "success", res.Commands["info"])

	out := filepath.Join(root, "info", "AAPL."+now.Format("2006-01-02")+".json")
	_, err := os.Stat(out)
	require.NoError(t, err)

	// 第二次應 skip(monthly tier, 同月)
	res2 := runBatchForTicker(context.Background(), nil, "AAPL", []string{"info"}, false, root, now)
	require.Equal(t, "skipped", res2.Commands["info"])
}
```

> 註:測試需要 `commandRegistry` 可被注入 stub。實作時將 registry 與 client 作為 `runBatchForTicker` 參數傳入(依賴注入),測試傳 stub;`clientStub` 為測試輔助型別。若注入成本過高,改以一個真實 `*yahoo.Client` + httptest 伺服器替代——擇一,但須在本 Task 完成可執行測試。

- [ ] **Step 2: 執行確認失敗**

Run: `go test ./cmd/yfin/ -run TestRunBatchForTicker -v`
Expected: FAIL — `runBatchForTicker undefined`。

- [ ] **Step 3: 最小實作**

```go
// cmd/yfin/batch.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bizshuk/yfinance-go/internal/cache"
	"github.com/bizshuk/yfinance-go/svc/yahoo"
	"github.com/spf13/cobra"
)

const batchRetries = 3

type tickerResult struct {
	Ticker   string
	Commands map[string]string // command -> success|skipped|failed|not_found
}

func runBatchForTicker(ctx context.Context, c *yahoo.Client, ticker string,
	commands []string, force bool, rawDir string, now time.Time) tickerResult {

	res := tickerResult{Ticker: ticker, Commands: map[string]string{}}
	for _, command := range commands {
		if cache.ShouldSkip(command, ticker, force, rawDir, now) {
			res.Commands[command] = "skipped"
			continue
		}
		fn, ok := commandRegistry[command]
		if !ok {
			res.Commands[command] = "failed"
			continue
		}

		var lastErr error
		var data any
		for attempt := 0; attempt < batchRetries; attempt++ {
			data, lastErr = fn(ctx, c, ticker)
			if lastErr == nil {
				break
			}
			if attempt < batchRetries-1 {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
			}
		}
		if lastErr != nil {
			errPath := filepath.Join(rawDir, "_failed", fmt.Sprintf("%s.%s.err", ticker, command))
			_ = os.MkdirAll(filepath.Dir(errPath), 0o755)
			_ = os.WriteFile(errPath, []byte(lastErr.Error()), 0o644)
			res.Commands[command] = "failed"
			continue
		}

		outPath := filepath.Join(rawDir, command, fmt.Sprintf("%s.%s.json", ticker, now.Format("2006-01-02")))
		_ = os.MkdirAll(filepath.Dir(outPath), 0o755)
		b, _ := json.MarshalIndent(data, "", "  ")
		_ = os.WriteFile(outPath, b, 0o644)
		res.Commands[command] = "success"
	}
	return res
}

var (
	batchTicker     string
	batchMaxWorkers int
	batchForce      bool
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Batch-fetch all commands for a ticker universe (yf/scripts parity)",
	RunE:  runBatch,
}

func init() {
	batchCmd.Flags().StringVar(&batchTicker, "ticker", "", "Single ticker (default: ticker_list.csv)")
	batchCmd.Flags().IntVar(&batchMaxWorkers, "max-workers", 10, "Max concurrent workers")
	batchCmd.Flags().BoolVar(&batchForce, "force", false, "Force re-fetch, ignore cache")
}

func runBatch(cmd *cobra.Command, args []string) error {
	rawDir := filepath.Join(os.Getenv("HOME"), ".config", "stock", "data", "raw")
	now := time.Now()

	var tickers []string
	if batchTicker != "" {
		tickers = []string{batchTicker}
	} else {
		var err error
		tickers, err = cache.ReadTickerList(
			filepath.Join("yf", "references", "ticker_list.csv"))
		if err != nil {
			return err
		}
	}

	allCommands := make([]string, 0, len(commandRegistry))
	for k := range commandRegistry {
		allCommands = append(allCommands, k)
	}

	c := buildAuthedClient() // 建立帶 crumb 的 *yahoo.Client(見下)
	ctx := context.Background()

	sem := make(chan struct{}, batchMaxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var success, skipped, failed int

	for _, t := range tickers {
		wg.Add(1)
		sem <- struct{}{}
		go func(tk string) {
			defer wg.Done()
			defer func() { <-sem }()
			r := runBatchForTicker(ctx, c, tk, allCommands, batchForce, rawDir, now)
			mu.Lock()
			for _, st := range r.Commands {
				switch st {
				case "success":
					success++
				case "skipped":
					skipped++
				case "failed":
					failed++
				}
			}
			mu.Unlock()
			fmt.Printf("  %s: %d commands processed\n", tk, len(r.Commands))
		}(t)
	}
	wg.Wait()
	fmt.Printf("Done. success=%d skipped=%d failed=%d\n", success, skipped, failed)
	return nil
}

// buildAuthedClient 組裝帶 cookie jar + crumb 的 yahoo.Client。
func buildAuthedClient() *yahoo.Client {
	// 使用既有 httpx 設定建立 client;baseURL 用 query1。
	// hc := httpx.NewClient(httpx.DefaultConfig())
	// cm := yahoo.NewCrumbManager(hc, "", "")
	// return yahoo.NewClientWithAuth(hc, "https://query1.finance.yahoo.com", cm)
	panic("wire with project's httpx config in implementation")
}
```

> 實作時 `buildAuthedClient` 須以專案實際 `httpx` 設定(參考 `cmd/yfin/main.go` 內 `createScrapeClient` 的設定流程)組裝,移除 `panic`。

- [ ] **Step 4: 執行確認通過 + 註冊指令**

在 `cmd/yfin/main.go` 的 `rootCmd.AddCommand(...)` 區塊新增:
```go
rootCmd.AddCommand(batchCmd)
```
Run: `go test ./cmd/yfin/ -run TestRunBatchForTicker -v && go build ./cmd/yfin`
Expected: PASS + 編譯成功

- [ ] **Step 5: Commit**

```bash
git add cmd/yfin/batch.go cmd/yfin/batch_test.go cmd/yfin/main.go
git commit -m "feat(cli): batch subcommand with concurrency, retry, tiered cache"
```

---

## Phase 5 — 整合驗證與文件

### Task 21: 端到端整合測試 + 文件更新

**Files:**
- Create: `tests/batch_integration_test.go`
- Modify: `README.md`(新增 `batch` 指令說明)、`yf/SKILL.md`(標注 Go 已對等)

- [ ] **Step 1: 整合測試(httptest 模擬 Yahoo,跑 1 ticker × 數指令)**

```go
// tests/batch_integration_test.go — 用 httptest 模擬 cookie/getcrumb/quoteSummary/options
// 驗證:① crumb 注入 ② 各 DecodeXxx 串通 ③ 輸出 JSON 落地至 tmp rawDir
// (完整 server handler 比照 Task 3 測試,擴充各 module 路徑回傳 fixture)
```

- [ ] **Step 2: 執行全測試套件**

Run: `go test ./... && go build ./cmd/yfin`
Expected: 全 PASS、編譯成功

- [ ] **Step 3: 真實單股冒煙測試(手動)**

Run: `go run ./cmd/yfin batch --ticker AAPL --max-workers 2`
Expected: `~/.config/stock/data/raw/<command>/AAPL.<today>.json` 產生多個檔案;`info`/`holders`/`calendar` 等非空。

- [ ] **Step 4: 更新文件**

- `README.md`:新增 `### batch` 段,說明 30 指令、輸出路徑、flags。
- `yf/SKILL.md`:在維度表後加註「Go client (`yfin batch`) 已達 Python 對等,建議改用 Go」。

- [ ] **Step 5: Commit**

```bash
git add tests/batch_integration_test.go README.md yf/SKILL.md
git commit -m "test+docs: batch e2e integration and parity documentation"
```

---

## Self-Review Checklist

- Spec 覆蓋:13 缺失指令(Task 5-14)+ 2 部分(Task 15-16)+ 編排層(Task 17-20)+ 驗證(Task 21)= 全部 30 指令對齊。✅
- 型別一致:`RawValue`/`RawInt`(Task 4)貫穿 Task 5-13;`commandRegistry`(Task 19)被 Task 20 使用;`ShouldSkip`(Task 18)簽名與 Task 20 呼叫一致。✅
- 未決風險(實作時須處理,非 placeholder):
  - `buildAuthedClient` 需接專案 httpx 設定(Task 20 已標明)。
  - EU consent flow:若 `getcrumb` 回 HTML,Task 2 已偵測並報錯;遇到時需補 consent cookie 流程。
  - quoteSummary 模組欄位可能隨 Yahoo 改版漂移——各 `DecodeXxx` 採寬鬆解析(缺欄位回 nil 而非 error)。
```
