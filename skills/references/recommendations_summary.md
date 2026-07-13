# recommendations_summary

**Type:** `DataFrame` — 分析師建議摘要

提供分析師建議的彙總統計數據。

## DataFrame Structure

欄位結構較為簡潔，與 `recommendations` 的彙總版本類似，具體欄位依資料可用性而定。

## 預期欄位

| Column | Type | Description |
|--------|------|-------------|
| `period` | `str` | 期間標識 |
| 摘要欄位 | `dict` | 包含各評級的計數或比例 |

## Field Details

此為 `recommendations` 的衍生資料，專注於提供更簡潔的彙總視圖。

## 與 recommendations 的差異

| 特性 | `recommendations` | `recommendations_summary` |
|------|---------------------|--------------------------|
| 粒度 | 按分析師/時間的逐筆建議 | 彙總統計 |
| 歷史 | 完整的建議歷史 | 可能只有最新或摘要 |
| 用途 | 追蹤建議變化 | 快速查看整體氛圍 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
summary = ticker.recommendations_summary
print(summary)
```

## 如何解讀摘要

```python
# 計算建議分數（越高越正面）
# 5 = StrongBuy, 4 = Buy, 3 = Hold, 2 = Sell, 1 = StrongSell
score = summary['strongBuy'] * 5 + summary['buy'] * 4 + summary['hold'] * 3 + summary['sell'] * 2 + summary['strongSell'] * 1
total = summary['strongBuy'] + summary['buy'] + summary['hold'] + summary['sell'] + summary['strongSell']
avg_score = score / total
print(f"Recommendation Score: {avg_score:.2f}")
```

## Notes

- 並非所有股票都有此資料
- Yahoo Finance 有時會動態決定可用欄位
- 建議同時參考 `recommendations` 取得更完整的歷史
