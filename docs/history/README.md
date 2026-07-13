# History (歷史歸檔)

本目錄收納 `yfin` 早期 review / audit / release / spec 工作的歷史 snapshot。這些文件僅作為當時狀態的紀錄，**不再主動維護** — 若與現行實作衝突，請以 `docs/` 下對應主題的最新文件為準。

## 子目錄

- [`audit/`](audit/) — 過往稽核報告，包含 `AUDIT_REPORT.md`、`AUDIT_SUMMARY.md`、`FINAL_AUDIT_SUMMARY.md`，記錄架構、錯誤處理、測試覆蓋率等面向的稽核發現
- [`releases/`](releases/) — 舊版 release notes 與 release guide，記錄發行流程與版本歷史
- [`specs/`](specs/) — 早期 Yahoo Finance → ampy-proto mapping 規格 (2025-09-18 版)，為欄位對應的初版設計
- [`mapping/`](mapping/) — Yahoo 原始欄位至 `ampy-proto` 欄位的對照表，作為資料轉換的參考依據

## 頂層歷史報告

- [`production-readiness.md`](production-readiness.md) — 早期 production readiness 自我評估，列出上線前須完成項目
- [`testing-implementation.md`](testing-implementation.md) — 早期測試實作紀錄，記錄測試策略、覆蓋率目標與 mock 設計
- [`documentation-improvements.md`](documentation-improvements.md) — 早期 documentation improvement 提案，作為本次 documentation 結構重整的前置參考

---

- [Back to docs index](../README.md)
- [Back to repo root](../../README.md)