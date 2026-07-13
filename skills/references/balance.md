# balance

**Type:** `DataFrame` — 資產負債表（日期為索引）

記錄企業在特定時點的資產、負債與股東權益狀況。

## DataFrame Structure

| Row | Type | Description |
|-----|------|-------------|
| `Total Assets` | `float` | 總資產，企業擁有或控制的全部資源 |
| `Total Liabilities` | `float` | 總負債，企業的全部義務 |
| `Total Equity` | `float` | 總股東權益，資產減去負債 |
| `Treasury Shares Number` | `int` | 庫藏股股數，公司回購但未註銷的股票 |
| `Ordinary Shares Number` | `int` | 普通股股數，發行在外的普通股數量 |
| `Share Issued` | `int` | 已發行股數，含庫藏股在內的總發行股數 |
| `Total Debt` | `float` | 總負債，包含短期與長期借款 |
| `Tangible Book Value` | `float` | 有形帳面價值，排除無形資產後的淨值 |
| `Intangible Assets` | `float` | 無形資產（專利、商譽、商標等） |
| `Goodwill` | `float` | 商譽，收購時溢價支付的無形資產 |
| `Current Assets` | `float` | 流動資產，預計一年內變現的資產 |
| `Current Liabilities` | `float` | 流動負債，預計一年內到期的負債 |
| `Short Term Debt` | `float` | 短期負債，一年內到期的借款 |
| `Long Term Debt` | `float` | 長期負債，一年後到期的借款 |
| `Cash And Cash Equivalents` | `float` | 現金及約當現金 |
| `Cash Equivalents` | `float` | 現金等價物（國庫券、商業本票等） |
| `Inventory` | `float` | 存貨，用於銷售的商品與原料 |
| `Accounts Receivable` | `float` | 應收帳款，客戶欠款 |
| `Accounts Payable` | `float` | 應付帳款，欠供應商的款項 |
| `Property Plant Equipment` | `float` | 不動產、廠房及設備（PP&E） |

## Field Details

### Total Assets vs Total Liabilities

```
資產 = 負債 + 股動權益
```

此為會計基本恒等式，所有資產必須有對應的資金來源。

### Current Assets / Liabilities（流動資產/負債）

| 項目 | 定義 | 範例 |
|------|------|------|
| 流動資產 | 一年內可變現 | 現金、應收帳款、存貨 |
| 非流動資產 | 一年以上持有 | 房地產、機器設備 |

### Goodwill（商譽）

- 收購價格超過被收購公司淨資產公允價值的部分
- 需要每年進行減值測試
- 反應品牌的無形價值

## 關鍵財務比率

```python
# 負債比率 = 總負債 / 總資產
debt_ratio = balance.loc['Total Liabilities'] / balance.loc['Total Assets']

# 流動比率 = 流動資產 / 流動負債
current_ratio = balance.loc['Current Assets'] / balance.loc['Current Liabilities']

# 每股帳面價值 = 股東權益 / 流通股數
book_value_per_share = balance.loc['Total Equity'] / balance.loc['Ordinary Shares Number']
```

## 頻率類型

| 參數值 | 說明 |
|--------|------|
| `annual` | 年度資產負債表 |
| `quarterly` | 季度資產負債表（預設） |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
balance_sheet = ticker.balance
print(balance_sheet.loc['Total Assets'])
```

## Notes

- 資產負債表是「快照」，只反映特定時點的狀況
- 需與損益表、現金流量表配合分析
- 不同會計準則下的數值可能有所差異
