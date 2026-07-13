# sec_filings

**Type:** `DataFrame` — SEC 申報文件資料

提供公司向美國證券交易委員會（SEC）提交的重要官方文件。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Date` | `datetime` | 申報日期 |
| `Type` | `str` | 文件類型 |
| `Description` | `str` | 文件描述 |
| `Link` | `str` | SEC EDGAR 連結 |

## Field Details

### Date（日期）
- 文件提交或生效的日期
- 通常為美國交易日時段

### Type（類型）

| 類型 | 說明 | 頻率 |
|------|------|------|
| 10-K | 年度報告 | 每年 |
| 10-Q | 季度報告 | 每季 |
| 8-K | 重大事件報告 | 事件驅動 |
| DEF 14A | 委託書聲明 | 年度會議前 |
| S-1 | 首次公開發行 | 上市前 |
| 4 | 內部人交易 | 交易後2日內 |

### Description（描述）
- 對文件內容的簡要說明
- 可幫助快速理解文件性質

### Link（連結）
- SEC EDGAR 系統的直接連結
- 可免費下載完整 PDF 文件

## 常見 SEC 文件

| 文件 | 頻率 | 主要內容 |
|------|------|----------|
| 10-K | 年報 | 完整年度財務表現 |
| 10-Q | 季報 | 未審計季度財務 |
| 8-K | 不定期 | 重大事件即時披露 |
| 13F | 季報 | 機構持股變動 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
filings = ticker.sec_filings
# 查看近期的 8-K 文件
recent_8k = filings[filings['Type'] == '8-K'].tail(5)
print(recent_8k)
```

## 投資研究價值

1. **10-K/10-Q**：最全面的財務資訊來源
2. **8-K**：了解重大變化的第一手資料
3. **14A**：了解公司治理和薪酬計劃
4. **4**：追蹤內部人交易

## Notes

- SEC 文件僅適用於美國上市公司
- 外國公司（如 ADRs）可能使用不同的申報標準
- EDGAR 連結可能因格式變更而失效
- 建議直接訪問 SEC EDGAR 網站獲取最新資訊
