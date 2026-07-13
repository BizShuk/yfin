# revenue_estimate

**Type:** `DataFrame` — 營收預估資料

提供分析師對未來營收的詳細預估數據。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Avg` | `float` | 平均/共識預估營收 |
| `Low` | `float` | 最低預估營收 |
| `High` | `float` | 最高預估營收 |
| `Year` | `str` | 財年/季度 |
| `Period` | `str` | 期間（Annual/Quarterly） |

## Field Details

### Avg（共識預估）
- 所有分析師的平均營收預估
- 通常以百萬或十億為單位

### Low（最低預估）
- 最悲觀分析師的預估
- 代表最差情境

### High（最高預估）
- 最樂觀分析師的預估
- 代表最佳情境

### Year（年度）
- 適用的財年
- 例如 "2025", "2026"

### Period（期間類型）

| 值 | 說明 |
|----|------|
| `Annual` | 全年預估 |
| `Quarterly` | 季度預估 |

## 與 EPS 預估的差異

| 特性 | `revenue_estimate` | `earnings_estimate` |
|------|--------------------|--------------------|
| 衡量的 | 營收規模 | 最終利潤 |
| 穩定性 | 通常更穩定 | 受費用結構影響大 |
| 重要性 | 成長股更關注 | 價值投資更關注 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
rev_est = ticker.revenue_estimate
print(rev_est)
```

## 營收分析重點

```python
# 計算營收預估成長率
current_rev = rev_est[rev_est['Year'] == '2025']['Avg'].values[0]
next_rev = rev_est[rev_est['Year'] == '2026']['Avg'].values[0]
growth_rate = (next_rev - current_rev) / current_rev * 100
print(f"Revenue Growth: {growth_rate:.1f}%")
```

## 為何重要

1. **營收 vs 盈餘**：公司可能積極投資導致低盈餘但高營收成長
2. **營收指引**：公司管理層通常對營收預測更有信心
3. **市場份額**：營收成長代表市場地位變化

## Notes

- 營收數據通常以當地貨幣計算
- 跨國公司需注意貨幣匯率影響
- 併購可能影響年對年可比性
