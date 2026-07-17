# Yahoo Circuit Groups

## 結論

在現有 per-authority circuit breaker 上增加由 request context 顯式指定的
`circuit group`。Breaker identity 從單一 authority 改為 `{authority, group}`；Yahoo service
依 endpoint family 標記 group，其他未標記的 HTTP caller 繼續使用既有 host-level breaker。

這項變更只修復同一 Yahoo authority 內的跨 family 污染：`getcrumb`／`quoteSummary` 的
`429` 可以打開 `yahoo-auth`，但不得阻擋健康的 chart、fundamentals-timeseries、news 或
options request。

## 背景與證據

正式設定 `minimum_requests: 10` 執行 AAPL 30-command batch 時，四個新 endpoint 都成功，
但最終只有 `7 success / 23 failed`。原因是 query2 上的多個 `getcrumb 429` 打開 host breaker，
讓稍後的 `metadata` chart request 回傳 `circuit breaker is open`。

將 minimum request 暫時提高到 `1000` 後，相同程式得到 `8 success / 22 failed`；`metadata`
恢復成功，實際剩餘 root failure 為 crumb 429、HTML 503、options 401 與 earnings-dates 空資料。
設定已還原為正式值 `10`。

因此 breaker 的 rolling window、failure-rate threshold 與 logical-request outcome 都能正常運作；
剩餘問題是 isolation key 粒度太粗，而不是閾值本身錯誤。

## 範圍

本次包含：

- 新增 context-based circuit group API。
- Breaker registry 以 authority + group 建立獨立 breaker。
- Yahoo endpoint family 顯式標記。
- Yahoo HTML scrape request 標記。
- 保留未標記 caller 的 host-level 行為。
- unit、race、architecture 與 live batch 驗證。

本次不包含：

- 修改 failure rate、minimum requests、rolling window 或 reset timeout。
- 修改 `429`／`5xx`／network error 的 breaker outcome classification。
- 修改 retry、backoff、rate limiter 或 cookie jar。
- 修復 cookie、crumb、browser fingerprint 或剩餘 22 個 endpoint failure。
- 依 symbol、quoteSummary module 或完整 URL 建立 breaker，避免無界 cardinality。
- 將 Yahoo path pattern 寫入 `utils/httpx`。

## API 設計

`utils/httpx` 新增：

```go
func WithCircuitGroup(ctx context.Context, group string) context.Context
```

規則：

- `group` 會 `TrimSpace` 並轉成 lowercase。
- 空 group 不建立新的 context value，等同既有 host-level breaker。
- group 存在未 export context key 中，不會寫入 HTTP header 或傳送至 upstream。
- `Client.Do` 從 `req.Context()` 取得 group。Request context 是 group 的單一真相來源。
- Service 建立 grouped request 時，同一個 grouped context 同時傳給
  `http.NewRequestWithContext` 與 `Client.Do`，避免 context 分歧。

不新增 YAML 或 `httpx.Config` 選項。Group 是 service 的固定架構契約，不是 runtime tuning knob。

## Registry identity

Registry 使用有界 struct key：

```go
type circuitBreakerKey struct {
    authority string
    group     string
}
```

`authority` 使用 lowercase 的 `req.URL.Host`，包含 port；`group` 使用正規化後的 context value。

```text
{query2.finance.yahoo.com, yahoo-auth}
{query2.finance.yahoo.com, yahoo-chart}
{query2.finance.yahoo.com, yahoo-timeseries}
{query2.finance.yahoo.com, yahoo-options}
```

Registry 提供未 export 的 `forRequest(authority, group)`。既有 `forHost(host)` 保留並委派到
`forRequest(host, "")`，使現有 Go tests 與未標記 caller 不需改變。

同一 authority + group 永遠取得相同 breaker；authority 或 group 任一不同即取得獨立 breaker。
Factory、rolling window、half-open 與 outcome recording 不變。

## Yahoo family mapping

Group constants 留在 `svc/yahoo`，`utils/httpx` 不認識 Yahoo URL：

| Group | Requests | Batch commands |
| --- | --- | --- |
| `yahoo-auth` | cookie bootstrap、`/v1/test/getcrumb`、`/v10/finance/quoteSummary` | info、holders、insider、recommendations、upgrades、calendar、SEC、ESG |
| `yahoo-chart` | `/v8/finance/chart` | history、actions、metadata，以及 quote/company MIC inference |
| `yahoo-timeseries` | `/ws/fundamentals-timeseries/...` | income、balance、cashflow |
| `yahoo-options` | `/v7/finance/options` | options |
| `yahoo-news` | `POST /xhr/ncp` | news |
| `yahoo-web` | Yahoo HTML pages | earnings-dates、analysis/insights scrape |

`isin` 使用不同的 external authority，維持 ungrouped host-level breaker。TWSE 與其他 caller 也維持
ungrouped，不因本次改變。

### Request construction

`svc/yahoo` 新增單一小型 helper，負責以 group 包裝 context：

```go
func circuitContext(ctx context.Context, group string) context.Context {
    return httpx.WithCircuitGroup(ctx, group)
}
```

每個 endpoint file 在建立 request 前選擇固定常數。這保留一檔一職責：helper 只管理 Yahoo
family label，不負責 URL、HTTP method、decode 或 retry。

`svc/scrape.Client.Fetch` 在呼叫 `Caller.Get` 前使用 `yahoo-web` group。Registry key 仍包含實際
authority，因此相同 label 不會令不同網站共享 breaker。

## 行為與錯誤處理

Breaker outcome 保持現況：

| 最終結果 | Outcome |
| --- | --- |
| 2xx | success |
| 一般 4xx | availability success |
| 429 | failure |
| 5xx | failure |
| network error | failure |
| caller cancellation | neutral |

Consequences：

- `yahoo-auth` 累積足夠 429 後仍會開路；後續受保護 command 可回傳 `ErrCircuitOpen`。
- `yahoo-chart` 沒有收到 auth outcome，因此 metadata 仍會送出真正的 chart request。
- options 的 401 屬一般 4xx，維持 availability success，並回傳實際 HTTP error而非開路。
- 某個 family 的 half-open probe 只影響該 authority + family。
- 沒有 group 的 request 完全保持目前行為。

## 可觀測性

本次不增加 Prometheus label，避免新增高基數或破壞現有 dashboard。既有 endpoint、host、status、
retry 與 `circuit_open` 指標繼續使用。

Unit test 透過 registry identity 直接驗證 group isolation；live artifact 與 failure text 用來證明
沒有跨 family 開路。若未來需要 production group telemetry，另案設計固定 enum label。

## 測試設計

### `utils/httpx`

- 同 host + 同 group 回傳相同 breaker。
- 同 host + 不同 group 回傳不同 breaker。
- group 大小寫與空白正規化。
- ungrouped `forHost` 與 `forRequest(host, "")` 相同。
- `yahoo-auth` threshold=1 開路後，同 host 的 `yahoo-chart` request 仍成功。
- 同 group 的下一個 request 回傳 `ErrCircuitOpen`。
- 既有 host isolation、logical-request outcome、half-open 與 cancellation tests 全部保持通過。

### `svc/yahoo`

使用同一個 `httptest.Server` authority：

1. crumb endpoint 回傳 429，使 `yahoo-auth` 開路。
2. metadata/chart endpoint 回傳 200。
3. 驗證 metadata 正常 decode，而不是 `ErrCircuitOpen`。
4. 驗證下一個 auth request仍被 auth breaker 拒絕。

另外為 timeseries、options 與 news request 驗證各自 group，不依賴 URL path inference。

### `svc/scrape`

驗證 HTML fetch 使用 `yahoo-web`，且其 503 不會阻擋同 authority 的 `yahoo-news` group。

## 驗收

Deterministic gate：

```bash
go test ./... -count=1
go test -race ./utils/httpx ./svc/yahoo ./svc/scrape ./facade ./cmd/dispatch -count=1
go build -o /tmp/yfin-circuit-groups-check .
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... | grep svc/
git diff --check
```

Live gate 使用正式 `config/effective.yaml`，不得暫時提高 minimum requests：

```bash
go run . --config config/effective.yaml --retry-max 1 --timeout 12s batch --ticker AAPL --force
```

預期：

- `8 success / 22 failed`。
- history、actions、income、balance、cashflow、news、isin、metadata 成功。
- metadata artifact 為本次 run 新寫入的有效 JSON。
- chart、timeseries、news、options family 不得因 auth failure 回傳 `circuit breaker is open`。
- auth 或 web family 自身仍可正常開路；不把所有 circuit-open text 視為失敗，判斷重點是不得跨 family。

## 相容性與回復

- `httpx.Client.Do`、`Caller.Get`、`NewClient` 與 Config public signature 不變。
- `WithCircuitGroup` 是 additive API。
- 未標記 request 的 breaker identity 與現況相同。
- 若需回復，只要移除 Yahoo group annotations 並讓 `Client.Do` 使用 `forHost`；資料 schema、artifact
  path 與 CLI manifest 均不受影響。

