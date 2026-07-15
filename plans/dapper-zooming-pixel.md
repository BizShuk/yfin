# Plan: Consolidate HTTP config under `utils/httpx.Config`

## Context

`utils/httpx.Config` 已是 HTTP-layer 的事實 source-of-truth（`httpx.NewClient(*Config)` 的唯一參數）。但目前 `svc/scrape.Config`、`config.ScrapeConfig`、三條手寫 mapper 各自重新定義了它的子集欄位，造成欄位逐一對應且容易 drift，違反 CLAUDE.md「禁止跨套件複製貼上」與「既有慣例不得額外建立自訂選項」原則。

**目標**：把 HTTP 設定的所有權統一在 `utils/httpx` 之下；`svc/scrape` 與 `config` 套件改為**持有**（embed / 欄位引用）`httpx.Config`，不再重複欄位、不再需要手寫 mapper。

## 現況：3 條手寫 mapper + 2 個欄位重複結構

```tree
config 層
├── config.HTTPConfig (config/http.go + adapters.go)         ← 1st HTTP-layer 結構
│   ├── 由 Yahoo + RateLimit + Retry + CircuitBreaker 4 個 yaml sub-section 扁平組裝
│   └── (含 ms→Duration、seconds→Duration、FailureThreshold 比例→百分比 轉換)
│
└── config.ScrapeConfig (config/scrape.go)                   ← 2nd HTTP-layer 結構
    ├── 與 config.HTTPConfig 部分重複 (UserAgent, TimeoutMs, QPS, Burst, Retry{})
    └── 缺 8 個欄位 (IdleTimeout, MaxConnsPerHost, BackoffJitterMs,
        CircuitWindow, FailureThreshold, ResetTimeout, MaxBodyBytes)

runtime 層
├── svc/scrape.Config (svc/scrape/types.go)                 ← 3rd HTTP-layer 結構
│   ├── 與 config.ScrapeConfig 結構幾乎一樣
│   └── NewClient (deprecated wrapper) 自行 hardcode 8 個預設值填進 httpx.Config
│
└── utils/httpx.Config (utils/httpx/client.go)              ← canonical，但被埋沒
    └── 16 欄位的真實 consumer (httpx.NewClient 唯一參數)

3 條手寫 mapper（須逐一對應、容易 drift）
├── cmd/client.go:56    httpConfigToHttpx         — config.HTTPConfig → httpx.Config
├── facade/scrape.go:67 NewScrapeClientFromConfig — config.ScrapeConfig → svc/scrape.Config
└── svc/scrape/client.go:60 NewClient (deprecated) — svc/scrape.Config → httpx.Config
                                                    + 硬編碼 8 個 scrape-side 預設值
```

每一條都是「欄位逐一對應」。任何一個值改了，要同步 3 個地方，否則 drift。

## 目標結構

```tree
config 層
├── config.Config
│   ├── Yahoo / RateLimit / Retry / CircuitBreaker  ← yaml-side 形狀（不變）
│   ├── Scrape                                       ← yaml-side 形狀（不變）
│   ├── HTTP *httpx.Config                           ← 新增；post-load 從上面 4 個 section 組裝
│   └── (其他不變)
│
└── config.HTTPConfig                                ← 改為 alias / 移除

runtime 層
├── svc/scrape.Config
│   ├── HTTP *httpx.Config                           ← 新增；取代現有 7 個欄位
│   ├── Enabled / RobotsPolicy / CacheTTLMs / Endpoints (scrape-only，不變)
│   └── RetryConfig                                  ← 移除（移到 HTTP.*）
│
└── utils/httpx.Config                               ← canonical（不變）

Mapper 數量：0
（HTTP-related field mapping 只發生一次：yaml → *httpx.Config，集中在 config loader）
```

## 設計決策

### D1. 用「欄位引用」而非 Go embedding

`svc/scrape.Config` 與 `config.ScrapeConfig` 改為：

```go
type ScrapeConfig struct {
    HTTP         *httpx.Config   // 取代現有 UserAgent/TimeoutMs/QPS/Burst/Retry{}
    Enabled      bool            `yaml:"enabled"`
    RobotsPolicy string          `yaml:"robots_policy"`
    CacheTTLMs   int             `yaml:"cache_ttl_ms"`
    Endpoints    EndpointConfig  `yaml:"endpoints"`
}
```

理由：
- **避免 yaml 衝突**：embedding 會讓 httpx.Config 的 Go 欄位名（`Timeout`, `QPS`）直接對應到 yaml key，但 yaml 端用的是 `timeout_ms`（int）+ `qps`（float），且 `Timeout` 是 `time.Duration`（yaml 不會自動從 int 轉）。改用 named field `HTTP *httpx.Config` + `yaml:"-"` 可完全繞過這個問題。
- **避免命名衝突**：outer struct 也有 `Config` 名稱，embedding `httpx.Config` 會得到一個叫 `Config` 的內部欄位，命名衝突。named field 沒有這個問題。
- **保留外部測試可讀性**：測試現在用 `cfg.Scrape.Enabled` 直接存取，refactor 後變 `cfg.Scrape.HTTP.Timeout`（明確是 HTTP 設定）。
- **向後相容性最高**：HTTP 設定從 yaml 扁平欄位轉換到 `*httpx.Config` 的動作集中於 config loader，呼叫端完全感知不到。

### D2. yaml 形狀完全不變

`config/effective.yaml` 的 `scrape:` 區段維持目前的扁平結構：

```yaml
scrape:
  enabled: true
  user_agent: "Mozilla/5.0 ..."
  timeout_ms: 10000
  qps: 0.7
  burst: 1
  retry:
    attempts: 4
    base_ms: 300
    max_delay_ms: 4000
  robots_policy: enforce
  cache_ttl_ms: 60000
  endpoints:
    key_statistics: true
    ...
```

理由：用戶已佈署的 `config/effective.yaml` 不能破。loader 內部把扁平欄位組裝成 `*httpx.Config`，呼叫端只看到後者。

### D3. `config.HTTPConfig` 改為 type alias

```go
// config/http.go
type HTTPConfig = httpx.Config  // type alias — 對外部程式完全 back-compat
```

理由：保留 `config.GetHTTPConfig()` 簽章，避免改動 `cmd/client.go` 與測試呼叫端（除非要順手清掉）。

### D4. `cmd/client.go:httpConfigToHttpx` 移除

`cfg.GetHTTPConfig()` 直接回傳 `*httpx.Config` 後，mapper 不再需要。`cmd/client.go` 簡化為：

```go
httpxConfig := cfg.GetHTTPConfig()
if httpxConfig == nil {
    httpxConfig = httpx.DefaultConfig()
}
// CLI overrides
if Global.QPS > 0 { httpxConfig.QPS = Global.QPS }
...
```

## 檔案異動清單

| 檔案 | 動作 |
| --- | --- |
| `utils/httpx/client.go` | 不動（canonical） |
| `config/scrape.go` | 移除 HTTP 欄位（`UserAgent/TimeoutMs/QPS/Burst/Retry{}`），新增 `HTTP *httpx.Config` |
| `config/scrape_retry.go`（新） | 把 `ScrapeRetryConfig` 移除（合併進 `RetryConfig`） |
| `config/adapters.go` | 新增 `Config.assembleHTTPConfig()` 從 yaml-side 欄位組裝 `*httpx.Config`；`GetHTTPConfig()` 直接回傳組裝結果 |
| `config/defaults.go` | `scrape:` defaults 結構不變；loader 組裝 `HTTP *httpx.Config` 後把它一併填入 |
| `config/http.go` | `type HTTPConfig = httpx.Config`（type alias） |
| `config/loader_test.go` | 更新 `TestGetHTTPConfig` 為新的回傳型別 |
| `tests/unit/config_test.go` | 更新 `cfg.Yahoo.TimeoutMs → cfg.HTTP.Timeout.Milliseconds()` 等測試斷言 |
| `svc/scrape/types.go` | 移除 HTTP 欄位（`UserAgent/TimeoutMs/QPS/Burst/Retry{}`），新增 `HTTP *httpx.Config` |
| `svc/scrape/client.go` | `NewClient` deprecated wrapper 移除手寫 mapper：`caller = httpx.NewClient(config.HTTP)`，預設值不再 hardcode（改由 caller 端補 `httpx.DefaultConfig()` 或 caller 自己填） |
| `facade/scrape.go:NewScrapeClientFromConfig` | 移除 `Retry`/`UserAgent`/`TimeoutMs`/`QPS`/`Burst` 欄位逐一複製，改為 `scrape.Config{HTTP: cfg.HTTP, ...}` 單行 |
| `cmd/client.go:httpConfigToHttpx` | 刪除；`CreateClient()` 直接用 `cfg.GetHTTPConfig()` |
| `cmd/soak/main.go` | 把 TODO 註解更新（`cfg.HTTP.Timeout` 等） |

總計 ~13 檔，集中在 config/scrape/svc-scrape/facade 四個套件。

## 移除的程式碼

```go
// 移除 #1: svc/scrape/client.go 的 hardcoded 預設值
httpx.NewClient(&httpx.Config{
    BaseURL:          "https://finance.yahoo.com",
    IdleTimeout:      90 * time.Second,
    MaxConnsPerHost:  10,
    BackoffJitterMs:  config.Retry.BaseMs / 2,
    CircuitWindow:    60 * time.Second,
    FailureThreshold: 5,
    ResetTimeout:     30 * time.Second,
    MaxBodyBytes:     8 << 20,
    // ... 加上其它從 scrape.Config 複製的欄位
})

// 移除 #2: facade/scrape.go 的 17 行手寫 mapper
return scrape.NewClient(&scrape.Config{
    Enabled:   cfg.Enabled,
    UserAgent: cfg.UserAgent,
    TimeoutMs: cfg.TimeoutMs,
    QPS:       cfg.QPS,
    Burst:     cfg.Burst,
    Retry: scrape.RetryConfig{
        Attempts:   cfg.Retry.Attempts,
        BaseMs:     cfg.Retry.BaseMs,
        MaxDelayMs: cfg.Retry.MaxDelayMs,
    },
    RobotsPolicy: cfg.RobotsPolicy,
    CacheTTLMs:   cfg.CacheTTLMs,
    Endpoints: scrape.EndpointConfig{
        KeyStatistics: cfg.Endpoints.KeyStatistics,
        ...
    },
}, nil)

// 移除 #3: cmd/client.go:httpConfigToHttpx 的 16 行手寫 mapper
func httpConfigToHttpx(cfg *config.HTTPConfig) *httpx.Config {
    return &httpx.Config{
        BaseURL:          cfg.BaseURL,
        ...
        FailureThreshold: int(cfg.FailureThreshold * 100),
        ...
    }
}
```

## 集中化的轉換（單一來源）

所有「yaml 扁平欄位 → `*httpx.Config`」的轉換集中於 `config/adapters.go`：

```go
// assembleHTTPConfig builds *httpx.Config from yaml-side sub-sections.
// Called once during Loader.Load().
func (c *Config) assembleHTTPConfig() *httpx.Config {
    jitter := c.Retry.BaseMs / 2
    if c.Retry.BaseMs <= 0 {
        jitter = 0
    }
    return &httpx.Config{
        BaseURL:          c.Yahoo.BaseURL,
        Timeout:          time.Duration(c.Yahoo.TimeoutMs) * time.Millisecond,
        IdleTimeout:      time.Duration(c.Yahoo.IdleTimeoutMs) * time.Millisecond,
        MaxConnsPerHost:  c.Yahoo.MaxConnsPerHost,
        UserAgent:        c.Yahoo.UserAgent,
        MaxAttempts:      c.Retry.Attempts,
        BackoffBaseMs:    c.Retry.BaseMs,
        BackoffJitterMs:  jitter,
        MaxDelayMs:       c.Retry.MaxDelayMs,
        QPS:              c.RateLimit.PerHostQPS,
        Burst:            c.RateLimit.PerHostBurst,
        CircuitWindow:    time.Duration(c.CircuitBreaker.Window) * time.Second,
        FailureThreshold: int(c.CircuitBreaker.FailureThreshold * 100), // 0–1 → 0–100
        ResetTimeout:     time.Duration(c.CircuitBreaker.ResetTimeoutMs) * time.Millisecond,
        // MaxBodyBytes = 0 (unlimited) — same default as before
    }
}

func (c *Config) assembleScrapeHTTPConfig() *httpx.Config {
    // Same shape but reads from c.Scrape sub-tree.
    // Used to populate c.Scrape.HTTP after load.
}
```

## 風險評估

| 風險 | 緩解 |
| --- | --- |
| `cfg.HTTP.Timeout` 是 `time.Duration`；yaml 端是 `timeout_ms: int`。loader 必須在 Load() 時組裝；放在 assembleHTTPConfig 集中處理 | ✅ 集中單一點 |
| `config.HTTPConfig` 改為 `httpx.Config` alias；呼叫端已使用 `httpx.Config` 欄位名稱的測試不會壞，但若有人用 `time.Duration` 比較等需要驗證 | ✅ alias 提供 back-compat |
| 既有 `cmd/soak/main.go` 的 TODO 註解指向 `cfg.Yahoo.TimeoutMs`；refactor 後變 `cfg.HTTP.Timeout` | ✅ 順手更新 |
| `tests/unit/config_test.go` 既有 `cfg.Yahoo.TimeoutMs` 等斷言需要更新為 `cfg.HTTP.Timeout` | ✅ 預期更新測試 |
| `svc/scrape.Config.HTTP` 為 nil 時 `httpx.NewClient(nil)` 行為：目前 `httpx.NewClient` 已處理 nil → `DefaultConfig()` | ✅ 安全 |
| BackoffJitterMs 預設值 `Retry.BaseMs / 2`：當 BaseMs=0 時會 panic；既有程式碼就有這個 bug（`0 / 2 = 0`，不會 panic）。但若有人把 BaseMs 設成負數則 jitter 是負數，會有 backoff bug | ✅ 加上 guard |
| scrape 的 `MaxBodyBytes = 8 << 20`（8 MiB）先前在 hardcoded；refactor 後會變 0（unlimited），造成 scrape 行為改變 | ✅ 必須在 `assembleScrapeHTTPConfig` 中明確設 8 MiB，避免 silent regression |

## 驗證步驟

```bash
# 1. 全套 build + vet
go build ./...
go vet ./...

# 2. 全套 unit + integration 測試（含 race detector）
go test ./... -count=1
go test ./... -count=1 -race

# 3. scrape 套件覆蓋率（驗證 mapper 真的被拿掉而不是繞路）
go test ./svc/scrape/... ./tests/unit/scrape/... -coverpkg=./svc/scrape/... -cover

# 4. 手動驗證 scrape CLI 仍可用
go build -o /tmp/yfin .
/tmp/yfin scrape --check --ticker AAPL --endpoint profile
/tmp/yfin scrape --preview-json --ticker AAPL --endpoints key-statistics
/tmp/yfin comprehensive-stats --ticker AAPL

# 5. 驗證 yaml 向下相容：刪除有效 config 後由 defaults 重建，確認新組裝出的 HTTP config 等於舊組裝
# （寫一個 _test.go 比對 cfg.HTTP 與手算的舊 GetHTTPConfig 結果）
```

## 遷移順序

1. **Phase 1**: 在 `utils/httpx` 不動的情況下，先在 `config/adapters.go` 加 `assembleHTTPConfig()` + `assembleScrapeHTTPConfig()`，讓 `Config.HTTP` 與 `Config.Scrape.HTTP` 在 `Load()` 結尾被填好。
2. **Phase 2**: 移除 `config.HTTPConfig` 自訂 struct，改用 type alias `httpx.Config`；更新 `config.GetHTTPConfig()` 回傳型別。
3. **Phase 3**: 移除 `svc/scrape.Config` 的 HTTP 重複欄位，改用 `HTTP *httpx.Config`；簡化 `NewClient`。
4. **Phase 4**: 簡化 `facade.NewScrapeClientFromConfig`（單行 pass-through）與 `cmd/client.go`（刪 `httpConfigToHttpx`）。
5. **Phase 5**: 更新測試斷言，跑完整 race + coverage 驗證。

每個 Phase 結束都跑 `go test ./... -count=1 -race`，確保零迴歸才進下一個 Phase。
