# 變更紀錄 (Changelog)

由 `docs-consolidation` 自 `CLAUDE.md` 與 `README.todo` 搬移彙整，新的在上。

| 日期 (Date) | 變更 (Change) | 來源 (Source) | 原始文件 (Reference) |
| ----------- | ------------- | ------------- | -------------------- |
| 2026-07-13 | docs/\* 同步至現況 (2026-07-13) @2026-07-13_01:26:12 @Completed<br>- CLI 樹對齊 flat subcommand 結構（`yfin pull`、`yfin quote`、`yfin batch`、`yfin config-effective`、`yfin scrape`、`yfin twse` 等）<br>- 套件路徑 `internal/*` → `config/` + `svc/{yahoo,scrape,emit,norm,twse}/` + `utils/{httpx,bus,cache,obsv}/`<br>- `yfinance.NewClient()` → `facade.NewClient()`（CLI 改走 `cmd.CreateClient()`）<br>- 移除 `NewClientWithSessionRotation()`、Docker image、Homebrew tap（皆未發佈）<br>- `yfinance-go` repo 名 → `yfin`（僅 CLAUDE.md 保留「曾用名」標記）<br>- session rotation 全線標記為「已移除」（`--sessions` flag 仍存在但 vestigial）<br>- 移除 cmd/fetch.go 誤刪的 Tier C 檔（scrapping.md、ampy-proto_scrapping.md、audit/、releases/、mapping/ 已 git checkout 還原） | `README.todo` | 無 |
