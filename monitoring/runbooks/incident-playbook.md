# Incident Response Playbook

## Scope

本 playbook 處理 `yfin` 擁有的 provider access、config、parsing、cache 與 local CLI behavior。Message bus、DLQ、Kubernetes、PM2 或其他 deployment platform 不屬於此 repo；相關變更必須依目標環境自己的 runbook 執行。

## Severity

| Severity | Example | Response |
| --- | --- | --- |
| SEV-1 | 所有主要 provider surfaces 無法使用，且無可接受替代路徑 | 立即隔離流量並通知 owner |
| SEV-2 | 單一高價值 endpoint、market 或大量 symbols 失敗 | 降速、縮小 scope、開始 evidence collection |
| SEV-3 | 個別 symbol、非關鍵 parser 或文件問題 | 建立 issue 並在正常工作流修復 |

## Initial Response

記錄：

- UTC start time
- affected command / endpoint / ticker
- exit code
- deployment or binary version
- `--run-id`
- user-visible impact

第一輪診斷：

```bash
yfin version
yfin config --print-effective
yfin --help
yfin scrape --help
```

若使用自訂 config：

```bash
yfin --config ./config/production.yaml config --print-effective
```

## Incident 1: Rate-Limit Spike

### Symptoms

- HTTP `429` / `503` 增加
- `yfin_retries_total`、`yfin_backoff_total` 或 `yfin_cb_open_total` 上升
- Latency 明顯增加

### Containment

```bash
yfin --qps 0.5 scrape \
  --check \
  --ticker AAPL \
  --endpoint key-statistics
```

1. 降低 scheduled workload 的 QPS；若 workload 是 `pull` universe / multi-ticker `quote`，再降低 concurrency。
2. 暫停非必要 batch。
3. 等待 cooling period。
4. 不要提高 `--sessions`；session rotation 已移除。

### Recovery

逐步恢復 QPS，每一步都執行代表性 connectivity / parsing check：

```bash
yfin --qps 1 scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics
```

## Incident 2: Layout or Schema Drift

### Symptoms

- HTTP 200，但 `ErrNoQuoteSummary`、`ErrJSONDecode`、`ErrMissingFieldBase` 或 `ErrSchemaDriftBase`
- 同一 endpoint 跨多個 tickers 解析失敗

### Triage

```bash
for ticker in AAPL MSFT BRK-B; do
  yfin --log-level debug scrape \
    --preview-json \
    --ticker "$ticker" \
    --endpoints key-statistics
done
```

若只有 news 受影響：

```bash
yfin --log-level debug scrape --preview-news --ticker AAPL
```

### Resolution

1. 在 `svc/scrape` 的 owner extractor 重現。
2. 保存最小、合法且已去除敏感資料的 fixture。
3. 先加入 failing parser test。
4. 修正 selector / decoder / normalization。
5. 執行 unit tests。
6. 對至少三個 symbols 做 live verification。

## Incident 3: Authentication or Paid Surface

### Symptoms

- HTTP `401` / `403`
- `fundamentals` 回 paid-subscription error，exit code `2`

### Response

先確認該 surface 是否本來就需要 subscription。若需求可由 explicit scrape surface滿足：

```bash
yfin scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints financials,key-statistics,analysis,profile
```

不要將 scrape 描述為隱式 fallback；operator 與 downstream caller 必須明確選擇 surface。

## Incident 4: Provider or Network Outage

### Triage

```bash
# Yahoo JSON surface
yfin quote --tickers AAPL

# Yahoo web surface
yfin --timeout 10s scrape \
  --check \
  --ticker AAPL \
  --endpoint profile
```

如果兩者都失敗：

1. 檢查 DNS、egress、proxy 與 provider status。
2. 確認不是 config path 或 local certificate 問題。
3. 暫停非必要 writes / schedules。
4. 由 deployment owner 決定 rollback、restart 或 traffic routing；本 repo 不假設 deployment platform。

## Incident 5: Configuration Failure

### Symptoms

- `open ...: no such file or directory`
- config validation error
- observability initialization error
- `scrape` 被設定為 disabled

### Triage

```bash
yfin config --print-effective
yfin --config ./config/production.yaml config --print-effective
```

檢查：

- command 是否從可解析 `config/effective.yaml` 的 working directory 執行
- custom path 是否放在 Global Flag `--config`
- `markets.allowed_intervals` 是否只有 `1d`
- retry / circuit-breaker / observability required fields 是否有效
- `scrape.endpoints.key_statistics` 等一般欄位是否未被誤 redact

## Incident 6: Data Quality or Empty Result

空 chart series 可以是有效 no-data result，不等於 transport failure。確認：

- ticker 與 exchange
- date range / market holiday
- output metadata
- downstream 是否錯把 empty batch 當成 error

不要自行補造 currency、bar 或 observation。若需官方 market fallback，應由擁有該 market context 的 downstream application 處理。

## Evidence Collection

```bash
incident_dir="./tmp/incident-$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "$incident_dir"

yfin version > "$incident_dir/version.txt"
yfin config --print-effective > "$incident_dir/effective-config.txt"
yfin --log-level debug scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics \
  > "$incident_dir/command.log" 2>&1
```

`tmp/` 僅保存 incident-specific artifacts，不放 source logic。檢查檔案內容後再分享，避免 response payload 含個人或敏感資料。

## Communication Template

```text
Status: investigating | mitigated | resolved
Severity: SEV-1 | SEV-2 | SEV-3
Started (UTC):
Affected surface:
Affected symbols/markets:
User impact:
Current containment:
Evidence:
Next update:
Owner:
```

## Recovery Gate

標記 resolved 前：

1. Root cause 與 affected surface 已確定。
2. Config inspection 成功。
3. Relevant unit tests 通過。
4. Representative live checks 成功。
5. Error / rate-limit metrics 回到可接受區間。
6. Deployment acceptance 與 code verification 分開記錄。
7. Follow-up items 寫入 `README.todo` 或 issue tracker。

## Post-Incident

在 `docs/memory/YYYY-MM-DD-<topic>.md` 記錄需要長期遵守的 hard decision；已發生的 release event 寫入 `docs/CHANGELOG.md`。暫時措施與未完成事項不要寫入 `CLAUDE.md`。

## See Also

- [Scrape Surface Runbook](scrape-surface.md)
- [Error Handling](../../docs/operations/error-handling.md)
- [Configuration](../../docs/operations/configuration.md)
- [Observability](../../docs/operations/observability.md)
