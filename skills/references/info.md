# info

**Type:** `dict` — Key-value pairs of ticker metadata

取得個股的基本資訊、財務數據、市場數據、與分析師建議等綜合資料。

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `address1` | `str` | 主要地址欄位，公司登記地址的第一行 |
| `city` | `str` | 公司所在城市 |
| `state` | `str` | 州/省名稱 |
| `zip` | `str` | 郵遞區號 |
| `country` | `str` | 國家 |
| `phone` | `str` | 聯絡電話 |
| `website` | `str` | 公司官方網站 URL |
| `industry` | `str` | 所屬產業名稱（如 "Semiconductors"） |
| `sector` | `str` | 所屬部門名稱（如 "Technology"） |
| `longBusinessSummary` | `str` | 詳細的公司業務描述摘要 |
| `fullTimeEmployees` | `int` | 全職員工總數 |
| `companyOfficers` | `list[dict]` | 公司高階主管與董事名單 |
| `maxAge` | `int` | 資料年齡限制（秒），超過此時間的資料視為過期 |
| `priceHint` | `int` | 價格報價精確度，用於小數位數 |
| `previousClose` | `float` | 前一交易日收盤價 |
| `open` | `float` | 今日開盤價 |
| `dayLow` | `float` | 今日最低價 |
| `dayHigh` | `float` | 今日最高價 |
| `regularMarketPreviousClose` | `float` | 正常交易時段的前一日收盤價 |
| `regularMarketOpen` | `float` | 正常交易時段的開盤價 |
| `regularMarketDayLow` | `float` | 正常交易時段的日最低價 |
| `regularMarketDayHigh` | `float` | 正常交易時段的日最高價 |
| `dividendRate` | `float` | 年度股息率（每股股息金額） |
| `dividendYield` | `float` | 股息殖利率（股息/股價） |
| `exDividendDate` | `datetime` | 除息日，過此日期購入無法參與本次股息 |
| `payoutRatio` | `float` | 股息支付率（股息/盈餘） |
| `beta` | `float` | Beta 係數，與大盤的相關性指標 |
| `trailingPE` | `float` | 追蹤本益比（股價/近12個月EPS） |
| `forwardPE` | `float` | 預估本益比（股價/未來12個月EPS） |
| `volume` | `int` | 今日成交量（含盤前盤後） |
| `regularMarketVolume` | `int` | 正常交易時段成交量 |
| `averageVolume` | `int` | 平均日成交量（3個月） |
| `averageVolume10days` | `int` | 10日平均成交量 |
| `marketCap` | `int` | 總市值（股價×流通股數） |
| `fiftyTwoWeekHigh` | `float` | 52週最高價 |
| `fiftyTwoWeekLow` | `float` | 52週最低價 |
| `fiftyDayAverage` | `float` | 50日移動平均線 |
| `twoHundredDayAverage` | `float` | 200日移動平均線 |
| `trailingAnnualDividendRate` | `float` | 近12個月股息率 |
| `trailingAnnualDividendYield` | `float` | 近12個月股息殖利率 |
| `currency` | `str` | 交易貨幣（如 "USD", "TWD"） |
| `enterpriseValue` | `int` | 企業價值（市值 + 總債務 - 現金） |
| `profitMargins` | `float` | 淨利率 |
| `floatShares` | `int` | 流通股數（可供交易的股票數量） |
| `sharesOutstanding` | `int` | 發行在外流通股數 |
| `sharesShort` | `int` | 融券餘額（放空數量） |
| `sharesPercentOutInsiders` | `float` | 內部人持股比例（%） |
| `sharesPercentInstitutions` | `float` | 機構投資人持股比例（%） |
| `bookValue` | `float` | 每股帳面價值 |
| `priceToBook` | `float` | 股價淨值比（P/B） |
| `earningsGrowth` | `float` | 盈餘成長率（YoY） |
| `revenueGrowth` | `float` | 營收成長率（YoY） |
| `grossMargins` | `float` | 毛利率 |
| `ebitdaMargins` | `float` | EBITDA 毛利率 |
| `operatingMargins` | `float` | 營業利益率 |
| `returnOnAssets` | `float` | 資產報酬率（ROA） |
| `returnOnEquity` | `float` | 股東權益報酬率（ROE） |
| `totalRevenue` | `int` | 總營收 |
| `netIncomeToCommon` | `int` | 普通股淨利 |
| `operatingCashflow` | `int` | 營業現金流量 |
| `freeCashflow` | `int` | 自由現金流量 |
| `currentPrice` | `float` | 目前股價 |
| `targetHighPrice` | `float` | 分析師目標最高價 |
| `targetLowPrice` | `float` | 分析師目標最低價 |
| `targetMeanPrice` | `float` | 分析師目標均價 |
| `targetMedianPrice` | `float` | 分析師目標中位數價 |
| `recommendationKey` | `str` | 綜合建議（"strongBuy", "buy", "hold", "sell", "strongSell"） |
| `numberOfAnalystOpinions` | `int` | 評估的分析師數量 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("2330.TW")
info = ticker.info
print(info['currentPrice'], info['currency'])
```

## Notes

- 並非所有欄位都會有值，某些可選欄位可能為 `None`
- 不同股票可能會有不同的欄位子集，取決於資料完整性
