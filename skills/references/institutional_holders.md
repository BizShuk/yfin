# institutional_holders

**Type:** `DataFrame` — 機構投資人持股資料

記錄所有機構投資人（如共同基金、退休基金、保險公司、對沖基金等）的持股狀況。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Holder` | `str` | 機構名稱（基金公司或投資機構） |
| `Shares` | `int` | 持股數量（股數） |
| `Date Reported` | `datetime` | 13F 報告日期 |
| `% Held` | `float` | 持股比例（%） |
| `Value` | `float` | 持股市值（美元） |

## Field Details

### Holder（機構名稱）
- 機構投資人的完整名稱
- 通常是資產管理公司名稱
- 可能包含基金名稱（如 "Vanguard Total Stock Market Index Fund"）

### Shares（持股數量）
- 該機構持有的股數
- 以「股」為單位，非張數

### Date Reported（報告日期）
- 13F 文件向 SEC 提交的日期
- 美國機構投資人每季結束後45天內需申報

### % Held（持股比例）
- 該機構持股佔總發行股數的百分比
- 所有機構的合計通常為 50-80%

### Value（持股市值）
- 持股數量 × 申報日收盤價
- 以美元計價

## 機構投資人類型

| 類型 | 說明 | 範例 |
|------|------|------|
| 共同基金 | 開放式基金 | Vanguard, BlackRock, Fidelity |
| 指數基金 | 被動追蹤指數 | S&P 500 ETF |
| 退休基金 | 員工退休金 | CalPERS, TSP |
| 保險公司 | 保險公司資產 | MetLife, AXA |
| 對沖基金 | 另類投資 | Bridgewater, Two Sigma |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
inst_holders = ticker.institutional_holders
# 找出持股最多的機構
top_holder = inst_holders.sort_values('% Held', ascending=False).iloc[0]
print(f"{top_holder['Holder']}: {top_holder['% Held']}%")
```

## 重要詮釋

| 指標 | 說明 |
|------|------|
| 機構持股合計 | 所有機構持股總和，>70% 表示高度機構青睞 |
| 新進/退出 | 最近一季新增或減少持股的機構 |
| 持股集中度 | 前10大機構持股佔比 |

## Notes

- 資料基於 SEC 13F 報告，有 45 天延遲
- Yahoo Finance 可能只顯示前 10 或前 20 大機構
- 不同機構的投資目標與期限各異，需分別解讀
- 國外股票（的美股 ADR）可能沒有完整的機構持股資料
