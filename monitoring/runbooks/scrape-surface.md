# Explicit Scrape Surface Operator Runbook

## Purpose

本 runbook 用於診斷 Yahoo API 與 explicit scrape surface。主 `yfin` CLI 不會自動在 API 與 scrape 之間切換，也沒有 `yfin scrape --fallback` mode；operator 必須依資料需求明確選擇 command。

Standalone `cmd/soak` 另有 `--fallback auto|api-only|scrape-only`，只用於 soak testing。

## Preconditions

- 從 repository root 執行，或以 `--config <path>` 指定可讀的 YAML。
- 未指定 `--config` 時，loader 使用 `config/effective.yaml`。
- Live provider checks 會對 Yahoo 發出 request；先確認執行環境與 rate-limit policy。

## Quick Triage

```bash
# 1. 確認 binary 與 command surface
yfin version
yfin scrape --help

# 2. 驗證 effective config
yfin config --print-effective

# 3. 檢查單一頁面 connectivity
yfin scrape --check --ticker AAPL --endpoint key-statistics

# 4. 驗證 extractor
yfin scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics

# 5. 驗證 news parser
yfin scrape --preview-news --ticker AAPL
```

`--check` 下載 response 但不解析 DTO。`--preview-json` 與 `--preview-news` 才驗證 parsing。

## Selecting the Data Surface

### Yahoo API / SDK Surface

使用對應 command 或 `facade.Client` method：

```bash
yfin quote --tickers AAPL
yfin pull --ticker AAPL --start 2026-07-01 --end 2026-07-30
yfin fundamentals --ticker AAPL --preview
```

`fundamentals` quarterly surface 可能需要 paid subscription。Annual statements 與 news 另有一等 JSON methods。

### Explicit Scrape Surface

```bash
yfin scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints financials,key-statistics,analysis,profile
```

支援 endpoint：

- `profile`
- `key-statistics`
- `financials`
- `balance-sheet`
- `cash-flow`
- `analysis`
- `analyst-insights`
- `news`

## Rate-Limit Incident

症狀包括 HTTP `429`、反覆 `503`、backoff 或 circuit breaker open。

```bash
# 以較低 request pressure 重現
yfin --qps 0.5 scrape \
  --check \
  --ticker AAPL \
  --endpoint key-statistics
```

處置：

1. 降低 `--qps`；`scrape` 沒有 multi-item worker pool，因此 `--concurrency` 不影響此 command。
2. 等待 provider cooling period。
3. 檢查 `yfin_requests_total`、`yfin_retries_total`、`yfin_backoff_total`、`yfin_cb_open_total`。
4. 不要使用 `--sessions`；它是無作用的 compatibility flag。
5. 恢復流量時逐步提高 QPS，並在每一步重新驗證。

## Layout / Schema Drift

若 connectivity 成功但 extractor 失敗：

```bash
for ticker in AAPL MSFT BRK-B; do
  yfin --log-level debug scrape \
    --preview-json \
    --ticker "$ticker" \
    --endpoints key-statistics
done
```

收集：

- command 與 exit code
- UTC timestamp 與 `--run-id`
- affected ticker / endpoint
- redacted effective config
- debug log
- 可合法保存的 response fixture

修正應落在 `svc/scrape` 對應 extractor，並加入 parser fixture test。不要在 `cmd/` 複製 provider parsing。

## Robots.txt Decision

`robots_policy` 位於 YAML 的 `scrape` section：

```yaml
scrape:
  enabled: true
  robots_policy: enforce
```

`--force` 是無額外效果的 compatibility flag，也不會停用 robots enforcement。若 `ErrRobotsDenied`：

1. 保留 `enforce`。
2. 確認受影響 path。
3. 改用 Yahoo API 或其他允許的 endpoint。
4. 將 policy 變更視為獨立的 compliance decision，不在 incident 中臨時繞過。

## Observability

當 `observability.metrics.prometheus.enabled` 為 `true` 時，檢查設定的 `addr`：

```bash
curl -fsS http://127.0.0.1:9090/metrics | rg 'yfin_(requests|retries|backoff|cb_open|decode_fail)'
```

News parser 另提供 `yfin_scrape_news_total` 與 `yfin_scrape_news_parse_latency_ms`。

Metric presence 只證明 collector / endpoint 有輸出；仍需以實際 command 驗證 provider 與 parser。

## Recovery Gate

Incident 可結束前需確認：

1. `yfin config --print-effective` 成功，且敏感值已 redacted。
2. 單一 connectivity check 成功。
3. 至少三個代表性 tickers 的相關 parser mode 成功。
4. 降速後沒有持續的 429 / circuit-open pattern。
5. Unit tests 通過；live provider acceptance 另有記錄。

## See Also

- [Incident Response Playbook](incident-playbook.md)
- [Scrape CLI](../../docs/scrape/cli.md)
- [Error Handling](../../docs/operations/error-handling.md)
- [Configuration](../../docs/operations/configuration.md)
- [Soak Testing](../../docs/cli/soak-testing.md)
