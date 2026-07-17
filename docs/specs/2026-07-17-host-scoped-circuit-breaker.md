# Host-scoped Rolling-window Circuit Breaker

## 結論

`utils/httpx.Client` 將使用每個 host 獨立的斷路器 (circuit breaker)，並以一次完整
`Client.Do` 的最終結果作為一筆樣本。正式 runtime 採時間型 rolling window、失敗率閾值與
最小樣本數；既有 Go API 的整數失敗次數模式保留為相容 fallback。

這項變更修復兩個已確認的放大因素：

- 同一個請求的每次 retry 都被重複計為 failure。
- `query1.finance.yahoo.com`、`query2.finance.yahoo.com` 與
  `finance.yahoo.com` 共用同一份 breaker 狀態。

它不處理 Yahoo TLS/browser fingerprint、crumb 429 或 endpoint 遷移；這些是獨立後續工作。

## 現況與根因

目前 `Client` 只持有一個 `*CircuitBreaker`。`Do` 在 retry loop 內呼叫
`RecordFailure()`，retry 用盡後又再呼叫一次。因此，一個設定為五次 attempt 的失敗請求
最多可累積六次 breaker failure。

設定契約也互相矛盾：

- YAML `failure_threshold` 驗證範圍是 `(0, 1]`，表示失敗率。
- adapter 將 `0.30` 乘以 100 轉成整數 `30`。
- breaker 將 `30` 當作累積 failure 次數。
- `CircuitWindow` 雖已傳入 breaker，實際判斷從未使用。
- closed state 的 success 不會移除或抵銷先前 failure。

結果是少數原始錯誤可快速打開全域 breaker，讓後續不同 host 的健康請求全部回傳
`ErrCircuitOpen`。

## 範圍

本次包含：

- per-host breaker registry。
- retry 完成後才記錄一次 request outcome。
- 時間型 rolling window 與 failure-rate 判斷。
- minimum-request gate。
- single-probe half-open 狀態。
- legacy failure-count fallback。
- YAML、effective config、文件與測試同步。

本次不包含：

- 修改 retry/backoff 次數或延遲。
- 修改全域 rate limiter。
- Yahoo cookie、crumb、TLS fingerprint 或 browser impersonation。
- Yahoo API/scrape endpoint 遷移。
- 重新定義 CLI command manifest。

## 設定契約

`httpx.Config` 保留既有欄位並新增：

```go
type Config struct {
    CircuitWindow       time.Duration
    FailureThreshold    int     // legacy count mode; > 0 時優先
    FailureRateThreshold float64 // preferred mode; (0, 1]
    MinimumRequests     int
    ResetTimeout        time.Duration
}
```

選擇規則：

1. `FailureThreshold > 0`：使用 legacy count mode。
2. 否則 `FailureRateThreshold > 0`：使用 failure-rate mode。
3. 兩者皆未設定：套用 `DefaultConfig()` 的 rate-mode 預設值。

`NewClient` 會先正規化 breaker 專用的零值；不改動 timeout、retry、QPS 等其他欄位。這讓
直接傳入部分 `httpx.Config` 的 caller 也能取得有效 breaker 設定。

CLI effective config 預設：

| 設定 | 值 | 語意 |
| --- | ---: | --- |
| `CircuitWindow` | `50s` | 只計算最近 50 秒 outcome |
| `FailureRateThreshold` | `0.30` | failure rate 達 30% 才能開路 |
| `MinimumRequests` | `10` | active window 至少 10 個 outcomes |
| `ResetTimeout` | `30s` | open 後等待 30 秒才允許 probe |
| `FailureThreshold` | `0` | 正式 runtime 不使用 legacy count mode |

`httpx.DefaultConfig()` 與 scrape HTTP defaults 使用同一套 rate/minimum 設定，但保留其既有
`CircuitWindow: 60s`。CLI 的 YAML `window: 50` 繼續優先，因此 SDK 與 CLI 的 window 不被
意外合併成同一來源。

YAML `circuit_breaker` 新增 `minimum_requests`，`failure_threshold` 保持 `(0, 1]`，不再轉成
整數百分比。`config.Config.GetHTTPConfig()` 直接將它映射到
`FailureRateThreshold`。`minimum_requests` 必須大於零；省略時由 defaults 填入 `10`。

直接使用 Go API 的既有 caller 若設定 `FailureThreshold: 2`，仍表示 active window 內累積
兩次 failure 後開路，因此不需立即改碼。若同時設定 count 與 rate，count mode 優先，避免
舊程式從 `DefaultConfig()` 覆寫 `FailureThreshold` 時意外仍走 rate mode。

## 元件設計

### Breaker registry

`Client` 將單一 `circuitBreaker` 改為未 export 的 registry。registry 以正規化後的
`req.URL.Host` 為 key；host 包含 port，因此不同測試 server 與不同 upstream authority 不會
共享狀態。

```text
Client
└── breaker registry
    ├── query1.finance.yahoo.com -> CircuitBreaker
    ├── query2.finance.yahoo.com -> CircuitBreaker
    └── finance.yahoo.com        -> CircuitBreaker
```

registry 使用 mutex 保護 lazy creation。breaker 本身仍各自持有 mutex，因此不同 host 的
狀態更新不需要共用同一把 breaker lock。

registry 只會為實際送出的 authority 建立 breaker；一個 `Client` 的 upstream host 集合是
有限的，因此本次不加入額外的 eviction timer。

### Rolling outcomes

每個 breaker 保存 active window 內的 outcomes：

```go
type circuitOutcome struct {
    at     time.Time
    failed bool
}
```

每次 `Allow`、`RecordSuccess` 或 `RecordFailure` 都先移除
`at < now-CircuitWindow` 的資料。

breaker 內部以可替換的 `now func() time.Time` 取得時間。production 使用 `time.Now`，unit
tests 注入 fake clock，避免以 `time.Sleep` 驗證 window expiry。

rate mode 的開路條件：

```text
sample count >= MinimumRequests
AND
failure count / sample count >= FailureRateThreshold
```

legacy count mode 的開路條件：

```text
active-window failure count >= FailureThreshold
```

closed state 的成功 outcome 會保留在 window 內，成為 failure-rate 分母；不再採用永久累積、
只能靠 half-open success 歸零的舊行為。

### Request outcome boundary

`Client.Do` 在 retry loop 內不再更新 breaker。每次邏輯請求最多完成一次 breaker outcome：

| 最終結果 | Breaker outcome | 理由 |
| --- | --- | --- |
| HTTP 2xx | success | upstream 可用 |
| HTTP 400/401/403/404/422 | success | upstream 有回應，屬 request/business 結果 |
| HTTP 429 | failure | upstream 暫時拒絕服務 |
| HTTP 5xx | failure | upstream availability failure |
| network/transport error | failure | 無法取得 upstream response |
| caller context canceled/deadline | neutral | caller 主動終止，不歸責 upstream |
| rate limiter 在送出前失敗 | neutral | request 未抵達 upstream |
| middleware 在送出前失敗 | neutral | request 未抵達 upstream |

HTTP 4xx 在 `Do` 的回傳語意仍是 error；只是在 breaker 的 availability 判斷中記為 success。

若 transport error 同時伴隨 `ctx.Err() != nil`，以 neutral 處理。若是 HTTP client 自身 timeout
且 caller context 尚未取消，則仍視為 network failure。

### Half-open single probe

breaker 進入 open 後，`Allow()` 在 `ResetTimeout` 到期時只允許一個 request，並標記
`probeInFlight`：

- 其他同 host request 仍收到 `ErrCircuitOpen`。
- probe success：轉回 closed，清空歷史 outcomes。
- probe failure：轉回 open，更新 open 時間並重新等待 `ResetTimeout`。
- probe neutral：轉回 open，不新增 failure，重新允許下一個 timeout 週期後再 probe。

這避免現有 half-open 狀態同時放行任意數量 request。

## API 與相容性

- `httpx.Caller` 不變。
- `Client.Do` 與 `Client.Get` 簽名不變。
- `NewCircuitBreaker(window, failureThreshold, resetTimeout)` 保留，建立 legacy count-mode
  breaker。
- 新增 `NewFailureRateCircuitBreaker(window, failureRateThreshold, minimumRequests,
  resetTimeout)`，供 `Client` runtime 建立 per-host breaker。
- `FailureThreshold int` 保留，避免外部 Go caller 編譯失敗。
- YAML `failure_threshold` 的既有 `(0, 1]` 值維持有效。
- 新增的 `minimum_requests` 若省略，loader 套用 `10`。

## 可觀測性與錯誤

`ErrCircuitOpen` 及既有 metrics/error wrapping 不變。endpoint label 仍沿用目前
`extractEndpoint` 結果；本次不增加高 cardinality host label。

`Meta.Attempts` 繼續反映實際 attempt 次數，與 breaker outcome 次數分離。

## 測試策略

### CircuitBreaker unit tests

- rate mode 未達 `MinimumRequests` 時不開路。
- 達 minimum 且 failure rate 達 threshold 時開路。
- window 外 outcomes 被移除。
- closed success 會影響 failure-rate 分母。
- half-open 僅允許一個 concurrent probe。
- probe success/failure/neutral 的狀態轉移。
- legacy count mode 維持 threshold 行為。

### Client integration tests

- 一個含五次 retry 的失敗 `Do` 只產生一個 failure outcome。
- host A 開路不會阻擋 host B。
- HTTP 400/401/404 回傳 error，但 breaker 記為 availability success。
- caller context cancellation 不新增 outcome。
- success、error、`Meta.Attempts` 與既有 response middleware 行為不退化。

### Config tests

- YAML `0.30` 映射為 `FailureRateThreshold == 0.30`。
- `FailureThreshold == 0`，避免 runtime 誤走 legacy mode。
- `minimum_requests` 的 default、override 與 validation。
- effective YAML 與文件範例同步。

### Regression gates

- `go test ./utils/httpx ./config -count=1`
- `go test -race ./utils/httpx ./facade ./cmd/dispatch -count=1`
- `go test ./... -count=1`
- `go build .`
- `cmd -> facade -> svc -> model` import gate
- `git diff --check`

## 完成條件

- retry attempt 不再放大 breaker failure count。
- 任一 Yahoo host 開路不影響其他 host。
- `CircuitWindow`、failure rate 與 minimum requests 實際參與判斷。
- half-open 同 host 同時最多一個 probe。
- 既有 legacy Go count configuration 可繼續使用。
- 所有 regression gates 通過。
