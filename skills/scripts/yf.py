# -*- coding: utf-8 -*-

import json
from datetime import datetime, timedelta, timezone
import pandas as pd
import yfinance as yf

DEFAULT_TICKER = "2330.TW"

def format_data_to_json(data) -> str:
    """Helper to format data as JSON string instead of printing."""
    if data is None:
        return json.dumps({}, indent=2)

    if isinstance(data, pd.DataFrame):
        if data.empty:
            return json.dumps({}, indent=2)
        else:
            data_copy = data.copy()
            if isinstance(data_copy.index, pd.DatetimeIndex):
                data_copy.index = data_copy.index.strftime('%Y-%m-%d %H:%M:%S')
            if isinstance(data_copy.columns, pd.DatetimeIndex):
                data_copy.columns = data_copy.columns.strftime('%Y-%m-%d %H:%M:%S')
            data_copy = data_copy.astype(object).where(pd.notnull(data_copy), None)
            return json.dumps(data_copy.to_dict(), indent=2, default=str)
    elif isinstance(data, pd.Series):
        if data.empty:
            return json.dumps({}, indent=2)
        else:
            data_copy = data.copy()
            if isinstance(data_copy.index, pd.DatetimeIndex):
                data_copy.index = data_copy.index.strftime('%Y-%m-%d %H:%M:%S')
            data_copy = data_copy.astype(object).where(pd.notnull(data_copy), None)
            return json.dumps(data_copy.to_dict(), indent=2, default=str)
    elif isinstance(data, (dict, list)):
        def stringify_keys(obj):
            if isinstance(obj, dict):
                return {str(k): stringify_keys(v) for k, v in obj.items()}
            if isinstance(obj, list):
                return [stringify_keys(i) for i in obj]
            if isinstance(obj, float) and obj != obj:
                return None
            return obj
        return json.dumps(stringify_keys(data), indent=2, default=str)
    else:
        try:
            if isinstance(data, float) and data != data:
                data = None
            return json.dumps(data, indent=2, default=str)
        except Exception:
            return json.dumps({"result": str(data)}, indent=2)

# ============================================================================
# Price Data Commands
# ============================================================================

def info(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).info)

def history(ticker: str = DEFAULT_TICKER, period: str = "30d", interval: str = "1d"):
    return format_data_to_json(yf.Ticker(ticker).history(period=period, interval=interval))

def actions(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).actions)

# ============================================================================
# Financial Statements Commands
# ============================================================================

def income(ticker: str = DEFAULT_TICKER, freq: str = "yearly"):
    t = yf.Ticker(ticker)
    if freq == "quarterly":
        return format_data_to_json(t.quarterly_income_stmt)
    elif freq == "trailing":
        return format_data_to_json(t.ttm_income_stmt)
    else:
        return format_data_to_json(t.income_stmt)

def balance(ticker: str = DEFAULT_TICKER, freq: str = "yearly"):
    t = yf.Ticker(ticker)
    if freq == "quarterly":
        return format_data_to_json(t.quarterly_balance_sheet)
    else:
        return format_data_to_json(t.balance_sheet)

def cashflow(ticker: str = DEFAULT_TICKER, freq: str = "yearly"):
    t = yf.Ticker(ticker)
    if freq == "quarterly":
        return format_data_to_json(t.quarterly_cash_flow)
    elif freq == "trailing":
        return format_data_to_json(t.ttm_cash_flow)
    else:
        return format_data_to_json(t.cash_flow)

# ============================================================================
# Holders Commands
# ============================================================================

def major_holders(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).major_holders)

def institutional_holders(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).institutional_holders)

def mutualfund_holders(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).mutualfund_holders)

# ============================================================================
# Insider Commands
# ============================================================================

def insider_transactions(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).insider_transactions)

def insider_purchases(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).insider_purchases)

def insider_roster(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).insider_roster_holders)

# ============================================================================
# Analyst Recommendations Commands
# ============================================================================

def recommendations(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).recommendations)

def recommendations_summary(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).recommendations_summary)

def upgrades(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).upgrades_downgrades)

def earnings_dates(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).earnings_dates)

def earnings_history(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).earnings_history)

def eps_trend(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).eps_trend)

def eps_revisions(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).eps_revisions)

# ============================================================================
# Estimates Commands
# ============================================================================

def earnings_estimates(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).earnings_estimate)

def revenue_estimates(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).revenue_estimate)

def growth_estimates(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).growth_estimates)

def price_targets(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).analyst_price_targets)

# ============================================================================
# News Command
# ============================================================================

def news(ticker: str = DEFAULT_TICKER, days: int | None = None):
    news_data = yf.Ticker(ticker).news
    if days:
        cutoff = datetime.now(timezone.utc) - timedelta(days=days)
        filtered = []
        for item in news_data:
            pub_date_str = item.get('content', {}).get('pubDate')
            if pub_date_str:
                pub_date_str_clean = pub_date_str.rstrip('GMTUTCB ')
                try:
                    pub_date = datetime.strptime(pub_date_str_clean, "%Y-%m-%dT%H:%M:%S").replace(tzinfo=timezone.utc)
                    if pub_date >= cutoff:
                        filtered.append(item)
                except ValueError:
                    filtered.append(item)
            else:
                filtered.append(item)
        return format_data_to_json(filtered)
    else:
        return format_data_to_json(news_data)

# ============================================================================
# Other Commands
# ============================================================================

def calendar(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).calendar)

def sec_filings(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).sec_filings)

def sustainability(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).sustainability)

def isin(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).isin)

def options(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).options)

def metadata(ticker: str = DEFAULT_TICKER):
    return format_data_to_json(yf.Ticker(ticker).history_metadata)


def fetch_data(command: str, ticker: str, **kwargs) -> str:
    """API entrypoint for other scripts to fetch yfinance data as JSON string."""
    func_name = command.replace("-", "_")
    func = globals().get(func_name)
    if func is None:
        raise ValueError(f"Unknown command: {command}")
    return func(ticker=ticker, **kwargs)
