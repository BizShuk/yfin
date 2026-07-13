# recommendations

**Type:** `DataFrame` — 分析師建議列表

記錄各分析師機構對該股票的投資建議歷史。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `period` | `str` | 建議期間/目標年度 |
| `strongBuy` | `int` | 強力買入建議數 |
| `buy` | `int` | 買入建議數 |
| `hold` | `int` | 持有建議數 |
| `sell` | `int` | 賣出建議數 |
| `strongSell` | `int` | 強力賣出建議數 |

## Field Details

### period（期間）
- 通常為年份（如 "2025"）
- 或為季度（如 "Q1 2025"）
- 代表該建議的適用期間或目標年度

### strongBuy（強力買入）
- 分析師給予最高評等的次數
- 通常需要非常強的基本面支持

### buy（買入）
- 正面評價，預期股價上漲
- 常見於成長型股票

### hold（持有）
- 中性評價，不建議買入或賣出
- 常見於藍籌股或穩定型股票

### sell（賣出）
- 負面評價，預期股價下跌
- 可能基於估值過高或基本面惡化

### strongSell（強力賣出）
- 最負面的評等
- 建議立即脫手

## 評級系統對照

| 評級 | 說明 | 預期漲跌幅 |
|------|------|-----------|
| Strong Buy | 強力買入 | > +20% |
| Buy | 買入 | +10% ~ +20% |
| Hold | 持有 | -10% ~ +10% |
| Sell | 賣出 | -10% ~ -20% |
| Strong Sell | 強力賣出 | < -20% |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
recs = ticker.recommendations
# 查看最近的分析師建議
print(recs.tail(10))
```

## 推薦比例計算

```python
# 買入比例 = (Buy + StrongBuy) / 總建議數
total = recs['buy'].sum() + recs['strongBuy'].sum() + recs['hold'].sum() + recs['sell'].sum() + recs['strongSell'].sum()
buy_ratio = (recs['buy'].sum() + recs['strongBuy'].sum()) / total
print(f"Buy Ratio: {buy_ratio:.2%}")
```

## Notes

- 此為彙總資料，非逐筆分析師建議
- 建議數量隨時間變動，新建議會新增覆蓋
- 不同分析師機構的評級標準可能略有差異
- Yahoo Finance 可能只保留近期的建議記錄
