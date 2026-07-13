# cashflow

**Type:** `DataFrame` — 現金流量表（日期為索引）

記錄企業一段時間內現金的流入與流出，分為營業、投資、與融資三大活動。

## DataFrame Structure

| Row                           | Type    | Description                              |
| ----------------------------- | ------- | ---------------------------------------- |
| `Free Cash Flow`              | `float` | 自由現金流量，營運產生的現金減去資本支出 |
| `Operating Cash Flow`         | `float` | 營運現金流量，日常業務產生的現金         |
| `Capital Expenditure`         | `float` | 資本支出，購置固定資產的現金支出         |
| `Repurchase Of Capital Stock` | `float` | 庫藏股買回，公司回購自家股票             |
| `Repayment Of Debt`           | `float` | 債務還本，償還借款的本金                 |
| `Issuance Of Debt`            | `float` | 債務發行，新增借款籌資                   |
| `Interest Paid`               | `float` | 支付利息，債務的利息費用現金支出         |
| `Dividends Paid`              | `float` | 支付股息，現金股息發放                   |
| `Cash Flow From Financing`    | `float` | 融資現金流量，借款、發行股票、股息等     |
| `Cash Flow From Investing`    | `float` | 投資現金流量，收購、處分資產             |
| `Cash Flow From Operations`   | `float` | 營運現金流量，核心業務的現金             |
| `Change In Cash`              | `float` | 現金變動淨額，期末與期初現金差額         |
| `Net Income`                  | `float` | 淨利潤，來自損益表                       |
| `Depreciation`                | `float` | 折舊費用，非現金支出                     |
| `Stock Based Compensation`    | `float` | 股票報酬，給員工的股票期權價值           |

## Field Details

### Free Cash Flow（自由現金流量）

```
Free Cash Flow = Operating Cash Flow - Capital Expenditure
```

- 衡量企業產生現金的能力
- 排除資本支出後可自由運用的現金
- 是評估企業價值的重要指標

### Operating Cash Flow（營運現金流量）

- 來自核心業務的現金
- 理想情況應為正數且穩定
- 與淨利潤的差異反映會計調整

### Capital Expenditure（資本支出）

- 購買長期資產的現金支出
- 例如：設備、房地產、軟體
- 是維持與擴張業務的必要投資

## 現金流量表的三大活動

| 活動     | 說明     | 範例                         |
| -------- | -------- | ---------------------------- |
| 營運活動 | 核心業務 | 銷售商品、支付供應商、發薪水 |
| 投資活動 | 長期資產 | 購買設備、處分子公司         |
| 融資活動 | 資本結構 | 借款、還款、發行股票、發股息 |

## 頻率類型

| 參數值      | 說明              |
| ----------- | ----------------- |
| `annual`    | 年度現金流量表    |
| `quarterly` | 季度現金流量表    |
| `trailing`  | 過去12個月（TTM） |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
cf = ticker.cashflow
print(cf.loc['Free Cash Flow'])
```

## 現金流量品質分析

```python
# 盈餘品質比率 = 營運現金流量 / 淨利潤
# > 1 表示盈餘品質佳
quality_ratio = cf.loc['Cash Flow From Operations'] / cf.loc['Net Income']
```

## Notes

- 現金流量表以現金為基礎，不受會計方法影響
- 淨利潤與營運現金流量的差異源於非現金項目（折舊、應收應付變動等）
- 「非常規項目」如重組費用在現金流量表中會被加回
