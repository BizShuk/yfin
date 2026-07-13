from pathlib import Path
from datetime import date

RAW_DIR = Path.home() / ".config" / "stock" / "data" / "raw"
FAILED_DIR = RAW_DIR / "_failed"

COMMANDS = [
    "info",
    "history",
    "actions",
    "income",
    "balance",
    "cashflow",
    "major-holders",
    "institutional-holders",
    "mutualfund-holders",
    "insider-transactions",
    "insider-purchases",
    "insider-roster",
    "recommendations",
    "recommendations-summary",
    "upgrades",
    "earnings-dates",
    "earnings-history",
    "eps-trend",
    "eps-revisions",
    "earnings-estimates",
    "revenue-estimates",
    "growth-estimates",
    "price-targets",
    "news",
    "calendar",
    "sec-filings",
    "sustainability",
    "isin",
    "options",
    "metadata",
]

REFRESH_MAP = {
    # daily
    "history": "daily",
    "recommendations": "daily",
    "recommendations-summary": "daily",
    "upgrades": "daily",
    "news": "daily",
    "metadata": "daily",
    # monthly
    "info": "monthly",
    "insider-transactions": "monthly",
    "insider-purchases": "monthly",
    "insider-roster": "monthly",
    "calendar": "monthly",
    # quarterly
    "actions": "quarterly",
    "income": "quarterly",
    "balance": "quarterly",
    "cashflow": "quarterly",
    "major-holders": "quarterly",
    "institutional-holders": "quarterly",
    "mutualfund-holders": "quarterly",
    "earnings-dates": "quarterly",
    "earnings-history": "quarterly",
    "eps-trend": "quarterly",
    "eps-revisions": "quarterly",
    "earnings-estimates": "quarterly",
    "revenue-estimates": "quarterly",
    "growth-estimates": "quarterly",
    "price-targets": "quarterly",
    "sec-filings": "quarterly",
    "sustainability": "quarterly",
    # annually
    "isin": "annually",
    "options": "annually",
}


def get_quarter(month: int) -> int:
    return (month - 1) // 3 + 1


def should_skip(command: str, ticker: str, force: bool, raw_dir: Path = RAW_DIR) -> bool:
    if force:
        return False

    refresh_level = REFRESH_MAP.get(command, "daily")
    command_dir = raw_dir / command

    if not command_dir.exists():
        return False

    # Find file matching {ticker}.*.json
    for f in command_dir.glob(f"{ticker}.*.json"):
        try:
            # Parse date from filename: stem format is {ticker}.{date_str}
            parts = f.stem.split(".")
            if len(parts) < 2:
                continue
            date_str = parts[-1]
            file_date = date.fromisoformat(date_str)
        except ValueError:
            # date parsing failed
            return False

        today = date.today()

        if refresh_level == "daily":
            return file_date == today  # only skip if already fetched today
        elif refresh_level == "weekly":
            age = (today - file_date).days
            return age < 7
        elif refresh_level == "monthly":
            return file_date.year == today.year and file_date.month == today.month
        elif refresh_level == "quarterly":
            file_quarter = get_quarter(file_date.month)
            today_quarter = get_quarter(today.month)
            return file_date.year == today.year and file_quarter == today_quarter
        elif refresh_level == "annually":
            return file_date.year == today.year

    return False


def ensure_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


MA_PERIODS = [5, 10, 20, 60]