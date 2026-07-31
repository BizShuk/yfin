# Scrape CLI Guide

`yfin scrape` 是 Yahoo web 的 explicit diagnostic / extraction surface。它不會被 `pull`、`quote` 或其他 API commands 自動呼叫。

## Modes

每次執行必須且只能選一個 mode：

| Mode | Required flags | Result |
| --- | --- | --- |
| `--check` | `--ticker` + `--endpoint` | Fetch response / metadata；不 parse DTO |
| `--preview-json` | `--ticker` + `--endpoints` | Parse selected endpoint DTOs to stdout |
| `--preview-news` | `--ticker` | Parse / normalize news to stdout |

```bash
yfin scrape --check --ticker AAPL --endpoint profile

yfin scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics,financials,analysis,profile

yfin scrape --preview-news --ticker AAPL
```

## Endpoints

Current names：

- `profile`
- `key-statistics`
- `financials`
- `balance-sheet`
- `cash-flow`
- `analysis`
- `analyst-insights`
- `news`

`--endpoint` 是 check mode 的 single value；`--endpoints` 是 preview-json 的 comma-separated values。

## Preview and Compatibility Flags

- `--preview` 只控制 check output preview。
- `--force` 是 compatibility flag，沒有額外 runtime effect。
- `--force` 不會停用 robots enforcement。
- `--concurrency` 雖是 inherited root flag，但 scrape 沒有 multi-symbol worker pool，因此不 consume。

## HTTP Overrides

以下範例統一把 global flags 放在 subcommand 前；Cobra parser 也接受放在 subcommand 後：

```bash
yfin \
  --qps 0.7 \
  --retry-max 4 \
  --timeout 20s \
  --log-level debug \
  scrape \
  --check \
  --ticker AAPL \
  --endpoint key-statistics
```

Overrides 套用在 shared HTTP connection pool / Yahoo web host profile。Scrape defaults 由 `utils/httpx.DefaultScrapeConfig()` 擁有；YAML 不再複製 timeout / qps / retry fields。

## Robots Policy

`scrape.robots_policy` 控制 behavior：

- `enforce`：依 robots result 決定是否 fetch。
- 其他值若不符合 loader / service contract，configuration validation 失敗。

不要把 connectivity success 當成 parser acceptance。`--check` 只證明 response 已取得；`--preview-json` / `--preview-news` 才執行 extractor。

## Schema Drift Workflow

1. 用單一 ticker / endpoint 重現。
2. 啟用 debug logging，保存 exact error 與 UTC timestamp。
3. 在合法範圍內保存最小 fixture。
4. 先加入 failing parser test。
5. 修改對應 `svc/scrape` extractor。
6. 以多個 symbols 做 live verification。

```bash
for ticker in AAPL MSFT BRK-B; do
  yfin \
    --log-level debug \
    scrape \
    --preview-json \
    --ticker "$ticker" \
    --endpoints key-statistics
done
```

## Error Classes

常見 typed conditions：

- robots denied
- timeout / network failure
- rate limit / circuit open
- response body too large
- missing quote-summary payload
- JSON unescape / decode failure
- missing field / schema drift
- news no-articles / parse failure

Classifier 支援 wrapped typed errors；string fallback 是 compatibility path。

## Batch / Soak

CLI scrape 本身不接受 universe file。少量 symbols 可用 shell loop；long-running workload 使用 standalone soak：

```bash
go run ./cmd/soak --help
```

Soak 的 `auto` 只允許有 equivalent contract 的 `fundamentals → financials` fallback，不是一般化 API-to-scrape switch。

## Verification

Deterministic：

```bash
go test ./svc/scrape ./facade ./cmd/scrape
go test -tags=integration ./... -run '^$'
```

Live：

```bash
yfin scrape --check --ticker AAPL --endpoint profile
```

Live result 依 provider availability，需獨立記錄。

## See Also

- [Scrape Overview](overview.md)
- [Configuration](../operations/configuration.md)
- [Error Handling](../operations/error-handling.md)
- [Soak Testing](../cli/soak-testing.md)
