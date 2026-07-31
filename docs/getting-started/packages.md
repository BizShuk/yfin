# Package Navigation

本頁是 code-reading guide，不重複維護 architecture tree。Canonical dependency graph、package ownership 與 hard decisions 請讀 [CLAUDE.md](../../CLAUDE.md)。

## 從哪裡開始 (Where to Start)

依工作類型選入口：

| 需求 | 第一個 package | 原因 |
| --- | --- | --- |
| 外部 Go application 抓資料 | `facade` | Public operation boundary |
| 使用或組合 DTO | `model` | Canonical data owner |
| 修改 Yahoo API decode | `svc/yahoo` | Provider API HTTP / decode / validation |
| 修改 HTML extractor | `svc/scrape` | Yahoo web parsing 與 robots policy |
| 修改 TWSE endpoint | `svc/twse` | TWSE HTTP / typed parser |
| 修改 shared transport | `utils/httpx` | Retry、rate limit、body cap、circuit breaker |
| 修改 persistent cache | `utils/cache` + `facade/cached_*` | Storage mechanics + public cache policy |
| 修改 CLI | `cmd/<domain>` | Cobra composition / output orchestration |
| 修改 YAML | `config` | Schema、loader、interpolation、validation |
| 修改 logs / metrics | `utils/obsv` | `slog`、Prometheus、tracing boundary |

## Public Boundaries

### `facade`

`facade.Client` 是 API、scrape 與 CLI 共用 handler。它持有一個 shared HTTP connection pool、provider clients 與 MIC cache。

主要 constructors：

```text
facade.NewClient()
facade.NewClientWithConfig(*httpx.Config)
```

TWSE 使用 facade-owned opaque `TwseClient`；consumer 不需命名 `svc/twse.Client`。Cache consumers 使用 `NewCachedClient`、`NewCachedTwseClient`、`NewCachedFoundationClient` 或 `NewRawCache`。

### `model`

`model` 是 lowest layer：

- plain SDK DTOs 使用 `float64`。
- `Normalized*` DTOs 使用 `ScaledDecimal` 保留 precision。
- Yahoo raw response shapes、scrape DTOs 與 TWSE typed rows 也由此擁有，讓 service 可以單向依賴 model。

`facade.Bar`、`facade.Quote` 等是 `model.*` compatibility aliases；新 code 可直接 import `model` 取得資料型別。

## Provider Packages

### `svc/yahoo`

負責 Yahoo JSON API / XHR：

- chart、quote、quoteSummary
- annual fundamentals-timeseries
- news 與 foundation command surfaces
- crumb / cookie authentication

Endpoint envelope 不穿越 facade boundary。

### `svc/scrape`

負責 explicit HTML scrape：

- `robots.txt` enforcement
- fetch metadata 與 response caps
- endpoint-specific parsing
- news JSON / regex extraction 與 normalization

Scrape 不是所有 API methods 的 automatic fallback。Caller 必須選擇 `Scrape*` surface；standalone soak 只有 `fundamentals → financials` 的 bounded equivalent fallback。

### `svc/twse`

負責 TWSE registry、dispatcher、fetchers 與 typed rows。CLI 只透過 `facade.TwseClient.Dispatch` 使用它。

## Infrastructure Packages

### `utils/httpx`

Shared HTTP client 提供：

- per-host rate limiter / profile
- retry + exponential backoff
- authority + optional logical-group circuit breaker
- gzip decode、body cap、middleware
- process-local transport wrapper

### `utils/cache`

提供 root-confined atomic JSON store 與 refresh tiers。Public read-through semantics 由 facade cache wrappers 擁有。

### `utils/obsv`

提供 structured `slog` 與 active HTTP Prometheus metrics。OpenTelemetry tracer 目前是 no-op boundary。

## CLI 與 Standalone Tools

`cmd` 是 composition root。User-facing CLI 是 flat tree；sub-package directory 不代表 nested command。

`cmd/soak` 是 standalone binary：

```bash
go run ./cmd/soak --help
```

它不會出現在 `yfin --help`。

## Verify Current Surface

不要從本頁複製 type definitions；直接查 code：

```bash
go doc ./facade
go doc ./model
go doc ./utils/httpx
go list -deps ./...
```

Dependency rule 的 executable check：

```bash
go test ./tests/unit -run TestInternalPackagesFollowDependencyContract
```

此 test 是 internal dependency boundary 的 pass/fail owner；本頁只說明如何找到它。
