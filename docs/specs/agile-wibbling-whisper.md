# Plan — `docs/` 結構重整與合併

## Context

`docs/` 目前 31 個 markdown 檔，總 13,527 行，主要問題：

1. **`scrapping.md` (45KB, 978L) 與 `scrape/*.md` + `ampy-proto_scrapping.md` 大幅重疊**
    - `scrapping.md` 涵蓋 scrape overview、8 個 endpoint 描述、ampy-proto 整合
    - 但 `scrape/overview.md` 已寫架構、`scrape/cli.md` 已寫 CLI、`ampy-proto_scrapping.md` 已寫 proto 整合
2. **`docs/configuration.md` 缺席**
    - 配置散布於 `install.md`、`usage.md`、`scrape/config.md` 三處（CLI `--config`、`config-effective`、scrape YAML、YAML top-level keys 各自只寫一次）
3. **`docs/errors.md` 缺席**
    - 錯誤處理分屬 `error-handling.md` (700L, 18KB)、`scrape/troubleshooting.md` (432L, 11KB)、`usage.md` 內 error 章節
4. **Tier C（歷史歸檔）零散**
    - `audit/`、`releases/`、`mapping/`、`specs/` 4 個子目錄 + 4 個 top-level 報告檔（`DOCUMENTATION_IMPROVEMENTS.md`、`PRODUCTION_READINESS_REPORT.md`、`TESTING_IMPLEMENTATION.md`、`mapping/yahoo_to_ampyproto_v1.md`、`specs/2025-09-18-yahoo-to-ampyproto.md`），共 ~9 檔歷史內容
5. **Examples 三處**
    - `examples.md` (888L, 24KB)、`api-reference.md` 內 inline 範例、`migration-guide.md` Python→Go 對照範例 — 重疊與否待 verify
6. **Data structures 兩處**
    - `data-structures.md` (498L, 15KB) + `api-reference.md` 內 struct 描述

經 user 在 2026-07-13 確認：「go, try to use traditional chinese by default with english for terminology」。Scope：**只重整結構**，不丟失正確內容；Tier C 整批移入 `history/`。

---

## 目標新結構 (after)

```tree
docs/
├── README.md                            # 新增 — docs 入口與 index
├── getting-started/                     # 新進讀者
│   ├── install.md                       # (原 docs/install.md)
│   ├── onboarding.md                    # (原 docs/tutorials/onboarding.md)
│   └── packages.md                      # (原 docs/tutorials/packages.md)
├── cli/                                 # CLI 操作者
│   ├── usage.md                         # (原 docs/usage.md，扣掉 error 章節 → operations/)
│   ├── commands.md                      # 新增 — 從 usage.md 抽出 CLI 旗標總表
│   └── soak-testing.md                  # (原 docs/soak-testing.md)
├── api/                                 # 整合開發者
│   ├── reference.md                     # (原 docs/api-reference.md，扣掉 data structures 章節 → api/data-structures.md)
│   ├── data-structures.md               # (原 docs/data-structures.md，併入 api-reference.md 的 struct 段落)
│   └── examples.md                      # (原 docs/examples.md)
├── operations/                          # 維運 / SRE
│   ├── configuration.md                 # 新增 — 合併 install/usage/scrape/config 三處配置章節
│   ├── observability.md                 # (原 docs/observability.md)
│   ├── performance.md                   # (原 docs/performance.md)
│   ├── error-handling.md                # (原 docs/error-handling.md + scrape/troubleshooting.md 合併)
│   └── data-quality.md                  # (原 docs/data-quality.md)
├── scrape/                              # 專家 (sub-dir 保留)
│   ├── overview.md                      # (原 docs/scrape/overview.md，併入原 scrapping.md 的 overview 章節)
│   ├── cli.md                           # (原 docs/scrape/cli.md)
│   └── troubleshooting.md               # (原 docs/scrape/troubleshooting.md 併入 operations/error-handling.md 後此檔刪除)
├── integrations/                        # 整合開發者
│   ├── ampy-proto.md                    # (原 docs/ampy-proto_scrapping.md，併入原 scrapping.md 的 proto 整合章節)
│   └── migration-guide.md              # (原 docs/migration-guide.md)
├── comparisons/                         # 整合開發者
│   └── method-comparison.md             # (原 docs/method-comparison.md)
└── history/                             # 歷史歸檔 (Tier C 整批移入)
    ├── README.md                        # 新增 — 說明此目錄為歷史快照，不主動維護
    ├── audit/                           # (原 docs/audit/，原封不動)
    ├── releases/                        # (原 docs/releases/，原封不動)
    ├── specs/                           # (原 docs/specs/，原封不動)
    ├── mapping/                         # (原 docs/mapping/，原封不動)
    ├── production-readiness.md          # (原 docs/PRODUCTION_READINESS_REPORT.md)
    ├── testing-implementation.md        # (原 docs/TESTING_IMPLEMENTATION.md)
    └── documentation-improvements.md    # (原 docs/DOCUMENTATION_IMPROVEMENTS.md)
```

**檔案數變化：31 → 28**（扣 1 個 scrapping.md、加 4 個新目錄入口）。重點不在數量，而在：

- 單一 `configuration.md` 取代 3 處配置
- 單一 `error-handling.md` 取代 2 處錯誤
- 單一 `scrape/overview.md` 取代 `scrapping.md` 重疊
- 歷史歸檔全部進 `history/`，根目錄不再被歷史文件佔位
- docs/README.md 提供完整 index，新進讀者有入口

---

## 合併 / 刪除決策

| 動作 | 來源 | 目標 | 理由 |
| --- | --- | --- | --- |
| **合併** | `install.md` config section | `operations/configuration.md` | 三處重複 |
| **合併** | `usage.md` config section (`yfin config-effective`) | `operations/configuration.md` | 同上 |
| **合併** | `scrape/config.md` 全文 | `operations/configuration.md` | 同上 |
| **合併** | `scrape/troubleshooting.md` | `operations/error-handling.md` | 兩個 error 文件統一 |
| **合併** | `scrapping.md` overview 章節 | `scrape/overview.md` | 兩處架構說明統一 |
| **合併** | `scrapping.md` ampy-proto 章節 | `integrations/ampy-proto.md` | 兩處 proto 整合統一 |
| **合併** | `api-reference.md` 內 struct 段落 | `api/data-structures.md` | 兩處資料結構統一 |
| **刪除** | `scrapping.md` | (合併後無內容) | 重疊消除 |
| **刪除** | `scrape/troubleshooting.md` | (合併後無內容) | 重疊消除 |
| **移動** | `audit/`、`releases/`、`specs/`、`mapping/` | `history/` 子目錄 | Tier C 集中 |
| **移動** | `PRODUCTION_READINESS_REPORT.md`、`TESTING_IMPLEMENTATION.md`、`DOCUMENTATION_IMPROVEMENTS.md` | `history/` | Tier C 集中 |
| **移動** | `tutorials/` | `getting-started/` | 重新分群（讀者） |
| **保留** | `examples.md`、`migration-guide.md`、`method-comparison.md`、`observability.md`、`performance.md`、`data-quality.md`、`soak-testing.md` | (原檔更名/搬位置，內容不動) | 無重疊 |
| **新增** | (空) | `docs/README.md` | docs index |
| **新增** | (空) | `cli/commands.md` | 從 usage.md 抽出旗標總表 |
| **新增** | (空) | `operations/configuration.md` | 合併三處配置 |
| **新增** | (空) | `integrations/ampy-proto.md` | 合併 proto 整合 |
| **新增** | (空) | `history/README.md` | 說明歷史歸檔性質 |

---

## 關鍵待修檔案

### 階段 1 — 純移動（內容不動）

1. `docs/tutorials/onboarding.md` → `docs/getting-started/onboarding.md`
2. `docs/tutorials/packages.md` → `docs/getting-started/packages.md`
3. `docs/audit/*` (3 檔) → `docs/history/audit/*`
4. `docs/releases/*` (2 檔) → `docs/history/releases/*`
5. `docs/specs/*` (1 檔) → `docs/history/specs/*`
6. `docs/mapping/*` (1 檔) → `docs/history/mapping/*`
7. `docs/PRODUCTION_READINESS_REPORT.md` → `docs/history/production-readiness.md`
8. `docs/TESTING_IMPLEMENTATION.md` → `docs/history/testing-implementation.md`
9. `docs/DOCUMENTATION_IMPROVEMENTS.md` → `docs/history/documentation-improvements.md`

並更新 `README.md` 內的相對連結（`docs/audit/AUDIT_REPORT.md` → `docs/history/audit/AUDIT_REPORT.md`）。

### 階段 2 — 合併重疊內容

10. **新增 `docs/operations/configuration.md`**
    - 來源章節：install.md「Configuration」、usage.md「Configuration Management」、scrape/config.md 全文
    - 結構：通用 config（app、yaml loader、effective、CLI flag） → scrape-specific config → bus/fx/observability config
11. **新增 `docs/integrations/ampy-proto.md`**
    - 來源章節：scrapping.md「AMPY-PROTO Integration」章節 + ampy-proto_scrapping.md 全文（去重）
    - 重點：scrape → norm → emit → ampy-proto pipeline、message schema 對應
12. **更新 `docs/scrape/overview.md`**
    - 從 scrapping.md 抽出「Architecture & Data Flow」章節合併進來
    - 刪除原本與 overview 重疊的架構段落
13. **合併 `docs/operations/error-handling.md`**（從 error-handling.md 改名）
    - 把 scrape/troubleshooting.md 的 scrape-specific error scenarios 整合進去
    - 結構：通用 error 類型 → scrape-specific → CLI exit codes → common scenarios
14. **更新 `docs/api/data-structures.md`**（從 data-structures.md 改名）
    - 把 api-reference.md 內的 struct 段落（BarBatch/Quote/CompanyInfo/etc）併入
    - 確保 facade struct 與 norm struct 對齊

### 階段 3 — 刪除被合併的檔案

15. 刪除 `docs/scrapping.md`（內容已進 scrape/overview.md + integrations/ampy-proto.md）
16. 刪除 `docs/scrape/troubleshooting.md`（內容已進 operations/error-handling.md）
17. 刪除 `docs/scrape/config.md`（內容已進 operations/configuration.md）

### 階段 4 — 新增入口文件

18. **`docs/README.md`** — 列出全部檔案與分類，附 audience 標籤
19. **`docs/cli/commands.md`** — 從 usage.md 抽出 10 個 subcommand 的旗標總表（與 `cmd/{admin,dispatch,fundamentals,market,scrape,twse}/*.go` 對齊）
20. **`docs/history/README.md`** — 說明此目錄為歷史快照，不主動維護

### 階段 5 — 更新交叉引用

21. **更新 `CLAUDE.md`**
    - 「docs/* 連結」段落（如有）對齊新位置
22. **更新 `README.md`**
    - 「Documentation」段落全部連結對齊新位置
23. **更新 `docs/tutorials/onboarding.md` 已搬到 `docs/getting-started/onboarding.md`**
    - 內部 `See Also` 連結對齊新位置

---

## 檔案重新配置表 (final)

| 原檔案 | 新位置 | 動作 |
| --- | --- | --- |
| `docs/README.md` | (新增) | 新建 index |
| `docs/install.md` | `docs/getting-started/install.md` | 移動 |
| `docs/tutorials/onboarding.md` | `docs/getting-started/onboarding.md` | 移動 |
| `docs/tutorials/packages.md` | `docs/getting-started/packages.md` | 移動 |
| `docs/usage.md` | `docs/cli/usage.md` | 移動 |
| (新增) | `docs/cli/commands.md` | 從 usage.md 抽出 |
| `docs/soak-testing.md` | `docs/cli/soak-testing.md` | 移動 |
| `docs/api-reference.md` | `docs/api/reference.md` | 移動 + 刪除內嵌 struct 段（已併入 data-structures） |
| `docs/data-structures.md` | `docs/api/data-structures.md` | 移動 + 併入 api-reference 的 struct 段 |
| `docs/examples.md` | `docs/api/examples.md` | 移動 |
| (新增) | `docs/operations/configuration.md` | 合併 install/usage/scrape config |
| `docs/observability.md` | `docs/operations/observability.md` | 移動 |
| `docs/performance.md` | `docs/operations/performance.md` | 移動 |
| `docs/error-handling.md` | `docs/operations/error-handling.md` | 移動 + 併入 scrape/troubleshooting |
| `docs/data-quality.md` | `docs/operations/data-quality.md` | 移動 |
| `docs/scrape/overview.md` | (同位置) | 併入 scrapping.md 的 overview |
| `docs/scrape/cli.md` | (同位置) | 無變動 |
| `docs/scrape/config.md` | **刪除** | 併入 operations/configuration.md |
| `docs/scrape/troubleshooting.md` | **刪除** | 併入 operations/error-handling.md |
| `docs/scrapping.md` | **刪除** | 併入 scrape/overview.md + integrations/ampy-proto.md |
| `docs/ampy-proto_scrapping.md` | `docs/integrations/ampy-proto.md` | 移動 + 併入 scrapping.md proto 章節 |
| `docs/migration-guide.md` | `docs/integrations/migration-guide.md` | 移動 |
| `docs/method-comparison.md` | `docs/comparisons/method-comparison.md` | 移動 |
| `docs/audit/*` | `docs/history/audit/*` | 移動（原封不動） |
| `docs/releases/*` | `docs/history/releases/*` | 移動（原封不動） |
| `docs/specs/*` | `docs/history/specs/*` | 移動（原封不動） |
| `docs/mapping/*` | `docs/history/mapping/*` | 移動（原封不動） |
| `docs/PRODUCTION_READINESS_REPORT.md` | `docs/history/production-readiness.md` | 移動 |
| `docs/TESTING_IMPLEMENTATION.md` | `docs/history/testing-implementation.md` | 移動 |
| `docs/DOCUMENTATION_IMPROVEMENTS.md` | `docs/history/documentation-improvements.md` | 移動 |
| (新增) | `docs/history/README.md` | 說明歷史歸檔 |

**最終檔案數：32**（新增 4：docs/README.md、cli/commands.md、operations/configuration.md、history/README.md；刪除 3：scrapping.md、scrape/config.md、scrape/troubleshooting.md）。

**最終行數估計：~12,500 行**（扣掉 scrapping.md 重疊約 1,000 行；合併後略減）。

---

## 執行策略

### Phase 1 — 純移動（無內容變更）
- 使用 `git mv` 確保 history 保留
- 同步更新 `README.md` 的 Documentation 連結區段（行 496–534）
- 此階段完成後，跑 `make build` 確認沒有 broken link（如果有 link-check 工具）

### Phase 2 — 合併（內容整合）
- 三個合併任務（configuration.md、error-handling.md、ampy-proto.md）可平行
- 每個合併任務由獨立 subagent 處理，確保 atomic

### Phase 3 — 刪除舊檔
- 在 Phase 2 完成確認後才刪，避免遺失

### Phase 4 — 新增入口
- docs/README.md 最後寫（需等 Phase 1-3 確定最終結構）

### 並行化建議

| Phase | Subagent 配置 |
| --- | --- |
| 1 | 1 個 agent：執行 git mv + 更新 README 連結 |
| 2 | 3 個 agent 平行：configuration、error-handling、ampy-proto |
| 3 | 1 個 agent：刪除 scrapping.md、scrape/config.md、scrape/troubleshooting.md |
| 4 | 1 個 agent：寫 docs/README.md、cli/commands.md、history/README.md |
| 5 | 1 個 agent：更新 CLAUDE.md 與各檔內部 See Also |

---

## Verification

### 1. 檔案結構檢查

```bash
find /Users/shuk/projects/yfin/docs -name "*.md" -type f | sort
# 預期：新結構樹狀（見上方 table）
```

### 2. 內容完整性檢查

```bash
# 確認所有原檔內容都已遷移（無遺失）
for f in $(git show HEAD~0 --name-only -- docs/ 2>/dev/null | grep '\.md$'); do
  # 對每個原檔，確認其內容（透過 git grep 找出所有原 heading）仍存在於新檔中
done
```

或人工抽樣：
- `docs/operations/configuration.md` 涵蓋 `--config`、`yfin config-effective`、`scrape:` YAML 區塊、`observability:` YAML 區塊
- `docs/operations/error-handling.md` 涵蓋 exit codes、scrape-specific errors、common scenarios

### 3. 連結檢查

```bash
# 確認沒有 dead link（grep 整個 repo）
grep -rE '\(docs/' /Users/shuk/projects/yfin/ --include='*.md' | grep -v node_modules
# 預期：所有連結對應到新檔案路徑
```

### 4. Tier C 完整性

```bash
# 確認歷史檔案 9 個全部在 history/
ls /Users/shuk/projects/yfin/docs/history/
ls /Users/shuk/projects/yfin/docs/history/{audit,releases,specs,mapping}/
# 預期：原 9 檔全部存在
```

### 5. 最終 grep

```bash
# 確認 docs/ 根目錄沒有 Tier C 散落檔
ls /Users/shuk/projects/yfin/docs/*.md
# 預期：可能為空（因為 docs/README.md 是新增，會在 docs/ 根）
# 或只有 docs/README.md 與 (如果有) 任何尚未分類的檔案
```

---

## Out of scope (本次不做)

- 改寫文件內容（只重整結構；內容正確性由前一輪 sync 保證）
- 把 `docs/scrape/` sub-dir 拆掉（保留 specialist 入口）
- 新建 docs/CONTRIBUTING.md（user 沒要求）
- 把 README.md 從 docs 根搬出（仍放 docs/ 根，符合慣例）
- 對 facade/samples/ 做分類（那是 Go 程式碼，不是文件）
- 對 plans/、monitoring/、skills/ 做任何變動
- 對歷史檔做任何內容修改（原封不動移動）

---

## 結束條件

- 31 檔原始內容全部在新位置找得到
- 沒有 dead link
- Tier C 9 個歷史檔全部在 `docs/history/`
- `docs/scrapping.md`、`docs/scrape/config.md`、`docs/scrape/troubleshooting.md` 已刪除
- `docs/README.md` 提供完整 index
- root `README.md` 與 `CLAUDE.md` 內所有 docs/ 連結對齊新位置