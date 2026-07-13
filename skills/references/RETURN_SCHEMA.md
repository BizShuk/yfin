# yf.py Return Schema Reference

## info

**Type:** `dict` — Key-value pairs of ticker metadata

| Field                         | Description                   |
| ----------------------------- | ----------------------------- |
| `address1`                    | Primary address line          |
| `city`                        | City                          |
| `state`                       | State/Province                |
| `zip`                         | Postal code                   |
| `country`                     | Country                       |
| `phone`                       | Contact phone                 |
| `website`                     | Company website               |
| `industry`                    | Industry name                 |
| `sector`                      | Sector name                   |
| `longBusinessSummary`         | Business description          |
| `fullTimeEmployees`           | Employee count                |
| `companyOfficers`             | Executive team list           |
| `maxAge`                      | Data age limit                |
| `priceHint`                   | Price quote precision         |
| `previousClose`               | Previous closing price        |
| `open`                        | Opening price                 |
| `dayLow`                      | Day low price                 |
| `dayHigh`                     | Day high price                |
| `regularMarketPreviousClose`  | Regular market previous close |
| `regularMarketOpen`           | Regular market open           |
| `regularMarketDayLow`         | Regular market day low        |
| `regularMarketDayHigh`        | Regular market day high       |
| `dividendRate`                | Annual dividend rate          |
| `dividendYield`               | Dividend yield                |
| `exDividendDate`              | Ex-dividend date              |
| `payoutRatio`                 | Dividend payout ratio         |
| `beta`                        | Beta coefficient              |
| `trailingPE`                  | Trailing P/E ratio            |
| `forwardPE`                   | Forward P/E ratio             |
| `volume`                      | Trading volume                |
| `regularMarketVolume`         | Regular market volume         |
| `averageVolume`               | Average volume                |
| `averageVolume10days`         | 10-day average volume         |
| `marketCap`                   | Market capitalization         |
| `fiftyTwoWeekHigh`            | 52-week high                  |
| `fiftyTwoWeekLow`             | 52-week low                   |
| `fiftyDayAverage`             | 50-day moving average         |
| `twoHundredDayAverage`        | 200-day moving average        |
| `trailingAnnualDividendRate`  | Trailing annual dividend      |
| `trailingAnnualDividendYield` | Trailing dividend yield       |
| `currency`                    | Trading currency              |
| `enterpriseValue`             | Enterprise value              |
| `profitMargins`               | Profit margins                |
| `floatShares`                 | Float shares                  |
| `sharesOutstanding`           | Shares outstanding            |
| `sharesShort`                 | Shares short                  |
| `sharesPercentOutInsiders`    | Insider ownership %           |
| `sharesPercentInstitutions`   | Institutional ownership %     |
| `bookValue`                   | Book value per share          |
| `priceToBook`                 | Price to book ratio           |
| `earningsGrowth`              | Earnings growth               |
| `revenueGrowth`               | Revenue growth                |
| `grossMargins`                | Gross margins                 |
| `ebitdaMargins`               | EBITDA margins                |
| `operatingMargins`            | Operating margins             |
| `returnOnAssets`              | Return on assets              |
| `returnOnEquity`              | Return on equity              |
| `totalRevenue`                | Total revenue                 |
| `netIncomeToCommon`           | Net income to common          |
| `operatingCashflow`           | Operating cash flow           |
| `freeCashflow`                | Free cash flow                |
| `currentPrice`                | Current price                 |
| `targetHighPrice`             | Analyst target high           |
| `targetLowPrice`              | Analyst target low            |
| `targetMeanPrice`             | Analyst target mean           |
| `targetMedianPrice`           | Analyst target median         |
| `recommendationKey`           | Buy/Hold/Sell recommendation  |
| `numberOfAnalystOpinions`     | Number of analysts            |

---

## history

**Type:** `DataFrame` — Historical price data

| Column         | Description        |
| -------------- | ------------------ |
| `Open`         | Opening price      |
| `High`         | Highest price      |
| `Low`          | Lowest price       |
| `Close`        | Closing price      |
| `Volume`       | Trading volume     |
| `Dividends`    | Dividend payments  |
| `Stock Splits` | Stock split events |

---

## actions

**Type:** `DataFrame` — Dividends and splits

| Column         | Description                   |
| -------------- | ----------------------------- |
| `Dividends`    | Dividend payments (per share) |
| `Stock Splits` | Stock split ratio             |

---

## income

**Type:** `DataFrame` — Income statement (quarterly/yearly/trailing)

| Row                                      | Description               |
| ---------------------------------------- | ------------------------- |
| `Net Income`                             | Net income                |
| `Revenue`                                | Total revenue             |
| `Cost of Revenue`                        | Cost of goods sold        |
| `Gross Profit`                           | Gross profit              |
| `Operating Expense`                      | Operating expenses        |
| `Operating Income`                       | Operating profit          |
| `Net Interest Income`                    | Interest income/expense   |
| `Interest Expense`                       | Interest expense          |
| `Pretax Income`                          | Income before tax         |
| `Income Tax Expense`                     | Tax provision             |
| `Discontinued Operations`                | Discontinued ops          |
| `Minority Interest`                      | Minority interest         |
| `Tax Effect Of Unusual Items`            | Tax effect                |
| `Normalized EBITDA`                      | Normalized EBITDA         |
| `Total Unusual Items`                    | Unusual items             |
| `Total Unusual Items Excluding Goodwill` | Unusual items ex-goodwill |

---

## balance

**Type:** `DataFrame` — Balance sheet (quarterly/yearly)

| Row                         | Description         |
| --------------------------- | ------------------- |
| `Total Assets`              | Total assets        |
| `Total Liabilities`         | Total liabilities   |
| `Total Equity`              | Total equity        |
| `Treasury Shares Number`    | Treasury shares     |
| `Ordinary Shares Number`    | Ordinary shares     |
| `Share Issued`              | Shares issued       |
| `Total Debt`                | Total debt          |
| `Tangible Book Value`       | Tangible book value |
| `Intangible Assets`         | Intangible assets   |
| `Goodwill`                  | Goodwill            |
| `Current Assets`            | Current assets      |
| `Current Liabilities`       | Current liabilities |
| `Short Term Debt`           | Short-term debt     |
| `Long Term Debt`            | Long-term debt      |
| `Cash And Cash Equivalents` | Cash                |
| `Cash Equivalents`          | Cash equivalents    |
| `Inventory`                 | Inventory           |
| `Accounts Receivable`       | Accounts receivable |
| `Accounts Payable`          | Accounts payable    |
| `Property Plant Equipment`  | PP&E                |

---

## cashflow

**Type:** `DataFrame` — Cash flow statement (quarterly/yearly/trailing)

| Row                           | Description          |
| ----------------------------- | -------------------- |
| `Free Cash Flow`              | Free cash flow       |
| `Operating Cash Flow`         | Operating cash flow  |
| `Capital Expenditure`         | Capex                |
| `Repurchase Of Capital Stock` | Share buyback        |
| `Repayment Of Debt`           | Debt repayment       |
| `Issuance Of Debt`            | Debt issuance        |
| `Interest Paid`               | Interest paid        |
| `Dividends Paid`              | Dividends paid       |
| `Cash Flow From Financing`    | Financing cash flow  |
| `Cash Flow From Investing`    | Investing cash flow  |
| `Cash Flow From Operations`   | Operations cash flow |
| `Change In Cash`              | Net cash change      |
| `Net Income`                  | Net income           |
| `Depreciation`                | Depreciation         |
| `Stock Based Compensation`    | SBC                  |

---

## major_holders

**Type:** `DataFrame` — Major shareholders

| Column          | Description  |
| --------------- | ------------ |
| `Holder`        | Holder name  |
| `Shares`        | Shares held  |
| `Date Reported` | Report date  |
| `% Held`        | Ownership %  |
| `Value`         | Market value |

---

## institutional_holders

**Type:** `DataFrame` — Institutional investors

| Column          | Description      |
| --------------- | ---------------- |
| `Holder`        | Institution name |
| `Shares`        | Shares held      |
| `Date Reported` | Report date      |
| `% Held`        | Ownership %      |
| `Value`         | Market value     |

---

## mutualfund_holders

**Type:** `DataFrame` — Mutual fund holders

| Column          | Description  |
| --------------- | ------------ |
| `Holder`        | Fund name    |
| `Shares`        | Shares held  |
| `Date Reported` | Report date  |
| `% Held`        | Ownership %  |
| `Value`         | Market value |

---

## insider_transactions

**Type:** `DataFrame` — Insider transactions

| Column         | Description       |
| -------------- | ----------------- |
| `Name`         | Insider name      |
| `Transaction`  | Transaction type  |
| `Shares`       | Shares traded     |
| `Date`         | Transaction date  |
| `Value`        | Transaction value |
| `Shares Total` | Total shares held |

---

## insider_purchases

**Type:** `DataFrame` — Insider purchases

| Column        | Description       |
| ------------- | ----------------- |
| `Name`        | Insider name      |
| `Transaction` | Transaction type  |
| `Shares`      | Shares purchased  |
| `Date`        | Transaction date  |
| `Value`       | Transaction value |

---

## insider_roster

**Type:** `DataFrame` — Insider roster

| Column    | Description  |
| --------- | ------------ |
| `Name`    | Insider name |
| `Title`   | Job title    |
| `火烧`    | Reports to   |
| `Insider` | Is insider   |

---

## recommendations

**Type:** `DataFrame` — Analyst recommendations

| Column       | Description       |
| ------------ | ----------------- |
| `period`     | Period            |
| `strongBuy`  | Strong buy count  |
| `buy`        | Buy count         |
| `hold`       | Hold count        |
| `sell`       | Sell count        |
| `strongSell` | Strong sell count |

---

## recommendations_summary

**Type:** `DataFrame` — Recommendations summary

---

## upgrades

**Type:** `DataFrame` — Upgrades/downgrades

| Column   | Description       |
| -------- | ----------------- |
| `Grade`  | Analyst grade     |
| `From`   | Previous rating   |
| `To`     | New rating        |
| `Action` | Upgrade/Downgrade |
| `Date`   | Announcement date |
| `Firm`   | Analyst firm      |

---

## earnings_dates

**Type:** `DataFrame` — Earnings announcement dates

| Column          | Description           |
| --------------- | --------------------- |
| `Earnings Date` | Announcement datetime |
| `EPS Estimate`  | EPS estimate          |
| `EPS Actual`    | EPS actual            |
| `Difference`    | Beat/Miss             |
| `Surprise %`    | Surprise percentage   |

---

## earnings_history

**Type:** `DataFrame` — Earnings history

---

## eps_trend

**Type:** `DataFrame` — EPS trend

| Column | Description      |
| ------ | ---------------- |
| `0`    | Current quarter  |
| `+7D`  | 7 days forward   |
| `+30D` | 30 days forward  |
| `-30D` | 30 days backward |
| `-7D`  | 7 days backward  |

---

## eps_revisions

**Type:** `DataFrame` — EPS revisions

---

## earnings_estimate

**Type:** `DataFrame` — Earnings estimates

| Column   | Description      |
| -------- | ---------------- |
| `Avg`    | Average estimate |
| `Low`    | Low estimate     |
| `High`   | High estimate    |
| `Year`   | Fiscal year      |
| `Period` | Quarter/Year     |

---

## revenue_estimate

**Type:** `DataFrame` — Revenue estimates

| Column   | Description      |
| -------- | ---------------- |
| `Avg`    | Average estimate |
| `Low`    | Low estimate     |
| `High`   | High estimate    |
| `Year`   | Fiscal year      |
| `Period` | Quarter/Year     |

---

## growth_estimates

**Type:** `DataFrame` — Growth estimates

| Column | Description         |
| ------ | ------------------- |
| `0`    | Current year growth |
| `+5`   | 5-year growth       |

---

## price_targets

**Type:** `DataFrame` — Analyst price targets

| Column    | Description   |
| --------- | ------------- |
| `Current` | Current price |
| `Mean`    | Mean target   |
| `High`    | High target   |
| `Low`     | Low target    |

---

## news

**Type:** `list[dict]` — News articles

| Key       | Description      |
| --------- | ---------------- |
| `id`      | Article ID       |
| `content` | Article content  |
| `title`   | Article title    |
| `pubDate` | Publication date |
| `link`    | Article URL      |
| `source`  | News source      |

---

## calendar

**Type:** `DataFrame` — Calendar events

---

## sec_filings

**Type:** `DataFrame` — SEC filings

| Column        | Description |
| ------------- | ----------- |
| `Date`        | Filing date |
| `Type`        | Filing type |
| `Description` | Description |
| `Link`        | Filing URL  |

---

## sustainability

**Type:** `DataFrame` — ESG/sustainability scores

---

## isin

**Type:** `str` — International Securities Identification Number

---

## options

**Type:** `tuple` — Available option expiration dates

---

## metadata

**Type:** `dict` — History data metadata

| Key                    | Description         |
| ---------------------- | ------------------- |
| `symbol`               | Ticker symbol       |
| `currency`             | Trading currency    |
| `exchange`             | Exchange name       |
| `quoteType`            | Asset type          |
| `firstTradeDate`       | First trade date    |
| `regularMarketTime`    | Market time         |
| `hasPrePostMarketData` | Has pre/post market |
| `timezone`             | Exchange timezone   |
