# price_targets

**Type:** `DataFrame` — 分析師目標價資料

提供分析師對股票未來價格的預測區間。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Current` | `float` | 目前股價 |
| `Mean` | `float` | 平均目標價（共識） |
| `High` | `float` | 最高目標價 |
| `Low` | `float` | 最低目標價 |

## Field Details

### Current（目前股價）
- 最近的成交價格
- 作為計算潛在漲跌幅的基準

### Mean（平均目標價）
- 所有分析師目標價的平均值
- 華爾街共識目標價
- 最重要的參考指標

### High（最高目標價）
- 最樂觀分析師的目標價
- 代表最佳情境
- 通常隱含 > 30% 的潛在上漲

### Low（最低目標價）
- 最悲觀分析師的目標價
- 代表最差情境
- 可能低於目前股價

## 潛在漲跌幅計算

```python
# 共識上漲潛力
upside = (price_targets['Mean'] - price_targets['Current']) / price_targets['Current'] * 100
print(f"Consensus Upside: {upside.iloc[0]:.1f}%")

# 區間上漲潛力
high_upside = (price_targets['High'] - price_targets['Current']) / price_targets['Current'] * 100
low_upside = (price_targets['Low'] - price_targets['Current']) / price_targets['Current'] * 100
print(f"High Upside: {high_upside.iloc[0]:.1f}%")
print(f"Low Upside: {low_upside.iloc[0]:.1f}%")
```

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
targets = ticker.price_targets
print(targets)
```

## 解讀框架

| 目標價 vs 股價 | 解讀 |
|----------------|------|
| Mean >> Current | 強烈買進訊號 |
| Mean > Current | 温和買進 |
| Mean ≈ Current | 中性 |
| Mean < Current | 賣出訊號 |

## 目標價的局限性

1. **時間不確定**：沒有明確的達成時間表
2. **假設簡化**：忽略公司重大變化
3. **群眾心理**：目標價趨向收斂而非領先

## Notes

- 目標價通常有 12-18 個月的時間範圍
- 建議同時關注評級與目標價的變化
- 不同機構的目標價可能採用不同時間點
