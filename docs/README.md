# yfin Documentation Index

`yfin` 是 Yahoo Finance 的 Go 資料擷取與正規化工具，模組路徑 `github.com/bizshuk/yfin` (Go 1.26.0)。提供 CLI 指令（flat 結構，無 nested group）、公開 `facade` 資料型別（外部專案導入邊界），以及 `config/`、`svc/{yahoo,scrape,emit,norm,twse}/`、`utils/{httpx,bus,cache,obsv}/` 等核心套件。本目錄依讀者角色分層收納使用手冊、API 參考、維運指引與歷史歸檔文件。

## Getting Started (新進讀者)

- [`getting-started/install.md`](getting-started/install.md) — 安裝 `yfin` CLI 與相依工具 (Go 1.26+)
- [`getting-started/onboarding.md`](getting-started/onboarding.md) — 第一次跑通本地開發環境、跑測試、調整設定
- [`getting-started/packages.md`](getting-started/packages.md) — 套件總覽 (`facade/`、`svc/*`、`utils/*`、`config/`、`cmd/*`)

## CLI (操作者)

- [`cli/usage.md`](cli/usage.md) — 常用 CLI 操作流程 (quote、scrape、dispatch、batch)
- [`cli/commands.md`](cli/commands.md) — 所有指令與旗標 (flags) 的完整參考
- [`cli/soak-testing.md`](cli/soak-testing.md) — 長時間穩定性壓測 (soak test) 與基準數據

## API (整合開發者)

- [`api/reference.md`](api/reference.md) — Go API 公開介面、函式簽名、回傳型別
- [`api/data-structures.md`](api/data-structures.md) — `facade.Bar`、`facade.Quote` 與 `ScaledDecimal` 資料模型
- [`api/examples.md`](api/examples.md) — 從外部專案呼叫 `yfin` 套件的範例程式

## Operations (維運 / SRE)

- [`operations/configuration.md`](operations/configuration.md) — YAML 設定檔結構、`~/.config/yfin/` 路徑、env 變數
- [`operations/observability.md`](operations/observability.md) — 指標 (metrics) 與日誌 (logs) 串接 `inf` (VictoriaMetrics / Loki)
- [`operations/performance.md`](operations/performance.md) — QPS rate limiting、重試、circuit breaker 參數調校
- [`operations/error-handling.md`](operations/error-handling.md) — 錯誤分類、暫時性錯誤處理、退避 (backoff) 策略
- [`operations/data-quality.md`](operations/data-quality.md) — 資料品質監控、缺值偵測、異常告警

## Scrape (專家)

- [`scrape/overview.md`](scrape/overview.md) — Yahoo web scraping 引擎架構、`robots.txt` 合規
- [`scrape/cli.md`](scrape/cli.md) — `yfin scrape` 子指令、selector、輸出格式

## Integrations (整合開發者)

- [`integrations/ampy-proto.md`](integrations/ampy-proto.md) — `yfin` 與 `ampy-proto` 資料格式對接
- [`integrations/migration-guide.md`](integrations/migration-guide.md) — 從舊版 API / 直接讀 Yahoo Finance 遷移至 `yfin`

## Comparisons

- [`comparisons/method-comparison.md`](comparisons/method-comparison.md) — 不同擷取方法 (HTML scrape vs API vs CSV) 的比較與適用情境

## History (歷史歸檔)

歷史 snapshot 集中於 [`history/README.md`](history/README.md)，包含早期 review / audit / release / spec 文件，已不再主動維護。

- `history/audit/` — 過往稽核報告 (`AUDIT_REPORT.md`、`AUDIT_SUMMARY.md`、`FINAL_AUDIT_SUMMARY.md`)
- `history/releases/` — 舊版 release notes 與 release guide
- `history/specs/` — 早期 Yahoo → ampy-proto mapping 規格 (2025-09-18)
- `history/mapping/` — Yahoo 欄位至 `ampy-proto` 欄位對照表
- `history/production-readiness.md` — 早期 production readiness 自我評估
- `history/testing-implementation.md` — 早期測試實作紀錄
- `history/documentation-improvements.md` — 早期 documentation improvement 提案

---

[Back to repo root](../README.md)