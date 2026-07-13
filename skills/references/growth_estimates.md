# growth_estimates

**Type:** `DataFrame` — 成長率預估資料

提供分析師對公司關鍵成長指標的預估。

## DataFrame Structure

| Column | Type    | Description           |
| ------ | ------- | --------------------- |
| `0`    | `float` | 今年度成長率預估（%） |
| `+5`   | `float` | 5 年期成長率預估（%） |

## Field Details

### 0（當年成長率）

- 當前財年的營收或 EPS 成長率預估
- 通常是共識平均值
- 以百分比表示

### +5（5 年成長率）

- 未來 5 年的年複合成長率（CAGR）
- 代表分析師對長期成長性的評估
- 通常低於短期預估（成長會放緩）

## 成長率指標類型

| 指標           | 說明           |
| -------------- | -------------- |
| EPS Growth     | 每股收益成長率 |
| Revenue Growth | 營收成長率     |
| EBITDA Growth  | EBITDA 成長率  |

具體指標取決於 Yahoo Finance 的資料來源。

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
growth = ticker.growth_estimates
print(growth)
```

## 解讀要點

```python
# 短期 vs 長期成長
short_term = growth['0'].iloc[0]
long_term = growth['+5'].iloc[0]

# 成長加速度/減速度
# 長期 > 短期：成長加速
# 長期 < 短期：成長放緩
```

## 成長率評估框架

| 成長率區間 | 評價        |
| ---------- | ----------- |
| > 25%      | 高成長股    |
| 15-25%     | 穩健成長    |
| 5-15%      | 溫和成長    |
| < 5%       | 成熟/價值股 |

## 與估值的關係

```python
# PEG Ratio = P/E / Growth Rate
# PEG < 1 = 股價相對便宜
# PEG > 1.5 = 股價相對昂貴

pe = ticker.info['trailingPE']
growth_rate = growth['0'].iloc[0]
peg = pe / growth_rate
print(f"PEG Ratio: {peg:.2f}")
```

## Notes

- 並非所有股票都有 5 年預估
- 不同分析師的假設可能差異很大
- 產業平均水準是重要的比較基準
