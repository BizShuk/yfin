# options

**Type:** `tuple` — 可用的期權到期日列表

提供該股票所有可交易的期權合約到期日期。

## 基本結構

```python
options = ticker.options
# 回傳格式：(datetime 或 str 的 tuple)
# 例如：('2025-06-20', '2025-06-27', '2025-07-18', ...)
```

## 組成說明

| 組成 | 說明 |
|------|------|
| 到期日 | 期權合約到期的日期 |
| 格式 | 通常為 "YYYY-MM-DD" 字串格式 |

## 期權基礎

### 兩種期權類型

| 類型 | 說明 | 買方權利 |
|------|------|----------|
| Call（買權） | 給予買入股票的權利 | 在到期日前以履約價買入 |
| Put（賣權） | 給予賣出股票的權利 | 在到期日前以履約價賣出 |

### 關鍵術語

| 術語 | 說明 |
|------|------|
| Strike Price（履約價） | 行使期權的價格 |
| Expiration（到期日） | 期權失效的日期 |
| Premium（權利金） | 買方支付的價格 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
expiration_dates = ticker.options
print(f"Available expiration dates: {expiration_dates}")

# 取得特定到期日的期權鏈
if expiration_dates:
    first_exp = expiration_dates[0]
    options_chain = ticker.option_chain(first_exp)
    print(options_chain.calls.head())
```

## 與股價的關係

| 到期日類型 | 說明 |
|------------|------|
| 週契約 | 每週五（美股） |
| 月契約 | 每月第三個週五 |
| 季度契約 | 3、6、9、12月第三週五 |

## Yahoo Finance 期權結構

```python
# 取得期權鏈
chain = ticker.option_chain('2025-06-20')
# 屬性：
# - chain.calls：所有買權
# - chain.puts：所有賣權
# - chain.strikes：所有履約價
```

## Notes

- Yahoo Finance 通常提供近 1-2 年的到期日
- 並非所有股票都有期權交易
- 小型股票的期權流動性可能較低
- 美國期權的到期日為合約月份的第三個週五
