# Yahoo Scrape Surfaces Overview

Yahoo web scrape 是 yfin 的明確 data source，用於取得 API 未提供的 financial statements、profile、statistics、analysis 與 news。Root CLI 不會在 API command 失敗時偷偷切換 source；caller 必須直接呼叫 scrape method 或 `yfin scrape`。

## Supported Surfaces

| Endpoint | Data |
| --- | --- |
| `key-statistics` | valuation、shares、margin 與 historical statistics |
| `financials` | income statement |
| `balance-sheet` | balance sheet |
| `cash-flow` | cash-flow statement |
| `analysis` | estimates、EPS trend/revisions、growth |
| `analyst-insights` | analyst opinions 與 price targets |
| `profile` | company identity、sector、employees、executives |
| `news` | ticker news items |

`facade.Client` 提供兩種 output shape：

- `ScrapeFinancials`、`ScrapeKeyStatistics` 等 snapshot methods：回傳 plain model/facade types。
- `ScrapeFinancialsData`、`ScrapeProfileData`、`ScrapeKeyStatisticsData`、`ScrapeAnalysisData`：保留完整 scrape DTO。

## Source Selection

Main CLI：

```text
pull / quote / fundamentals -> Yahoo API only
scrape                     -> Yahoo web only
```

Standalone `cmd/soak` 可用 `auto` strategy，但唯一的 API-to-scrape mapping 是：

```text
fundamentals -> financials
```

`quote` 與 `daily-bars` 沒有 scrape equivalent。請勿把 `auto` 解讀為一般化的 seamless fallback。

## Request Flow

```text
CLI or SDK caller
      |
 facade.Client
      |
 ScrapeClient.Fetch
      |
 robots policy -> shared httpx client -> finance.yahoo.com
      |
 parser -> model DTO / plain snapshot
```

Transport facts：

- Yahoo API、web scrape 與 cookie bootstrap 共用一個 `http.Client` / connection pool。
- Host profile 分別擁有 QPS、burst、body cap、headers 與 endpoint tag。
- Retry、backoff、circuit breaker 與 request metrics 由 `utils/httpx` 統一處理。
- Scrape HTTP defaults 由 `utils/httpx.DefaultScrapeConfig()` 單一擁有。

完整 dependency graph 由 [CLAUDE.md](../../CLAUDE.md) 擁有。

## Robots Policy

`scrape.robots_policy`：

| Value | Behavior |
| --- | --- |
| `enforce` | disallowed path 回傳 error，不送出 page request |
| `warn` | 記錄 violation 後繼續；只適合受控開發 |
| `ignore` | 跳過 robots check；只適合 deterministic test |

Production default 是 `enforce`。不要用 `ignore` 跨過 provider policy gate。

## Data Semantics

- Parser 只回傳 upstream 能證明的欄位；缺值不以固定數字補齊。
- Scrape monetary values 在 comprehensive DTO 中使用 `ScaledDecimal` alias。
- 轉成 plain `FundamentalsSnapshot` 後，public line values 使用 `float64`。
- Event/boundary times 使用 UTC；daily trading date 的 exchange-local 語意由 `model` 處理。
- Snapshot `Source` 使用 `yfin/scrape/...` lineage。
- yfin 不執行 FX conversion，也不宣告 scrape cache hit。

## CLI

```bash
# Connectivity only
go run . scrape --check --ticker AAPL --endpoint profile --preview

# Parse complete DTOs
go run . scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics,financials,analysis,profile

# Parse news
go run . scrape --preview-news --ticker AAPL
```

三種 mode（`--check` / `--preview-json` / `--preview-news`）互斥。`--preview` 只修飾 check output；不是 `pull` 或 `quote` flag。

## SDK

```go
client := facade.NewClient()

profile, err := client.ScrapeProfileData(ctx, "AAPL", "run-123")
if err != nil {
    return err
}
fmt.Println(profile.Sector)
```

需要從已載入 YAML 套用 policy：

```go
client := facade.NewClientWithConfig(cfg.HTTP)
if err := client.ConfigureScrape(&cfg.Scrape); err != nil {
    return err
}
```

## Observability and Acceptance

通用 HTTP metrics 可觀察 requests、retries、backoff、decode failures、circuit state、latency 與 inflight requests。News parser 另有 news-specific parse metrics。Tracing 目前是 no-op boundary。

Live scrape acceptance 必須保留 exact ticker、endpoint、run ID、robots policy 與時間；unit fixtures 只能證明 parser contract，不能證明 Yahoo 當前頁面可達。

延伸閱讀：

- [Scrape CLI](cli.md)
- [Configuration Reference](../operations/configuration.md)
- [Error Handling](../operations/error-handling.md)
- [Soak Testing](../cli/soak-testing.md)
