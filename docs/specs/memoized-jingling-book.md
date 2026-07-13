# docs/ 整理

## Context

`docs/` 目前有 18 個 top-level `.md` + 5 個子目錄，存在重複與不一致：

- 兩個舊版 scraping 大文件（`scrapping.md` 1257 行、`ampy-proto_scrapping.md` 978 行）已被 `docs/scrape/` 子目錄取代但未刪除
- `docs/audit/` 三份 audit 報告（`AUDIT_REPORT`/`AUDIT_SUMMARY`/`FINAL_AUDIT_SUMMARY`）日期均為 2025-01-30，內容互相覆蓋
- `docs/mapping/` 不符 global CLAUDE.md 慣例（specs 應為 `YYYY-MM-DD-<topic>.md`，放 `docs/specs/`）

目標：刪除明顯重複、依 global CLAUDE.md 慣例將 `mapping/` 改為 `specs/`、同步更新 README 引用。

> 用戶確認 scope：**極簡：只刪重複，其他保留** + **改名 docs/specs/**。
> 即不建立 `docs/memory/`、不建立 `docs/README.md` index、不重排 user-facing reference docs。

## Operations

### 1. 刪除重複 scraping 大文件

兩個檔案已被 `docs/scrape/` 完整取代，且無任何內部 cross-reference。

- `rm docs/scrapping.md`（44KB / 1257 行）
- `rm docs/ampy-proto_scrapping.md`（27KB / 978 行）

保留：`docs/scrape/{overview,config,cli,troubleshooting}.md`。

### 2. 精簡 audit 目錄

三份 audit 報告（`AUDIT_REPORT` 308 行、`AUDIT_SUMMARY` 205 行、`FINAL_AUDIT_SUMMARY` 327 行）是同一份 audit 的不同階段；`FINAL_AUDIT_SUMMARY.md` 為最終驗收版，已涵蓋另外兩份的所有結論。

- `rm docs/audit/AUDIT_REPORT.md`
- `rm docs/audit/AUDIT_SUMMARY.md`

保留：`docs/audit/FINAL_AUDIT_SUMMARY.md`（canonical retrospective）。

### 3. README.md 同步

`README.md` L542-544 引用全部 3 個 audit 檔。改為僅引用 FINAL：

```diff
-- **[Audit Report](docs/audit/AUDIT_REPORT.md)** - Comprehensive repository audit findings
-- **[Audit Summary](docs/audit/AUDIT_SUMMARY.md)** - Summary of audit results and fixes
-- **[Final Audit Summary](docs/audit/FINAL_AUDIT_SUMMARY.md)** - Final audit validation
++ **[Final Audit Summary](docs/audit/FINAL_AUDIT_SUMMARY.md)** - Production readiness audit & applied fixes (2025-01-30)
```

### 4. docs/mapping/ → docs/specs/ 改名（global CLAUDE.md 慣例）

`docs/specs/<file>` 應為 `YYYY-MM-DD-<topic>.md`。檔案 `yahoo_to_ampyproto_v1.md` 首次 commit 日期（git log `--diff-filter=A`）為 `2025-09-18`。

- `mv docs/mapping docs/specs`
- `mv docs/specs/yahoo_to_ampyproto_v1.md docs/specs/2025-09-18-yahoo-to-ampyproto.md`

### 5. 影響評估

**Cross-reference 全掃結果**：
- 兩個 scrapping.md → 無引用 ✓
- `docs/mapping/` → 無引用（spec 檔，純閱讀用）✓
- 三份 audit → 僅 `README.md` L542-544 引用，已於步驟 3 更新 ✓

**外部連結**：git remote 未檢查；如有 GitHub README 連結指到 `docs/mapping/`，rename 會壞。建議實作前 `git ls-remote` 與 issues 內 cross-link 抽樣確認。

## 不動的檔案（明確保留）

依「極簡」原則：

| 檔案 | 保留原因 |
| --- | --- |
| `docs/DOCUMENTATION_IMPROVEMENTS.md` | meta-summary 但作為 docs 工作 retrospective 有參考價值 |
| `docs/PRODUCTION_READINESS_REPORT.md` | dated status report |
| `docs/TESTING_IMPLEMENTATION.md` | testing infra retrospective |
| `docs/releases/RELEASE_NOTES.md`（v1.3.0, Dec 2025） | 即使可能過期，仍為正式 release 紀錄 |
| `docs/releases/RELEASE_GUIDE.md` | release 流程指引 |
| 11 個 user-facing reference docs（api-reference / data-structures / examples / error-handling / method-comparison / data-quality / performance / observability / soak-testing / versioning / usage / install / migration-guide） | 公開 API 文件 |
| `docs/tutorials/{onboarding,packages}.md` | 入門指南 |
| `docs/scrape/*` | 當前 scrape 系統文檔 |

## 順手回報（非本任務範圍）

驗證 MIC fix 狀態時發現：`cmd/scrape/scrape_run.go` L501/511/527/537/547/555/563 仍有 7 處硬編 `"XNAS"` 字面值。`FINAL_AUDIT_SUMMARY.md` 聲稱已修，實際僅 `client.go` 套用 `inferMICForSymbol`，CLI scrape fallback 路徑未涵蓋。後續可開 ticket 處理。

## Verification

```bash
# 1. 檔案存在性
ls docs/                          # 不再含 scrapping.md、ampy-proto_scrapping.md
ls docs/audit/                    # 僅 FINAL_AUDIT_SUMMARY.md
ls docs/specs/                    # 2025-09-18-yahoo-to-ampyproto.md

# 2. 連結完整性（README 引用、cross-link）
grep -rn 'docs/mapping\|scrapping\|ampy-proto_scrapping\|AUDIT_REPORT\|AUDIT_SUMMARY' docs/ README.md CLAUDE.md

# 3. 專案仍可建置
go build ./...
go vet ./...
```