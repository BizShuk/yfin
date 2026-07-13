#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import date
from pathlib import Path

import typer

from config import (
    RAW_DIR,
    FAILED_DIR,
    COMMANDS,
    REFRESH_MAP,
    ensure_dir,
    should_skip,
)
import yf

TICKER_LIST_PATH = Path(__file__).parent.parent / "references" / "ticker_list.csv"
DEFAULT_MAX_WORKERS = 10
RETRIES = 3


def get_ticker_list(path: Path) -> list[str]:
    """Read ticker list from CSV, extracting ticker from the last comma-separated field."""
    tickers = []
    with open(path, "r") as f:
        next(f)  # skip header line
        for line in f:
            parts = line.strip().split(",")
            if parts:
                ticker = parts[-1].strip()
                if ticker:
                    tickers.append(ticker)
    return tickers


def fetch_ticker(ticker: str, commands: list[str], force: bool) -> dict:
    """Fetch all commands for a single ticker, with caching and retry logic."""
    results = {"ticker": ticker, "commands": {}}

    for command in commands:
        if should_skip(command, ticker, force):
            results["commands"][command] = "skipped"
            continue

        last_error = None
        for attempt in range(RETRIES):
            try:
                # For special command like history (price), fetch 30 days daily data by default
                if command == "history":
                    json_str = yf.fetch_data(command, ticker, period="30d", interval="1d")
                else:
                    json_str = yf.fetch_data(command, ticker)

                # Save JSON to RAW_DIR/command/{ticker}.{YYYY-MM-DD}.json
                today = date.today().isoformat()
                out_path = RAW_DIR / command / f"{ticker}.{today}.json"
                out_path.parent.mkdir(parents=True, exist_ok=True)
                with open(out_path, "w") as f:
                    f.write(json_str)

                results["commands"][command] = "success"
                break
            except Exception as e:
                last_error = str(e)

                # Check for "not found" / "ticker" error — skip to next ticker
                err_lower = last_error.lower()
                if "not found" in err_lower or "ticker" in err_lower:
                    results["commands"][command] = "not_found"
                    break

                # Retry with exponential backoff if not last attempt
                if attempt < RETRIES - 1:
                    time.sleep(2**attempt)
        else:
            # All retries exhausted — write error to FAILED_DIR
            err_path = FAILED_DIR / f"{ticker}.{command}.err"
            err_path.parent.mkdir(parents=True, exist_ok=True)
            with open(err_path, "w") as f:
                f.write(last_error or "")
            results["commands"][command] = "failed"

    return results


app = typer.Typer()


@app.command()
def main(
    ticker: str | None = typer.Option(None, help="Single ticker to fetch"),
    max_workers: int = typer.Option(DEFAULT_MAX_WORKERS, help="Max concurrent threads"),
    force: bool = typer.Option(False, "--force", help="Force re-fetch, skip cache"),
) -> None:
    ensure_dir(RAW_DIR)
    ensure_dir(FAILED_DIR)

    tickers = [ticker] if ticker else get_ticker_list(TICKER_LIST_PATH)

    print(f"Starting fetch for {len(tickers)} ticker(s), max_workers={max_workers}, force={force}")

    success_count = 0
    skipped_count = 0
    failed_count = 0
    not_found_count = 0

    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = {
            executor.submit(fetch_ticker, t, COMMANDS, force): t for t in tickers
        }
        for future in as_completed(futures):
            result = future.result()
            for cmd, status in result["commands"].items():
                if status == "success":
                    success_count += 1
                elif status == "skipped":
                    skipped_count += 1
                elif status == "failed":
                    failed_count += 1
                elif status == "not_found":
                    not_found_count += 1
            print(f"  {result['ticker']}: {len(result['commands'])} commands processed")

    print(f"Done. success={success_count}, skipped={skipped_count}, failed={failed_count}, not_found={not_found_count}")


if __name__ == "__main__":
    app()
