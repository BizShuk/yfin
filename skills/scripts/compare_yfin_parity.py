#!/usr/bin/env python3
"""Compare Python yfinance-oracle and Go yfin batch artifact semantics."""

from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any

from config import COMMANDS


LIST_COMMANDS = {
    "major-holders",
    "institutional-holders",
    "mutualfund-holders",
    "insider-transactions",
    "insider-roster",
    "recommendations",
    "upgrades",
    "earnings-dates",
    "sec-filings",
}

OBJECT_COMMANDS = {
    "insider-purchases",
    "info",
    "calendar",
    "sustainability",
    "metadata",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate 30-command semantic parity between Python and Go artifacts."
    )
    parser.add_argument("--python-root", required=True, type=Path)
    parser.add_argument("--go-root", required=True, type=Path)
    parser.add_argument("--ticker", required=True)
    parser.add_argument("--day", required=True)
    return parser.parse_args()


def artifact_path(root: Path, command: str, ticker: str, day: str) -> Path:
    return root / command / f"{ticker}.{day}.json"


def load_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as artifact:
        return json.load(artifact)


def is_empty(value: Any) -> bool:
    if value is None:
        return True
    if isinstance(value, (dict, list, str)):
        return len(value) == 0
    return False


def main() -> int:
    args = parse_args()
    mismatches: list[str] = []

    for command in COMMANDS:
        python_path = artifact_path(
            args.python_root, command, args.ticker, args.day
        )
        go_path = artifact_path(args.go_root, command, args.ticker, args.day)

        if not python_path.is_file():
            mismatches.append(f"{command}: missing Python artifact {python_path}")
            continue
        if not go_path.is_file():
            mismatches.append(f"{command}: missing Go artifact {go_path}")
            continue

        try:
            python_value = load_json(python_path)
        except (OSError, json.JSONDecodeError) as error:
            mismatches.append(f"{command}: invalid Python JSON: {error}")
            continue
        try:
            go_value = load_json(go_path)
        except (OSError, json.JSONDecodeError) as error:
            mismatches.append(f"{command}: invalid Go JSON: {error}")
            continue

        if is_empty(python_value) != is_empty(go_value):
            mismatches.append(
                f"{command}: empty mismatch "
                f"python={is_empty(python_value)} go={is_empty(go_value)}"
            )

        if command in LIST_COMMANDS and not isinstance(go_value, list):
            mismatches.append(
                f"{command}: Go top-level type must be list, "
                f"got {type(go_value).__name__}"
            )
        if command in OBJECT_COMMANDS and not isinstance(go_value, dict):
            mismatches.append(
                f"{command}: Go top-level type must be object, "
                f"got {type(go_value).__name__}"
            )

    if mismatches:
        print("Yahoo Finance parity mismatches:")
        for mismatch in mismatches:
            print(f"- {mismatch}")
        return 1

    print(
        f"Yahoo Finance parity passed: {args.ticker} {args.day} "
        f"({len(COMMANDS)} commands)"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
