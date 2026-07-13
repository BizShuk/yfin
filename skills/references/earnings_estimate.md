# earnings_estimate

**Type:** `DataFrame` — 盈餘預估資料

提供分析師對未來盈餘的詳細預估數據。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Avg` | `float` | 平均/共識預估 EPS |
| `Low` | `float` | 最低預估 EPS（最悲觀分析師） |
| `High` | `float` | 最高預估 EPS（最樂觀分析師） |
| `Year` | `str` | 財年/季度 |
| `Period` | `str` | 期間（Annual/Quarterly） |

## Field Details

### Avg（共識預估）
- 所有覆蓋該股票的分析師預估平均值
- 華爾街共識是市場最重要的參考點
- 通常被視為「一般預期」

### Low（最低預估）
- 所有分析師中最悲觀的預估
- 代表最差情境假設
- 可能基於保守假設或已知風險

### High（最高預估）
- 所有分析師中最樂觀的預估
- 代表最佳情境假設
- 可能忽視了特定風險

### Year（年度）
- 適用的財年
- 例如 "2025", "2026"

### Period（期間類型）

| 值 | 說明 |
|----|------|
| `Annual` | 全年預估 |
| `Quarterly` | 季度預估 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
est = ticker.earnings_estimate
print(est)
```

## 應用分析

```python
# 計算預估分散度（不確定性指標）
spread = est['High'] - est['Low']
spread_pct = (spread / est['Avg']) * 100
print(f"Estimate Spread: {spread_pct.iloc[0]:.1f}%")

# 高分散度 = 分析師共識低 = 較高風險
# 低分散度 = 共識高 = 確定性較強
```

## 與價格的關係

| 情境 | 預期股價反應 |
|------|-------------|
| 實際 EPS > Avg | 正面，上漲 |
| 實際 EPS < Low | 負面，大跌 |
| 實際 EPS 在 Low-High 區間 | 中性，取決於其他因素 |

## Notes

- 預估通常基於截至目前的已知資訊
- 重大事件可能導致快速修正
- 建議每季更新預估以反映最新共識
