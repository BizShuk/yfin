# actions

**Type:** `DataFrame` — 除權息與股票分割資料（日期為索引）

記錄個股所有歷史上的除權息事件與股票分割事件。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Dividends` | `float` | 每股股息金額（僅在除息日有值） |
| `Stock Splits` | `float` | 分割比例（1.0 = 無分割） |

## Field Details

### Dividends（股息）
- 記錄每次除息事件發放的每股股息金額
- 大多數日期該欄位為 0
- 日期索引代表「除息日」即將除息的日期

### Stock Splits（股票分割）
- 記錄每次股票分割事件的比例
- 例如：2.0 表示 1 股分割成 2 股
- 0.5 表示 2 股合併為 1 股（反向分割）
- 1.0 表示無分割事件

## 常見除息日術語

| 術語 | 說明 |
|------|------|
| 宣布日 | 公司宣布發放股息的日期 |
| 除息日 | 過此日期買入的投資者無法參與本次股息 |
| 支付日 | 股息實際匯入帳戶的日期 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
actions = ticker.actions
# 找出所有分割事件
splits = actions[actions['Stock Splits'] != 0]
print(splits)
```

## 與 history 的差異

| 特性 | `actions` | `history` |
|------|-----------|-----------|
| 時間範圍 | 所有歷史 | 指定期間 |
| 用途 | 查詢所有除權息事件 | 取得實際價格資料 |
| 頻率 | 通常稀疏 | 每個交易日一列 |

## Notes

- `actions` 可視為 `history` 中 `Dividends` 與 `Stock Splits` 欄位的完整歷史版本
- 查詢整個歷史時，請注意 Yahoo Finance 可能限制最大查詢範圍
