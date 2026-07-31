# 組態參考 (Configuration Reference)

`yfin` 以 typed YAML 管理 HTTP、rate limit、retry、circuit breaker、scrape policy、observability 與 secret references。CLI flags 只覆寫其明確擁有的 runtime knobs；process-local test hooks 不寫入 YAML。

## 1. 總覽 (Overview)

### 覆寫優先順序 (Precedence)

由高至低，依序覆寫：

```
CLI Flags > Environment Variables > Config File > Built-in Defaults
```

- CLI flag：root 級 persistent flags，例如 `--log-level` / `--qps` / `--timeout` / `--concurrency` / `--retry-max`，在任何子命令之前解析；只有明確傳入的 flag 才覆寫下一層。
- 環境變數 (Environment Variables)：僅供 log / observability 設定使用（走 `gosdk` 的 `APP_*` 慣例）；YAML 內 `${VAR}` / `${VAR:-default}` 插值則於 `Loader.Load()` 階段展開。
- 設定檔 (Config File)：單一 YAML（`config/effective.yaml` 或 `--config` 指定路徑），由內建 loader 載入。
- 預設值 (Defaults)：YAML defaults 見 `config/effective.yaml`；scrape HTTP policy 由 `utils/httpx.DefaultScrapeConfig()` 擁有；root type 定義於 `config/config.go`。

### 設定檔位置 (File Locations)

| 路徑 | 用途 |
| --- | --- |
| `config/effective.yaml` | 預設設定檔（相對於工作目錄） |
| `--config <path>` | 自訂路徑（root flag） |
| `~/.config/yfin/data/` | raw cache 與 batch artifacts |
| `~/.config/yfin/logs/` | PM2 / redirected logs 的預留目錄 |

> 未指定 `--config` 時，所有 loader caller 都由 `config.NewLoader` 統一解析為 `config/effective.yaml`。
> `./run.sh` 可重複建立兩個 runtime directories，不會啟動程式或覆寫 config。

## 2. CLI 旗標 (CLI Flags)

以下為 root 級 persistent flags（`cmd/global.go`），每個子命令都會繼承：

| 旗標 | 型別 | 預設 | 用途 |
| --- | --- | --- | --- |
| `--config` | string | `""` | YAML 設定檔路徑；空白時載入 `config/effective.yaml` |
| `--log-level` | string | `"info"` | Default `slog` 與 command observability 的最低 level；明確傳入時覆寫 `APP_LOG_LEVEL` / YAML |
| `--run-id` | string | `""` | Run ID（追蹤用；空白則自動產生） |
| `--concurrency` | int | `0` | `pull` universe / multi-ticker `quote` global worker limit；`0` 沿用 YAML，實際值仍受 per-host limit 限制 |
| `--qps` | float | `0` | Per-host QPS（覆寫 YAML） |
| `--retry-max` | int | `0` | HTTP retry 嘗試次數 |
| `--sessions` | int | `0` | **Vestigial**：HTTP client 已無 session rotation；保留旗標為向後相容，無實際效果 |
| `--timeout` | duration | `0` | HTTP timeout（如 `6s`、`30s`） |
| `--observability-disable-tracing` | bool | `false` | Compatibility flag；目前 tracer 本來就是 no-op |
| `--observability-disable-metrics` | bool | `false` | 關閉 Prometheus metrics |

## 3. 設定檔結構 (Config File)

`config/effective.yaml` 為樹狀 YAML；頂層直接對應 `Config` struct（`config/config.go`）：

| Top-level Key | 職責 |
| --- | --- |
| `app` | 環境標籤（`dev` / `staging` / `prod`）與 run 識別 |
| `yahoo` | Yahoo API base URL、HTTP timeout、connection pool |
| `concurrency` | `pull` / `quote` multi-item fetch 的 global / per-host worker limits |
| `rate_limit` | Per-host QPS 與 burst |
| `retry` | HTTP retry 通用設定（`utils/httpx` 使用） |
| `circuit_breaker` | 失敗熔斷閾值與 reset timeout |
| `markets` | Allowed intervals（`["1d"]` daily-only）、MIC allowlist、adjustment policy |
| `fx` | 保留的 FX policy schema；目前沒有 runtime rate provider |
| `scrape` | 明確 Yahoo web surfaces 與 robots policy |
| `observability` | log level、Prometheus，以及保留給 future exporter 的 OTLP schema |
| `secrets` | URI 參照 metadata；`config` 命令輸出時會 redact，但 loader 不解析 URI |

### 範例片段 (Excerpt)

```yaml
app:
  env: "dev"
  run_id: ""

yahoo:
  base_url: "https://query2.finance.yahoo.com"
  timeout_ms: 6000
  max_conns_per_host: 64
  user_agent: "yfin/1.x"
```

完整 default 值見 `config/effective.yaml`。不同環境（`dev` / `staging` / `prod`）以 `app.env` 區分。

## 4. Effective Config 檢視 (Effective Config Inspection)

`yfin config --print-effective` 透過 `config.Loader.GetEffectiveConfig()` 還原載入後、插值完成、敏感值 redact 後的最終設定。

```bash
# dot-notation key=value（排序、redact secrets）
yfin config --print-effective

# JSON 格式
yfin config --print-effective --json

# 載入自訂 YAML（驗證無誤即印出）
yfin --config ./config/effective.yaml config --print-effective
```

範例輸出：

```
EFFECTIVE CONFIG (redacted)
app.env=dev
app.run_id=
yahoo.base_url=https://query2.finance.yahoo.com
...
scrape.robots_policy=enforce
```

> Redact 規則：`secrets[].ref` 與完整 secret 欄位名稱或其 suffix（如 `password`、`api_token`、`client_secret`、`private_key`）會替換為 `[REDACTED]`。一般名稱如 `key_statistics` 不會被遮蔽。

## 5. Scrape 設定 (Scrape-Specific Config)

對應 `ScrapeConfig`（`config/scrape.go`），控制 `svc/scrape` 引擎：

### 完整 scrape 區段

```yaml
scrape:
  enabled: true
  robots_policy: "enforce"        # enforce | warn | ignore
  cache_ttl_ms: 60000
  endpoints:
    key_statistics: true
    financials: true
    analysis: true
    profile: true
    news: true
```

### robots_policy 三種模式

| 模式 | 行為 |
| --- | --- |
| `"enforce"` | 嚴格遵守 robots.txt；違規時禁止抓取（生產推薦） |
| `"warn"` | 違規時 log warning 但放行（開發） |
| `"ignore"` | 完全跳過 robots.txt 檢查（僅測試） |

### Scrape HTTP 設定

Scrape 的 `user_agent`、timeout、QPS、burst、retry 與 body cap 不屬於 YAML schema。Loader 從 `utils/httpx.DefaultScrapeConfig()` 取得單一 default policy 並寫入 `ScrapeConfig.HTTP`；Library caller 可先 clone/修改該 `*httpx.Config`，再呼叫 `client.ConfigureScrape(&cfg.Scrape)`。

### Endpoints 結構（`ScrapeEndpointConfig`）

| 欄位 | 對應 CLI `--endpoint` 名稱 |
| --- | --- |
| `scrape.endpoints.key_statistics` | `key-statistics` |
| `scrape.endpoints.financials` | `financials` |
| `scrape.endpoints.analysis` | `analysis` |
| `scrape.endpoints.profile` | `profile` |
| `scrape.endpoints.news` | `news` |

> CLI 額外可指定 `balance-sheet` / `cash-flow` / `analyst-insights`，共用 financials / key-statistics 的 extractor；這幾項未於 YAML 中獨立列出。

## 6. FX / Observability

### FX（`fx` 區段）

`FXConfig` 是 compatibility policy schema：

- `fx.provider`：`none` / `yahoo-web`。
- `fx.target`：預設目標幣別（如 `EUR` / `JPY`）。
- `fx.cache_ttl_ms`：換算結果快取 TTL。
- `fx.yahoo_web.*`：保留的 Yahoo web policy 欄位。

目前 production path 沒有 FX rate provider，也不會讀寫 FX conversion cache。`pull --fx-target` 只驗證 fetched currency 是否已等於 target；跨幣別時回傳 error，不輸出假 rate 或 `cache_hit`。

### Observability（`observability` 區段）

`ObservabilityConfig` 控制 log / Prometheus / OTLP：

- `observability.logs.level`：`utils/obsv` handler 的 YAML level；precedence 為 explicit `--log-level` > `APP_LOG_LEVEL` > 此欄位 > `info`。
- `observability.metrics.prometheus.enabled` + `addr`：Prometheus exporter bind address。
- `observability.tracing.otlp.enabled` + `endpoint` + `sample_ratio`：保留的 tracing schema；enabled 時 loader 會驗證 endpoint，但目前不建立 exporter。

CLI 可用 `--observability-disable-metrics` 關閉 Prometheus server。`--observability-disable-tracing` 保留向後相容；目前 tracer 固定為 no-op。

### Concurrency scope

`concurrency.global_workers` 是 multi-item command 的預設 global limit；`concurrency.per_host_workers` 是 Yahoo host 的安全上限。`pull --universe-file` 與 multi-ticker `quote` 的實際 worker 數為：

```text
min(explicit --concurrency or global_workers, per_host_workers, item count)
```

Fetch 可並行，但 stdout 與 local export 仍依輸入順序處理。`scrape` 沒有 multi-ticker workload，不 consume 此 flag；`batch` 使用自己的 `--max-workers`；standalone `cmd/soak` 使用自己的 `--concurrency`。

## 7. 驗證規則 (Validation Rules)

Executable validation 由 `config.Loader.Load()` 與 `config/*_test.go` 單一擁有；文件不複製可由 test 判定的完整 assertion table。可用以下 command 驗證目前 schema：

```bash
go test ./config
go run . --config config/effective.yaml config --print-effective
```

Config violation 在 CLI composition 層對應 exit code `3`（`ExitConfigError`）。

## 8. 延伸閱讀 (See Also)

- [安裝與初次設定](../getting-started/install.md)
- [各子命令 CLI 範例](../cli/usage.md)
- [`yfin scrape` 三種 mode 與 endpoint 選擇](../scrape/cli.md)
- [Scrape 引擎架構](../scrape/overview.md)
- [Observability](observability.md)
