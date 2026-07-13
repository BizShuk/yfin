# income

**Type:** `DataFrame` — 損益表（日期為索引）

公司的收益與支出報表，展示企業在特定期間的經營成果。

## DataFrame Structure

| Row | Type | Description |
|-----|------|-------------|
| `Net Income` | `float` | 淨利潤，扣除所有成本與費用後的最終盈餘 |
| `Revenue` | `float` | 總營收，銷售商品或服務的全部收入 |
| `Cost of Revenue` | `float` | 營業成本，直接歸屬於產品/服務的成本 |
| `Gross Profit` | `float` | 毛利，營收減去直接成本 |
| `Operating Expense` | `float` | 營業費用（管理、行銷、研發等） |
| `Operating Income` | `float` | 營業利益，毛利減去營業費用 |
| `Net Interest Income` | `float` | 利息收支淨額，利息收入減去利息支出 |
| `Interest Expense` | `float` | 利息支出，負債的利息費用 |
| `Pretax Income` | `float` | 稅前盈餘，扣除所得稅前的利潤 |
| `Income Tax Expense` | `float` | 所得稅費用，當期所得稅 |
| `Discontinued Operations` | `float` | 停業部門損益，已終止部門的損益 |
| `Minority Interest` | `float` | 少數股權，非全資子公司中其他股东的權益 |
| `Tax Effect Of Unusual Items` | `float` | 非常規項目的稅務影響 |
| `Normalized EBITDA` | `float` | 常規化 EBITDA，排除一次性項目的 EBITDA |
| `Total Unusual Items` | `float` | 特殊項目總計，非經常性損益 |
| `Total Unusual Items Excluding Goodwill` | `float` | 排除商譽的非常規項目 |

## Field Details

### Revenue（營收）
- 公司從核心業務獲得的總收入
- 是衡量企業規模的首要指標

### Gross Profit（毛利）
- `Revenue - Cost of Revenue`
- 反映產品/服務的基本獲利能力

### Operating Income（營業利益）
- `Gross Profit - Operating Expense`
- 反映核心業務的經營績效

### Net Income（淨利）
- `Operating Income + 其他收入/支出 - 稅金`
- 最終歸屬於股東的盈餘

## 頻率類型

| 參數值 | 說明 | 期間 |
|--------|------|------|
| `annual` | 年度財報 | 每年一列 |
| `quarterly` | 季度財報 | 每季一列（預設） |
| `trailing` | 過去12個月 | TTM 數據 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
# 取得近8季財報
quarterly_income = ticker.income
print(quarterly_income.loc['Revenue'])
```

## 獲利能力指標計算

```python
# 毛利率
gross_margin = income.loc['Gross Profit'] / income.loc['Revenue']

# 營業利益率
operating_margin = income.loc['Operating Income'] / income.loc['Revenue']

# 淨利率
net_margin = income.loc['Net Income'] / income.loc['Revenue']
```

## Notes

- Yahoo Finance 的財報資料可能與公司實際公告有時間差
- 「非常規項目」通常包含重組費用、訴訟和解等一次性支出
- 不同公司的損益表欄位可能略有差異
